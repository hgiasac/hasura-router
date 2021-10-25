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
}

// New create Hasura action router
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

func (rt *Router) OnSuccess(callback func(ctx *Context, response interface{}, metadata map[string]interface{})) {
	rt.onSuccess = callback
}

func (rt *Router) OnError(callback func(ctx *Context, err error, metadata map[string]interface{})) {
	rt.onError = callback
}

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
	})

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
		actionError = types.Error{
			Code:    types.ErrCodeUnknown,
			Message: err.Error(),
		}
	}

	responseBytes, err := json.Marshal(map[string]string{
		"code":    actionError.Code,
		"message": actionError.Message,
	})

	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "message": "ERROR: %s" }`, err)))
	}

	w.Write(responseBytes)
}

func validateSessionVariables(variables map[string]string) error {
	role, hasRole := variables[types.XHasuraRole]

	if !hasRole || role == "" {
		return fmt.Errorf("%s header is required", types.XHasuraRole)
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
