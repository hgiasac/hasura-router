package main

import (
	"github.com/hgiasac/hasura-router/go/event"
	"github.com/hgiasac/hasura-router/go/types"
)

func newEventRouter() *event.Router {
	return event.New(map[string]event.Handler{
		"goUserInsert": goUserInsert,
		"goUserUpdate": goUserUpdate,
	}).WithDebug(true)
}

func goUserInsert(ctx *event.Context, payload event.EventTriggerPayload) (interface{}, error) {
	return map[string]interface{}{
		"message": "world!",
	}, nil
}

func goUserUpdate(ctx *event.Context, payload event.EventTriggerPayload) (interface{}, error) {
	return nil, types.NewError("event_failure", "fail event")
}
