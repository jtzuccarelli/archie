package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.health)
	mux.Handle("POST /calls", app.requireAPIKey(http.HandlerFunc(app.createCallsHandler)))
	mux.Handle("GET /calls/{id}", app.requireSession(http.HandlerFunc(app.getCallHandler)))
	mux.Handle("GET /calls", app.requireSession(http.HandlerFunc(app.getCallsHandler)))
	mux.HandleFunc("POST /login", app.loginHandler)
	return mux
}
