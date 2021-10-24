package cron

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hgiasac/hasura-router/go/tracing"
)

type EventPayload struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	ScheduledTime time.Time       `json:"scheduled_time"`
	Payload       json.RawMessage `json:"payload"`
}

type Handler func(ctx *Context, payload EventPayload) (interface{}, error)

type Context struct {
	Headers http.Header
	Tracing *tracing.Tracing
}
