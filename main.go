package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ElectronicsJson struct {
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
}

type LatLon struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func main() {
	http.HandleFunc("/electronics", getJson)

	fmt.Println("**** Starting server at",time.Now().Format("2006-01-02 03:04:05"))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getJson(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	keys, ok := r.URL.Query()["city"]
	if !ok || len(keys[0]) < 1 {
		fmt.Println("Failed to get city")
		notFound(w, r)
		return
	}

	city := keys[0]

	cityProv := city + " nb canada"
	query := fmt.Sprintf("https://nominatim.openstreetmap.org/search/%s?format=json&addressdetails=1&limit=1&polygon_svg=1", cityProv)
	queryResp, err := http.Get(query)
	if err != nil {
		fmt.Println("failed to get lat/lon")
		fmt.Println(query)
		notFound(w, r)
		return
	}
	defer queryResp.Body.Close()

	osData, err := ioutil.ReadAll(queryResp.Body)
	if err != nil {
		notFound(w, r)
		return
	}

	var osLat []LatLon
	err = json.Unmarshal(osData, &osLat)
	if err != nil {
		fmt.Println("failed to parse lat/lon")
		notFound(w, r)
		return
	}

	if len(osLat) == 0 {
		fmt.Println("nothing in json for lat/lon")
		notFound(w, r)
		return
	}

	lat := osLat[0].Lat
	lon := osLat[0].Lon

	theUrl := fmt.Sprintf(`https://www.recyclemyelectronics.ca/nb/wp-admin/admin-ajax.php?action=store_search&lat=%s&lng=%s&max_results=9999&search_radius=25&search=%s&statistics%%5Bcity%%5D=%s&statistics%%5Bregion%%5D=New+Brunswick&statistics%%5Bcountry%%5D=Canada`, lat, lon, url.QueryEscape(city), url.QueryEscape(city))

	//fmt.Println(theUrl)

	resp, err := http.Get(theUrl)
	if err != nil {
		fmt.Println("no results for", theUrl)
		notFound(w, r)
		return
	}

	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		notFound(w, r)
		return
	}

	var result []ElectronicsJson
	// show the HTML code as a string %s
	err = json.Unmarshal([]byte(html), &result)
	if err != nil {
		fmt.Println("failed to unmarshal json from remote")
		notFound(w, r)
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

type jsonData struct {
	OK        bool              `json:"ok"`
	Latitude  string            `json:"lat"`
	Longitude string            `json:"lon"`
	Locations []ElectronicsJson `json:"locations"`
}

func notFound(w http.ResponseWriter, r *http.Request) {
	theData := jsonData{
		OK: false,
	}

	out, _ := json.MarshalIndent(theData, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
}
