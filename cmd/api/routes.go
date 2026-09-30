package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("POST /calls", app.createCallsHandler)
	return mux
}
