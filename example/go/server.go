package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	actions, err := newActionRouter()
	if err != nil {
		panic(err)
	}
	mux.Handle("/actions", actions)
	mux.Handle("/events", newEventRouter())
	mux.Handle("/crons", newCronRouter())

	log.Printf("running server at port 9001")
	if err = http.ListenAndServe("0.0.0.0:9001", mux); err != nil {
		panic(err)
	}
}
