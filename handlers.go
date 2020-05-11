package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type jsonData struct {
	OK        bool              `json:"ok"`
	Latitude  string            `json:"lat"`
	Longitude string            `json:"lon"`
	Locations []ElectronicsJson `json:"locations"`
}

// GetListOfDepots returns json list of depots, or error
func (app *application) GetListOfDepots(w http.ResponseWriter, r *http.Request) {
	app.setupResponse(&w, r)

	keys, ok := r.URL.Query()["city"]
	if !ok || len(keys[0]) < 1 {
		fmt.Println("Failed to get city")
		app.NotFound(w, r)
		return
	}

	var lat, lon string
	foundCity := true
	search := keys[0]

	cityProv := search + " nb canada"
	query := fmt.Sprintf("https://nominatim.openstreetmap.org/search/%s?format=json&addressdetails=1&limit=1&polygon_svg=1", cityProv)
	queryResp, err := http.Get(query)
	if err != nil {
		app.errorLog.Println(err)
		foundCity = false
	}
	defer queryResp.Body.Close()

	osData, err := ioutil.ReadAll(queryResp.Body)
	if err != nil {
		foundCity = false
	}

	var osLat []LatLon

	err = json.Unmarshal(osData, &osLat)
	if err != nil {
		fmt.Println("failed to parse lat/lon")
		app.NotFound(w, r)
		return
	}

	if len(osLat) == 0 {
		fmt.Println("nothing in json for lat/lon")
		foundCity = false
	} else {
		lat = osLat[0].Lat
		lon = osLat[0].Lon
	}

	if foundCity == false {
		// look up by postal code
		latitude, longitude, err := app.GetLatLonForPostalCode(search)
		if err != nil {
			app.errorLog.Println(err)
			app.errorLog.Println("nothing in db for lat/lon when looking up by postal code")
			app.NotFound(w, r)
			return
		}
		lat = latitude
		lon = longitude
		app.infoLog.Println("Lat/lon from postal code of", search, "is", lat, "/", lon)
	}

	theUrl := fmt.Sprintf(`https://www.recyclemyelectronics.ca/nb/wp-admin/admin-ajax.php?action=store_search&lat=%s&lng=%s&max_results=9999&search_radius=25&search=%s&statistics%%5Bcity%%5D=%s&statistics%%5Bregion%%5D=New+Brunswick&statistics%%5Bcountry%%5D=Canada`, lat, lon, url.QueryEscape(search), url.QueryEscape(search))

	resp, err := http.Get(theUrl)
	if err != nil {
		fmt.Println("no results for", theUrl)
		app.NotFound(w, r)
		return
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	var result []ElectronicsJson

	err = json.Unmarshal(html, &result)
	if err != nil {
		fmt.Println("failed to unmarshal json from remote")
		app.NotFound(w, r)
		return
	}

	theData := jsonData{
		OK:        true,
		Latitude:  lat,
		Longitude: lon,
		Locations: result,
	}

	out, _ := json.MarshalIndent(theData, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// NotFound returns json with error
func (app *application) NotFound(w http.ResponseWriter, r *http.Request) {
	theData := jsonData{
		OK: false,
	}

	out, _ := json.MarshalIndent(theData, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// setupResponse allows handles cors
func (app *application) setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
}
