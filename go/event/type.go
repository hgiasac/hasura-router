package event

import (
	"encoding/json"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
)

type EventTriggerPayload struct {
	Event        Event        `json:"event"`
	CreatedAt    string       `json:"created_at"`
	ID           string       `json:"id"`
	TriggerInfo  TriggerInfo  `json:"trigger"`
	TableInfo    TableInfo    `json:"table"`
	DeliveryInfo DeliveryInfo `json:"delivery_info"`
}

type DeliveryInfo struct {
	MaxRetries   int `json:"max_retries"`
	CurrentRetry int `json:"current_retry"`
}

type Event struct {
	SessionVariables map[string]string `json:"session_variables"`
	OP               string            `json:"op"`
	Data             EventData         `json:"data"`
}

type EventData struct {
	Old json.RawMessage `json:"old"`
	New json.RawMessage `json:"new"`
}

type TriggerInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type TableInfo struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type Handler func(ctx *Context, payload EventTriggerPayload) (interface{}, error)

type Context struct {
	Headers http.Header
	Tracing *tracing.Tracing
}
