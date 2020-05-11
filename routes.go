package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/electronics", app.GetElectronics)
	mux.HandleFunc("/oil", app.GetOil)

	return mux
}
