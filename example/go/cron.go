package main

import (
	"github.com/hgiasac/hasura-router/go/cron"
	"github.com/hgiasac/hasura-router/go/types"
)

func newCronRouter() *cron.Router {
	return cron.New(map[string]cron.Handler{
		"goCronSuccess": goCronSuccess,
		"goCronFailure": goCronFailure,
	}).WithDebug(true)
}

func goCronSuccess(ctx *cron.Context, payload cron.EventPayload) (interface{}, error) {
	return map[string]interface{}{
		"message": "success!",
	}, nil
}

func goCronFailure(ctx *cron.Context, payload cron.EventPayload) (interface{}, error) {
	return nil, types.NewError("cron_failure", "fail cron event")
}
