package action

import (
	"encoding/json"
	"net/http"

	"github.com/hgiasac/hasura-router/go/tracing"
)

// ActionName represents an extensible action name in the actions map.
type ActionName string

// action request body
type bodyAction struct {
	Name string `json:"name"`
}

type actionBody struct {
	Action           bodyAction        `json:"action"`
	Input            json.RawMessage   `json:"input"` // This can be serialized into appropriate input type
	SessionVariables map[string]string `json:"session_variables"`
}

// Context represents an extensible action context in the actions map.
type Context struct {
	Headers          http.Header
	SessionVariables map[string]string
	Tracing          *tracing.Tracing
}

// Action represents the action to be executed.
type Action func(ctx *Context, rawBody []byte) (interface{}, error)
