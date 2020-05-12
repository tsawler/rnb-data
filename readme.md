# Recycle NB Data server

This app gets geo data for recycling locations in New
Brunswick.

## Data sources

Postal code lookup ([details](https://www.nrcan.gc.ca/earth-sciences/geography/topographic-information/web-services/geolocation-service/17304)):

`https://geogratis.gc.ca/services/geolocation/en/locate?q=e3g`

## Docker build/run

Build: `docker build -t rnb_data_image .`

Run: `docker container run --name -e FLAGS="runtime flags" rnb rnb_data_image`