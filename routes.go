package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/electronics", app.GetElectronics)
	mux.HandleFunc("/api/electronics", app.GetElectronics)
	mux.HandleFunc("/oil", app.GetOil)
	mux.HandleFunc("/api/oil", app.GetOil)
	mux.HandleFunc("/paint", app.GetPaint)
	mux.HandleFunc("/api/paint", app.GetPaint)
	mux.HandleFunc("/find-location/places", app.Cities)

	return mux
}
