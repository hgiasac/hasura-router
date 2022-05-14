package cron

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hgiasac/hasura-router/go/tracing"
)

// EventPayload represents Hasura schedule cron payload
type EventPayload struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	ScheduledTime time.Time       `json:"scheduled_time"`
	Payload       json.RawMessage `json:"payload"`
}

// Handler represents the event handler to be executed.
type Handler func(ctx *Context, payload EventPayload) (interface{}, error)

// Context represents an extensible event context.
type Context struct {
	context.Context
	Headers http.Header
	Tracing *tracing.Tracing
}
