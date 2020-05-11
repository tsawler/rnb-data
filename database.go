package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GetLatLonForPostalCode looks up geo data for a postal code prefix (3 chars)
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

// GetLatLonForOilDepot gets lat / lon for a depot
func (app *application) GetLatLonForOilDepot(depot, address string) (int, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var lat, lon string
	var id int

	query := `select id, lat, lon from depots where lower(depot_name) = ? and lower(physical_address) = ? limit 1`
	row := app.db.QueryRowContext(ctx, query, strings.ToLower(depot), strings.ToLower(address))

	err := row.Scan(
		&id,
		&lat,
		&lon,
	)

	if err != nil {
		return 0, lat, lon, err
	}

	return id, lat, lon, nil
}

// SaveDepot caches info so we don't have to query the network
func (app *application) SaveDepot(d Depot) (int, error) {
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
		return 0, err
	}

	stmt := "SELECT LAST_INSERT_ID()"
	row := app.db.QueryRowContext(ctx, stmt)

	var id int
	err = row.Scan(&id)

	return id, nil
}

func (app *application) InsertPaintDepot(p PaintData) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	products := ""
	if len(p.Products) > 0 {
		products = strings.Join(p.Products, ", ")
	}

	query := `insert into paint 
		(store, lat, lon, address_line_1, address_line_2, city, province,
		phone, hours, products) 
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := app.db.ExecContext(ctx,
		query,
		p.Depot,
		fmt.Sprintf("%f", p.Location.Lat),
		fmt.Sprintf("%f", p.Location.Lon),
		p.Address.Line1,
		p.Address.Line2,
		p.Address.City,
		p.Address.Province.Value,
		p.Contact.Phone,
		p.Hours,
		products,
	)

	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	stmt := "SELECT LAST_INSERT_ID()"
	row := app.db.QueryRowContext(ctx, stmt)

	var id int
	err = row.Scan(&id)

	return id, nil

}

// GetLatLonForOilDepot gets lat / lon for a depot
func (app *application) GetPaintMerchants() ([]DepotsJson, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stores []DepotsJson

	query := `select 
		id, store, lat, lon, address_line_1, city, province,
		phone, hours, products
	from paint`
	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s DepotsJson
		err = rows.Scan(
			&s.ID,
			&s.Store,
			&s.Lat,
			&s.Lon,
			&s.Address,
			&s.City,
			&s.State,
			&s.Phone,
			&s.Hours,
			&s.Products,
		)
		if err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stores, nil
}
