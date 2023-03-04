package cron

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

// New create an Hasura cron trigger router
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
		"type":         "cron-trigger",
		"http_headers": r.Header,
	})

	eventContext := &Context{
		Headers: r.Header,
		Tracing: tracer,
	}

	var input EventPayload

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, "bad_request", fmt.Errorf("json body could not be decoded: %w", err))
		return
	}

	tracer.SetRequestId(input.ID)
	tracer = tracer.WithFields(map[string]interface{}{
		"event_name":     input.Name,
		"scheduled_time": input.ScheduledTime,
	})

	if rt.debug {
		tracer = tracer.WithField("payload", string(input.Payload))
	}

	resp, err := rt.route(eventContext, input)
	if err != nil {
		rt.onError(eventContext, err, tracer.Values())
		sendError(w, "bad_request", fmt.Errorf("error in executing event: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	rt.onSuccess(eventContext, resp, tracer.Values())
}

func (rt *Router) route(ctx *Context, input EventPayload) ([]byte, error) {
	if rt.handlers == nil {
		return nil, fmt.Errorf("there should be at least one event handler")
	}

	handler, ok := rt.handlers[input.Name]
	if !ok {
		return nil, fmt.Errorf("unknown event %s", input.Name)
	}

	resp, err := handler(ctx, input)
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
		return
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
