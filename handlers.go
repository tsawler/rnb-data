package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ElectronicsJson describes response from json feed
type DepotsJson struct {
	Address      string `json:"address"`
	Description  string `json:"description"`
	Store        string `json:"store"`
	ID           string `json:"id"`
	City         string `json:"city"`
	State        string `json:"state"`
	Zip          string `json:"zip"`
	Lat          string `json:"lat"`
	Lon          string `json:"lng"`
	Terms        string `json:"terms"`
	ResultNumber int    `json:"resultNumber"`
	Phone        string `json:"phone"`
	Hours        string `json:"hours"`
	Products     string `json:"products"`
}

// LatLon holds latitude/longitude
type LatLon struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type jsonData struct {
	OK        bool         `json:"ok"`
	Latitude  string       `json:"lat"`
	Longitude string       `json:"lon"`
	Locations []DepotsJson `json:"locations"`
}

// GetListOfDepots returns json list of depots, or error
func (app *application) GetElectronics(w http.ResponseWriter, r *http.Request) {
	app.setupResponse(&w, r)

	lat, lon, search, err, done := app.getLatLonForCityOrPostalCode(w, r)
	if done {
		return
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

	var result []DepotsJson

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

func (app *application) getLatLonForCityOrPostalCode(w http.ResponseWriter, r *http.Request) (string, string, string, error, bool) {
	keys, ok := r.URL.Query()["city"]
	if !ok || len(keys[0]) < 1 {
		fmt.Println("Failed to get city")
		app.NotFound(w, r)
		return "", "", "", nil, true
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
		return "", "", "", nil, true
	}

	if len(osLat) == 0 {
		fmt.Println("nothing in json for lat/lon")
		foundCity = false
	} else {
		lat = osLat[0].Lat
		lon = osLat[0].Lon
	}

	if foundCity == false {
		// look up by postal code. We only need the first three chars
		postalCodePrefix := search[0:3]
		latitude, longitude, err := app.GetLatLonForPostalCode(postalCodePrefix)
		if err != nil {
			app.errorLog.Println(err)
			app.errorLog.Println("nothing in db for lat/lon when looking up by postal code")
			app.NotFound(w, r)
			return "", "", "", nil, true
		}
		lat = latitude
		lon = longitude
	}
	return lat, lon, search, err, false
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

func (app *application) GetOil(w http.ResponseWriter, r *http.Request) {
	app.setupResponse(&w, r)

	lat, lon, search, err, done := app.getLatLonForCityOrPostalCode(w, r)
	if done {
		return
	}

	res, err := http.Get(fmt.Sprintf(`https://nb.uoma-atlantic.com/en/collection-facilities?location=%s`, url.QueryEscape(search)))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result []DepotsJson

	// Find the review items
	doc.Find("#collection_facility-list-results li").Each(func(i int, s *goquery.Selection) {
		address := s.Find("a").Text()
		numberedDepot := s.Find("b").Text()
		exploded := strings.Split(numberedDepot, " ")
		depot := ""
		for n := 1; n < len(exploded); n++ {
			depot = fmt.Sprintf("%s %s", depot, exploded[n])
		}
		hours := s.Find("small").Text()
		var items []string
		s.Find("img").Each(func(j int, t *goquery.Selection) {
			title := t.AttrOr("title", "")
			if title != "" {
				items = append(items, title)
			}
		})

		var products string
		if len(items) > 0 {
			products = strings.Join(items, ", ")
		}

		j := DepotsJson{
			Store:    strings.TrimSpace(depot),
			Address:  strings.TrimSpace(address),
			Hours:    strings.TrimSpace(hours),
			Products: products,
		}
		result = append(result, j)
	})

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
