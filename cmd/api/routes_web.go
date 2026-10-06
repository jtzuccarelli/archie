package main

import "net/http"

func (app *application) registerWebRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /calls", app.getCallsHandler)
	mux.HandleFunc("GET /calls/{id}", app.getCallByIDHandler)
}

func (app *application) getCallsHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (app *application) getCallByIDHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
