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

// DepotsJson describes response from json feed
type DepotsJson struct {
	Address      string `json:"address"`
	Description  string `json:"description"`
	Store        string `json:"store"`
	ID           string `json:"id"`
	MyID         int    `json:"myID"`
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

// getLatLonForCityOrPostalCode returns lat/lon for a city or a postal code
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
		// Look up by postal code. We only need the first three characters.
		postalCodePrefix := search[0:3]
		latitude, longitude, err := app.GetLatLonForPostalCode(postalCodePrefix)
		if err != nil {
			app.errorLog.Println(err)
			app.errorLog.Println("nothing in db for lat/lon when looking up by postal code")
			app.NotFound(w, r)
			return "", "", "", nil, true
		}
		app.infoLog.Println("using lat/lon of", latitude, longitude)
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

// GetOil gets oil recycling locations
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

	// Find the items
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

		qString := strings.ReplaceAll(strings.TrimSpace(address), ", ", " ")
		explodedAddress := strings.Split(qString, " ")
		queryAddress := ""
		for k := 0; k < len(explodedAddress)-2; k++ {
			queryAddress = fmt.Sprintf("%s %s", queryAddress, explodedAddress[k])
		}

		// See if we have the lat/lon for this one
		var latitude, longitude string
		var id int
		id, latitude, longitude, err := app.GetLatLonForOilDepot(strings.TrimSpace(depot), strings.TrimSpace(address))

		if err != nil {
			// We don't have it. Look it up.
			query := fmt.Sprintf("https://nominatim.openstreetmap.org/search/%s?format=json&addressdetails=1&limit=1&polygon_svg=1", strings.TrimSpace(queryAddress))

			queryResp, err := http.Get(query)
			if err != nil {
				app.errorLog.Println(err)
			}
			defer queryResp.Body.Close()

			osData, err := ioutil.ReadAll(queryResp.Body)
			if err != nil {
				app.errorLog.Println("Can't find lat/lon for oil")
			}

			var osLat []LatLon

			err = json.Unmarshal(osData, &osLat)
			if err != nil {
				app.errorLog.Println("failed to parse lat/lon")
			} else {
				if len(osLat) > 0 {
					latitude = osLat[0].Lat
					longitude = osLat[0].Lon
				}
			}

			// save it
			d := Depot{
				DepotName: strings.TrimSpace(depot),
				Address:   strings.TrimSpace(address),
				Hours:     strings.TrimSpace(hours),
				Products:  products,
				Lon:       longitude,
				Lat:       latitude,
			}
			id, err = app.SaveDepot(d)
			if err != nil {
				app.errorLog.Println(err)
			}
		}

		j := DepotsJson{
			MyID:         id,
			Store:        strings.TrimSpace(depot),
			Address:      strings.TrimSpace(address),
			Hours:        strings.TrimSpace(hours),
			Products:     products,
			Lon:          longitude,
			Lat:          latitude,
			ResultNumber: i + 1,
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

type PaintProvince struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type PaintAddress struct {
	Line1    string        `json:"address_line_1"`
	Line2    string        `json:"address_line_2"`
	City     string        `json:"city"`
	Province PaintProvince `json:"province"`
}

type PaintContact struct {
	Phone string `json:"phone"`
}

type PaintLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lng"`
}

type PaintData struct {
	ID       int           `json:"myID"`
	Address  PaintAddress  `json:"address"`
	Products []string      `json:"accepted_products"`
	Contact  PaintContact  `json:"contact"`
	Location PaintLocation `json:"geolocation"`
	Hours    string        `json:"hours"`
	Depot    string        `json:"title"`
}

func (app *application) GetPaint(w http.ResponseWriter, r *http.Request) {
	app.setupResponse(&w, r)

	actions, ok := r.URL.Query()["action"]
	if !ok || len(actions[0]) < 1 {
		app.NotFound(w, r)
		return
	}

	lat, lon, _, _, done := app.getLatLonForCityOrPostalCode(w, r)
	if done {
		return
	}

	action := actions[0]
	var depots []DepotsJson

	if action == "all" {
		d, err := app.GetPaintMerchants()
		if err != nil {
			app.NotFound(w, r)
			return
		}
		depots = d
	} else {
		d, err := app.GetPaintMerchantsForLatLon(lat, lon)
		if err != nil {
			app.NotFound(w, r)
			return
		}
		depots = d
	}

	theData := jsonData{
		OK:        true,
		Latitude:  lat,
		Longitude: lon,
		Locations: depots,
	}

	out, _ := json.MarshalIndent(theData, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)

}
