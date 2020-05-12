# Recycle NB Data server

This app gets geo data for recycling locations in New
Brunswick.


## Data sources

Postal code lookup:

`https://geogratis.gc.ca/services/geolocation/en/locate?q=e3g`

## Docker build/run

Build: `docker build -t rnb_data_image .`

Run: `docker container run --name rnb rnb_data_image $FLAGS`