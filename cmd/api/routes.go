package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.health)
	mux.Handle("POST /calls", app.requireAPIKey(http.HandlerFunc(app.createCallsHandler)))
	mux.HandleFunc("GET /calls/{id}", app.getCallHandler)
	mux.HandleFunc("GET /calls", app.getCallsHandler)
	mux.HandleFunc("POST /login", app.loginHandler)
	return mux
}
