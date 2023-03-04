package main

import (
	"encoding/json"

	"github.com/hgiasac/hasura-router/go/action"
	"github.com/hgiasac/hasura-router/go/types"
)

func newActionRouter() (*action.Router, error) {
	actions, err := action.New(map[action.ActionName]action.Action{
		"goHello":   goActionHello,
		"goFailure": goActionFailure,
	})

	if err != nil {
		return nil, err
	}
	return actions.WithDebug(true), nil
}

func goActionHello(ctx *action.Context, rawBody []byte) (interface{}, error) {
	var input struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rawBody, &input); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": input.Message,
	}, nil
}

func goActionFailure(ctx *action.Context, rawBody []byte) (interface{}, error) {
	return nil, types.NewError("action_failure", "fail action")
}
