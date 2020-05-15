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
	ID   int    `json:"id"`
	City string `json:"text"`
}

type Cities struct {
	Cities []City `json:"results"`
}
