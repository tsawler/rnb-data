package main

import (
	"context"
	"strings"
	"time"
)

func (app *application) GetLatLonForPostalCode(pc string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var lat, lon string

	query := `select lat, lon from postal where lower(prefix) = ?`
	row := app.db.QueryRowContext(ctx, query, strings.ToLower(pc))

	err := row.Scan(
		&lat,
		&lon,
	)

	if err != nil {
		return lat, lon, err
	}

	return lat, lon, nil
}

func (app *application) GetLatLonForOilDepot(depot, address string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var lat, lon string

	query := `select lat, lon from depots where lower(depot_name) = ? and lower(physical_address) = ? limit 1`
	row := app.db.QueryRowContext(ctx, query, strings.ToLower(depot), strings.ToLower(address))

	err := row.Scan(
		&lat,
		&lon,
	)

	if err != nil {
		return lat, lon, err
	}

	return lat, lon, nil
}

func (app *application) SaveDepot(d Depot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into depots (lat, lon, depot_name, physical_address, hours, products, terms, description) 
		values (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := app.db.ExecContext(ctx,
		query,
		d.Lat,
		d.Lon,
		d.DepotName,
		d.Address,
		d.Hours,
		d.Products,
		d.Terms,
		d.Description,
	)

	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}
