package event

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
)

// OpName represents the name of the operation
type OpName string

const (
	OpInsert OpName = "INSERT"
	OpUpdate OpName = "UPDATE"
	OpDelete OpName = "DELETE"
	OpManual OpName = "MANUAL"
)

// EventTriggerPayload represents Hasura event trigger payload
// https://hasura.io/docs/latest/graphql/core/event-triggers/payload/
type EventTriggerPayload struct {
	Event        Event        `json:"event"`
	CreatedAt    string       `json:"created_at"`
	ID           string       `json:"id"`
	Trigger      TriggerInfo  `json:"trigger"`
	Table        EventTable   `json:"table"`
	DeliveryInfo DeliveryInfo `json:"delivery_info"`
}

type DeliveryInfo struct {
	MaxRetries   int `json:"max_retries"`
	CurrentRetry int `json:"current_retry"`
}

type Event struct {
	SessionVariables map[string]string `json:"session_variables"`
	OP               OpName            `json:"op"`
	Data             EventData         `json:"data"`
}

type EventData struct {
	Old json.RawMessage `json:"old"`
	New json.RawMessage `json:"new"`
}

type TriggerInfo struct {
	Name string `json:"name"`
}

type EventTable struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

// Handler represents the event handler to be executed.
type Handler func(ctx *Context, payload EventTriggerPayload) (interface{}, error)

// Context represents an extensible event context.
type Context struct {
	context.Context
	Headers http.Header
	Tracing *tracing.Tracing
}
