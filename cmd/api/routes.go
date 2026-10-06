package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	app.registerAPIRoutes(mux)
	app.registerWebRoutes(mux)
	return mux
}
