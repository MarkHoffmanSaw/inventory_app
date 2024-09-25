package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Customer struct {
	id           int    `field:"customer_id"`
	name         string `field:"name"`
	customerType string `field:"customer_type"`
	code         string `field:"customer_code"`
}

type Location struct {
	id          int    `field:"location_id"`
	name        string `field:"name"`
	warehouseID int    `field:"warehouse_id"`
}

func fetchCustomers(db *sql.DB) ([]Customer, error) {
	rows, err := db.Query("SELECT * FROM customers;")
	if err != nil {
		log.Println("Error fetchCustomers1: ", err)
		return nil, err
	}
	defer rows.Close()

	var customers []Customer

	for rows.Next() {
		var customer Customer
		if err := rows.Scan(&customer.id, &customer.name, &customer.customerType, &customer.code); err != nil {
			log.Println("Error fetchCustomers2: ", err)
			return customers, err
		}
		customers = append(customers, customer)
	}
	if err = rows.Err(); err != nil {
		return customers, err
	}

	return customers, nil
}

func fetchLocations(db *sql.DB) ([]Location, error) {
	rows, err := db.Query("SELECT * FROM locations;")
	if err != nil {
		log.Println("Error fetchLocations1: ", err)
		return nil, err
	}
	defer rows.Close()

	var locations []Location

	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.id, &location.name, &location.warehouseID); err != nil {
			log.Println("Error fetchLocations2: ", err)
			return locations, err
		}
		locations = append(locations, location)
	}
	if err = rows.Err(); err != nil {
		return locations, err
	}

	return locations, nil
}

func addMaterial(myWindow fyne.Window, db *sql.DB) {
	customers, _ := fetchCustomers(db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	locations, _ := fetchLocations(db)
	var locationsStr []string
	locationsMap := make(map[string]int)
	for _, location := range locations {
		locationsStr = append(locationsStr, location.name)
		locationsMap[location.name] = location.id
	}

	types := []string{"Envelope", "Card", "Carrier", "Insert", "Consumables"}

	customerSelector := widget.NewSelect(customersStr, func(s string) {})
	locationSelector := widget.NewSelect(locationsStr, func(s string) {})
	typeSelector := widget.NewSelect(types, func(s string) {})
	stockIDInput := widget.NewEntry()
	descrInput := widget.NewEntry()
	quantityInput := widget.NewEntry()
	notesInput := widget.NewEntry()

	dialog := dialog.NewForm("Add Material", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Customer", customerSelector),
			widget.NewFormItem("Location", locationSelector),
			widget.NewFormItem("Stock ID", stockIDInput),
			widget.NewFormItem("Description", descrInput),
			widget.NewFormItem("Type", typeSelector),
			widget.NewFormItem("Quantity", quantityInput),
			widget.NewFormItem("Notes", notesInput),
		}, func(confirm bool) {
			if confirm {
				quantity, _ := strconv.Atoi(quantityInput.Text)

				_, err := db.Exec(`INSERT INTO materials
				(stock_id, location_id, customer_id, material_type,description, notes, quantity, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					stockIDInput.Text, locationsMap[locationSelector.Selected], customersMap[customerSelector.Selected],
					typeSelector.Selected, descrInput.Text, notesInput.Text, quantity, time.Now())

				if err != nil {
					log.Println("Error saving material:", err)
					dialog.ShowInformation("Error", "Unable to save data: "+err.Error(), myWindow)
				} else {
					dialog.ShowInformation("Success", "Material saved", myWindow)
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(400, 400))

	dialog.Show()

}
