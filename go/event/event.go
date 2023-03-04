package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
	"github.com/hgiasac/hasura-router/go/types"
)

// Router represent a generic event trigger http handler
type Router struct {
	handlers  map[string]Handler
	onSuccess func(ctx *Context, response interface{}, metadata map[string]interface{})
	onError   func(ctx *Context, err error, metadata map[string]interface{})
	debug     bool
}

// New create an Hasura event trigger router
func New(handlers map[string]Handler) *Router {
	return &Router{
		handlers:  handlers,
		onSuccess: onSuccess,
		onError:   onError,
	}
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
		"event_name":        payload.Trigger.Name,
		"op":                payload.Event.OP,
		"session_variables": payload.Event.SessionVariables,
		"table_schema":      payload.Table.Schema,
		"table_name":        payload.Table.Name,
		"created_at":        payload.CreatedAt,
		"max_retries":       payload.DeliveryInfo.MaxRetries,
		"current_retry":     payload.DeliveryInfo.CurrentRetry,
	})

	if rt.debug {
		tracer = tracer.WithFields(map[string]interface{}{
			"data_old": string(payload.Event.Data.Old),
			"data_new": string(payload.Event.Data.New),
		})
	}

	if err := validateSessionVariables(payload.Event.SessionVariables); err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, types.ErrCodeBadRequest, err)
		return
	}

	resp, err := rt.route(eventContext, payload)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, types.ErrCodeBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	rt.onSuccess(eventContext, resp, tracer.Values())
}

func (rt *Router) route(ctx *Context, payload EventTriggerPayload) ([]byte, error) {
	if rt.handlers == nil {
		return nil, fmt.Errorf("there should be at least one event handler")
	}

	handler, ok := rt.handlers[payload.Trigger.Name]
	if !ok {
		return nil, fmt.Errorf("unknown event %s", payload.Trigger.Name)
	}

	resp, err := handler(ctx, payload)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

func sendError(w http.ResponseWriter, code string, err error) {
	w.WriteHeader(400)

	responseBytes, err := json.Marshal(map[string]interface{}{
		"code":    code,
		"message": err.Error(),
		"error":   err,
	})

	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "message": "ERROR: %s" }`, err)))
		return
	}

	w.Write(responseBytes)
}

func onSuccess(ctx *Context, response interface{}, metadata map[string]interface{}) {
	metadata["level"] = "info"
	metadata["message"] = "executed action successfully"

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		log.Println(metadata)
		return
	}

	log.Println(string(jsonBytes))
}

func onError(ctx *Context, err error, metadata map[string]interface{}) {
	metadata["level"] = "error"
	metadata["error"] = err
	metadata["message"] = err.Error()

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		log.Println(metadata)
		return
	}

	log.Println(string(jsonBytes))
}

func validateSessionVariables(variables map[string]string) error {
	role, hasRole := variables[types.XHasuraRole]

	if !hasRole || role == "" {
		return fmt.Errorf("%s session variable is required", types.XHasuraRole)
	}
	return nil
}
