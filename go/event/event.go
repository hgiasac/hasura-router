package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
	"github.com/hgiasac/hasura-router/go/types"
)

type Router struct {
	handlers  map[string]Handler
	onSuccess func(ctx *Context, response interface{}, metadata map[string]interface{})
	onError   func(ctx *Context, err error, metadata map[string]interface{})
}

func New(handlers map[string]Handler) *Router {
	return &Router{
		handlers:  handlers,
		onSuccess: onSuccess,
		onError:   onError,
	}
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
		"type":         "event-trigger",
		"http_headers": r.Header,
	})

	eventContext := &Context{
		Headers: r.Header,
		Tracing: tracer,
	}
	var payload EventTriggerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, types.ErrCodeBadRequest, fmt.Errorf("json body could not be decoded: %w", err))
		return
	}

	tracer.SetRequestId(payload.ID)
	tracer.WithFields(map[string]interface{}{
		"event_name":        payload.TriggerInfo.Name,
		"op":                payload.Event.OP,
		"session_variables": payload.Event.SessionVariables,
	})

	if err := validateSessionVariables(payload.Event.SessionVariables); err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, types.ErrCodeBadRequest, err)
	}

	resp, err := rt.route(eventContext, payload)
	if err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, types.ErrCodeBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	rt.onSuccess(eventContext, resp, tracer.Values())
}

func (rt *Router) route(ctx *Context, payload EventTriggerPayload) ([]byte, error) {
	if rt.handlers == nil {
		return nil, fmt.Errorf("there should be at least one event handler")
	}

	handler, ok := rt.handlers[payload.TriggerInfo.Name]
	if !ok {
		return nil, fmt.Errorf("unknown event %s", payload.TriggerInfo.Name)
	}

	resp, err := handler(ctx, payload)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

func sendError(w http.ResponseWriter, code string, err error) {
	w.WriteHeader(400)

	responseBytes, err := json.Marshal(map[string]string{
		"code":    code,
		"message": err.Error(),
	})

	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "message": "ERROR: %s" }`, err)))
	}

	w.Write(responseBytes)
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

func validateSessionVariables(variables map[string]string) error {
	role, hasRole := variables[types.XHasuraRole]

	if !hasRole || role == "" {
		return fmt.Errorf("%s header is required", types.XHasuraRole)
	}
	return nil
}
