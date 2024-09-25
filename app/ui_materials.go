package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Location struct {
	id          int    `field:"location_id"`
	name        string `field:"name"`
	warehouseID int    `field:"warehouse_id"`
}

type Customer struct {
	id           int    `field:"customer_id"`
	name         string `field:"name"`
	customerType string `field:"customer_type"`
	code         string `field:"customer_code"`
}

// materialId, stockIDEntrySelect.Text, quantity, notes, time.Now()
type Transaction struct {
	materialId int       `field:"material_id"`
	stockId    string    `field:"stock_id"`
	quantity   int       `field:"quantity_change"`
	notes      string    `field:"notes"`
	updateAt   time.Time `field:"updated_at"`
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

func fetchMaterialsByCustomer(db *sql.DB, customerId int) ([]Material, error) {
	rows, err := db.Query(`SELECT m.material_id, m.stock_id, l.name, m.description,
							m.notes, m.quantity, m.updated_at, c.name, m.material_type
							FROM materials m
							LEFT JOIN locations l ON m.location_id = l.location_id
							LEFT JOIN customers c ON c.customer_id = m.customer_id
							WHERE m.customer_id = $1`, customerId)
	if err != nil {
		log.Println("Error fetchMaterialsByCustomer1: ", err)
		return nil, err
	}
	defer rows.Close()

	var materials []Material

	for rows.Next() {
		var material Material
		if err := rows.Scan(
			&material.MaterialID,
			&material.StockID,
			&material.LocationName,
			&material.Description,
			&material.Notes,
			&material.Quantity,
			&material.UpdatedAt,
			&material.CustomerName,
			&material.MaterialType,
		); err != nil {
			log.Println("Error fetchMaterialsByCustomer2: ", err)
			return materials, err
		}
		materials = append(materials, material)
	}
	if err = rows.Err(); err != nil {
		return materials, err
	}

	return materials, nil
}

func addTranscation(trx *Transaction, db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO transactions_log (material_id, stock_id, quantity_change, notes, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 `, trx.materialId, trx.stockId, trx.quantity, trx.notes, trx.updateAt,
	)

	if err != nil {
		return err
	} else {
		return nil
	}
}

// ////////////////////////////////////////////////////
// ACTIONS
// ///////////////////////////////////////////////////
func createMaterial(myWindow fyne.Window, db *sql.DB) {
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

	dialog := dialog.NewForm("Create Material", "Save", "Cancel",
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
				(stock_id, location_id, customer_id, material_type, description, notes, quantity, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, // INSERT INTO transactions
					stockIDInput.Text, locationsMap[locationSelector.Selected], customersMap[customerSelector.Selected],
					typeSelector.Selected, descrInput.Text, notesInput.Text, quantity, time.Now())

				if err != nil {
					log.Println("Error saving material:", err)
					dialog.ShowInformation("Error", "Unable to save data: "+err.Error(), myWindow)
				} else {
					dialog.ShowInformation("Success", "Material created", myWindow)
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(400, 400))
	dialog.Show()
}

// Add an existing material to a location
func addMaterial(myWindow fyne.Window, db *sql.DB) {
	customers, _ := fetchCustomers(db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	var materials []Material
	var materialsStr []string
	materialsMap := make(map[string]int)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			materialsStr = append(materialsStr, material.StockID+"|"+material.LocationName)
			materialsMap[material.StockID] = material.MaterialID
		}
	})

	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "", customerSelector,
		func(confirm bool) {
			if confirm {
				stockIDEntrySelect := widget.NewSelectEntry(materialsStr)
				quantityInput := widget.NewEntry()
				notesInput := widget.NewEntry()

				dialogMaterial := dialog.NewForm("Add Material for "+customerSelector.Selected, "Save", "Cancel",
					[]*widget.FormItem{
						widget.NewFormItem("Stock ID", stockIDEntrySelect),
						widget.NewFormItem("Quantity", quantityInput),
						widget.NewFormItem("Notes", notesInput),
					}, func(confirm bool) {
						if confirm {
							quantity, _ := strconv.Atoi(quantityInput.Text)
							materialId := materialsMap[strings.Split(stockIDEntrySelect.Text, "|")[0]]
							notes := notesInput.Text

							_, err := db.Exec(`
							UPDATE materials
							SET quantity = (quantity + $1),
								notes = $2
							WHERE material_id = $3;
							`, quantity, notes, materialId,
							)

							if err != nil {
								dialog.ShowInformation("Error", "Updating material error: "+err.Error(), myWindow)
							} else {
								err := addTranscation(&Transaction{
									materialId: materialId,
									stockId:    stockIDEntrySelect.Text,
									quantity:   quantity,
									notes:      notes,
									updateAt:   time.Now(),
								}, db)

								if err != nil {
									dialog.ShowInformation("Error", "Updating transactions error: "+err.Error(), myWindow)
								} else {
									dialog.ShowInformation("Success", "Material updated", myWindow)
								}
							}
						}
					}, myWindow)

				dialogMaterial.Resize(fyne.NewSize(400, 400))
				dialogMaterial.Show()
			}
		}, myWindow)

	dialogCustomer.Resize(fyne.NewSize(100, 100))
	dialogCustomer.Show()

}
