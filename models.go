package main

// Depot holds information about a recycling location
type Depot struct {
	ID          int
	DepotName   string
	Lat         string
	Lon         string
	Address     string
	Hours       string
	Products    string
	Terms       string
	Description string
}

type City struct {
	LatLon string `json:"value"`
	City   string `json:"text"`
}
