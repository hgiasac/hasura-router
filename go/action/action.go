package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
	"github.com/hgiasac/hasura-router/go/types"
)

// Router represent a generic action http handler
type Router struct {
	actions   map[ActionName]Action
	onSuccess func(ctx *Context, response interface{}, metadata map[string]interface{})
	onError   func(ctx *Context, err error, metadata map[string]interface{})
	debug     bool
}

// New create an Hasura action router
func New(actions map[ActionName]Action) (*Router, error) {
	if len(actions) == 0 {
		return nil, errors.New("there should be at least one action")
	}

	return &Router{
		actions:   actions,
		onSuccess: onSuccess,
		onError:   onError,
	}, nil
}

// WithDebug set debug mode to add input data to the tracing context
func (rt *Router) WithDebug(debug bool) *Router {
	rt.debug = debug
	return rt
}

// OnSuccess set a function to handle success callback
func (rt *Router) OnSuccess(callback func(ctx *Context, response interface{}, metadata map[string]interface{})) {
	rt.onSuccess = callback
}

// OnError set a function to handle error callback
func (rt *Router) OnError(callback func(ctx *Context, err error, metadata map[string]interface{})) {
	rt.onError = callback
}

// ServeHTTP implements the serving http interface
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestId := r.Header.Get(types.XRequestId)
	tracer := tracing.New(requestId).WithFields(map[string]interface{}{
		"type":         "action",
		"http_headers": r.Header,
	})
	actionContext := &Context{
		Headers: r.Header,
		Tracing: tracer,
	}

	var payload actionBody
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		rt.onError(actionContext, err, tracer.Values())
		sendError(w, types.NewError(types.ErrCodeBadRequest, fmt.Sprintf("json body could not be decoded: %s", err.Error())))
		return
	}

	tracer.WithFields(map[string]interface{}{
		"action":            payload.Action.Name,
		"session_variables": payload.SessionVariables,
		"request_query":     payload.RequestQuery,
	})

	if rt.debug {
		tracer = tracer.WithField("input", string(payload.Input))
	}

	actionContext.RequestQuery = payload.RequestQuery

	if err := validateSessionVariables(payload.SessionVariables); err != nil {
		rt.onError(actionContext, err, tracer.Values())
		sendError(w, types.NewError(types.ErrCodeBadRequest, err.Error()))
	}

	actionContext.SessionVariables = payload.SessionVariables
	response, err := rt.route(actionContext, payload)

	if err != nil {
		rt.onError(actionContext, err, tracer.Values())
		sendError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	rt.onSuccess(actionContext, response, tracer.Values())
}

func (rt *Router) route(ctx *Context, payload actionBody) ([]byte, error) {

	execute, ok := rt.actions[ActionName(payload.Action.Name)]
	if !ok {
		return nil, types.NewError(types.ErrCodeNotFound, fmt.Sprintf("unknown action %s", payload.Action.Name))
	}

	resp, err := execute(ctx, payload.Input)
	if err != nil {
		return nil, err
	}

	bytes, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		return nil, types.NewError(types.ErrCodeInternal, jsonErr.Error())
	}

	return bytes, nil
}

func sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(400)

	var actionError types.Error

	if ok := errors.As(err, &actionError); !ok {
		actionError = types.NewError(types.ErrCodeUnknown, err.Error())
	}

	responseBytes, err := json.Marshal(actionError)

	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "message": "ERROR: %s" }`, err)))
		return
	}

	w.Write(responseBytes)
}

func validateSessionVariables(variables map[string]string) error {
	role, hasRole := variables[types.XHasuraRole]

	if !hasRole || role == "" {
		return fmt.Errorf("%s session variable is required", types.XHasuraRole)
	}
	return nil
}

func onSuccess(ctx *Context, response interface{}, metadata map[string]interface{}) {
	metadata["level"] = "info"
	metadata["message"] = "executed action successfully"

	jsonStr, err := json.Marshal(metadata)
	if err != nil {
		log.Println(metadata)
	}

	log.Println(jsonStr)
}

func onError(ctx *Context, err error, metadata map[string]interface{}) {
	metadata["level"] = "error"
	metadata["error"] = err
	metadata["message"] = err.Error()

	jsonStr, err := json.Marshal(metadata)
	if err != nil {
		log.Println(metadata)
	}

	log.Println(string(jsonStr))
}
