package main

import "net/http"

func (app *application) registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", app.health)
	mux.Handle("POST /calls", app.requireQVSecret(http.HandlerFunc(app.createCallsHandler)))
}
