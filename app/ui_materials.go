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

type MaterialInfo struct {
	materialId   int       `field:"material_id"`
	stockId      string    `field:"stock_id"`
	locationId   int       `field:"location_id"`
	customerId   int       `field:"customer_id"`
	materialType string    `field:"material_type"`
	description  string    `field:"description"`
	notes        string    `field:"notes"`
	quantity     int       `field:"quantity"`
	updated_at   time.Time `field:"updated_at"`
}

type TransactionInfo struct {
	materialId int       `field:"material_id"`
	stockId    string    `field:"stock_id"`
	quantity   int       `field:"quantity_change"`
	notes      string    `field:"notes"`
	cost       int       `field:"cost"`
	updatedAt  time.Time `field:"updated_at"`
	jobTicket  string    `field:"job_ticket"`
}

//////////////////////////////////////////
// FETCH DATA FROM THE DB
//////////////////////////////////////////

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

func addTranscation(trx *TransactionInfo, db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO transactions_log (material_id, stock_id, quantity_change, notes, cost, job_ticket, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 `, trx.materialId, trx.stockId, trx.quantity, trx.notes, trx.cost, trx.jobTicket, trx.updatedAt,
	)

	if err != nil {
		return err
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////
// ACTIONS
// Update the current materials quantity within locations
//////////////////////////////////////////////////////////

// Create a new material in a location
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
				var material Material
				quantity, _ := strconv.Atoi(quantityInput.Text)

				err := db.QueryRow(`INSERT INTO materials
				(stock_id, location_id, customer_id, material_type, description, notes, quantity, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING material_id;`,
					stockIDInput.Text, locationsMap[locationSelector.Selected], customersMap[customerSelector.Selected],
					typeSelector.Selected, descrInput.Text, notesInput.Text, quantity, time.Now()).Scan(&material.MaterialID)

				if err != nil {
					log.Println("Error saving material:", err)
					dialog.ShowInformation("Error", "Unable to save data: "+err.Error(), myWindow)
				} else {
					err := addTranscation(&TransactionInfo{
						materialId: material.MaterialID,
						stockId:    stockIDInput.Text,
						quantity:   quantity,
						notes:      notesInput.Text,
						updatedAt:  time.Now(),
					}, db)

					if err != nil {
						dialog.ShowInformation("Error", "Updating transactions error: "+err.Error(), myWindow)
					} else {
						dialog.ShowInformation("Success", "Material created", myWindow)
					}
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(600, 400))
	dialog.Show()
}

// Add a new material to a location
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
	materialsMap := make(map[[2]string]int)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			materialsStr = append(materialsStr, material.StockID+"|"+material.LocationName)
			materialsMap[[2]string{material.StockID, material.LocationName}] = material.MaterialID
		}
	})

	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "", customerSelector,
		func(confirm bool) {
			if confirm && customerSelector.Selected != "" {
				stockIDEntrySelect := widget.NewSelectEntry(materialsStr)
				quantityInput := widget.NewEntry()
				notesInput := widget.NewEntry()

				dialogMaterial := dialog.NewForm("Replenish Material", "Add", "Cancel",
					[]*widget.FormItem{
						widget.NewFormItem("Stock ID", stockIDEntrySelect),
						widget.NewFormItem("Add quantity", quantityInput),
						widget.NewFormItem("Notes", notesInput),
					},
					func(confirm bool) {
						if confirm {
							stockId := strings.Split(stockIDEntrySelect.Text, "|")[0]
							locationName := strings.Split(stockIDEntrySelect.Text, "|")[1]
							quantity, _ := strconv.Atoi(quantityInput.Text)
							materialId := materialsMap[[2]string{stockId, locationName}]
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
								err := addTranscation(&TransactionInfo{
									materialId: materialId,
									stockId:    stockId,
									quantity:   quantity,
									notes:      notes,
									updatedAt:  time.Now(),
								}, db)

								if err != nil {
									dialog.ShowInformation("Error", "Updating transactions error: "+err.Error(), myWindow)
								} else {
									dialog.ShowInformation("Success", "Material updated", myWindow)
								}
							}
						}
					}, myWindow)

				dialogMaterial.Resize(fyne.NewSize(600, 300))
				dialogMaterial.Show()
			}
		}, myWindow)

	dialogCustomer.Resize(fyne.NewSize(300, 100))
	dialogCustomer.Show()
}

// Remove a material from a location
func removeMaterial(myWindow fyne.Window, db *sql.DB) {
	customers, _ := fetchCustomers(db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	var materials []Material
	var materialsStr []string
	materialsMap := make(map[[2]string]int)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			materialsStr = append(materialsStr, material.StockID+"|"+material.LocationName)
			materialsMap[[2]string{material.StockID, material.LocationName}] = material.MaterialID
		}
	})

	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "", customerSelector,
		func(confirm bool) {
			if confirm && customerSelector.Selected != "" {
				stockIDEntrySelect := widget.NewSelectEntry(materialsStr)
				quantityInput := widget.NewEntry()
				notesInput := widget.NewEntry()
				jobTicketInput := widget.NewEntry()

				dialogMaterial := dialog.NewForm("Remove material", "Remove", "Cancel",
					[]*widget.FormItem{
						widget.NewFormItem("Stock ID", stockIDEntrySelect),
						widget.NewFormItem("Remove Quantity", quantityInput),
						widget.NewFormItem("Notes", notesInput),
						widget.NewFormItem("Job Ticket #", jobTicketInput),
					},
					func(confirm bool) {
						if confirm {
							quantity, _ := strconv.Atoi(quantityInput.Text)
							stockId := strings.Split(stockIDEntrySelect.Text, "|")[0]
							locationName := strings.Split(stockIDEntrySelect.Text, "|")[1]
							materialId := materialsMap[[2]string{stockId, locationName}]
							jobTicket := jobTicketInput.Text
							notes := notesInput.Text

							_, err := db.Exec(`
							UPDATE materials
							SET quantity = (quantity - $1),
								notes = $2
							WHERE material_id = $3;
							`, quantity, notes, materialId,
							)

							if err != nil {
								dialog.ShowInformation("Error", "Updating material error: "+err.Error(), myWindow)
							} else {
								err := addTranscation(&TransactionInfo{
									materialId: materialId,
									stockId:    stockId,
									quantity:   -quantity,
									notes:      notes,
									jobTicket:  jobTicket,
									updatedAt:  time.Now(),
								}, db)

								if err != nil {
									dialog.ShowInformation("Error", "Updating transactions error: "+err.Error(), myWindow)
								} else {
									dialog.ShowInformation("Success", "Material updated", myWindow)
								}
							}
						}
					}, myWindow)

				dialogMaterial.Resize(fyne.NewSize(600, 300))
				dialogMaterial.Show()
			}
		}, myWindow)

	dialogCustomer.Resize(fyne.NewSize(300, 100))
	dialogCustomer.Show()
}

// Move a material between locations
func moveMaterial(myWindow fyne.Window, db *sql.DB) {
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

	var materials []Material
	var materialsStr []string
	materialsMap := make(map[[2]string]int)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			materialsStr = append(materialsStr, material.StockID+"|"+material.LocationName)
			materialsMap[[2]string{material.StockID, material.LocationName}] = material.MaterialID
		}
	})

	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "Cancel", customerSelector,
		func(confirm bool) {
			log.Println(confirm, len(customerSelector.Selected))
			if confirm && customerSelector.Selected != "" {
				stockIDEntrySelect := widget.NewSelectEntry(materialsStr)
				locationSelect := widget.NewSelect(locationsStr, func(s string) {})
				quantityInput := widget.NewEntry()
				costInput := widget.NewEntry()
				notesInput := widget.NewEntry()

				dialogMaterial := dialog.NewForm("Move material to another location", "Move", "Cancel",
					[]*widget.FormItem{
						widget.NewFormItem("Stock ID", stockIDEntrySelect),
						widget.NewFormItem("New Location", locationSelect),
						widget.NewFormItem("Move Quantity", quantityInput),
						widget.NewFormItem("Cost (within warehouse)", costInput),
						widget.NewFormItem("Notes", notesInput),
					},
					func(confirm bool) {
						if confirm {
							stockId := strings.Split(stockIDEntrySelect.Text, "|")[0]
							currLocationName := strings.Split(stockIDEntrySelect.Text, "|")[1]
							currMaterialId := materialsMap[[2]string{stockId, currLocationName}]
							currentLocationId := locationsMap[currLocationName]

							newLocationId := locationsMap[locationSelect.Selected]
							quantity, _ := strconv.Atoi(quantityInput.Text)
							notes := notesInput.Text
							cost, _ := strconv.Atoi(costInput.Text)

							var currMaterial MaterialInfo

							// Update material in the current location
							err := db.QueryRow(`
								UPDATE materials
								SET quantity = (quantity - $1),
									notes = $2
								WHERE material_id = $3 AND location_id = $4
								RETURNING material_id, stock_id, location_id, customer_id, material_type,
										description, notes, quantity, updated_at;
							`, quantity, notes, currMaterialId, currentLocationId,
							).Scan(&currMaterial.materialId,
								&currMaterial.stockId,
								&currMaterial.locationId,
								&currMaterial.customerId,
								&currMaterial.materialType,
								&currMaterial.description,
								&currMaterial.notes,
								&currMaterial.quantity,
								&currMaterial.updated_at)

							if err != nil {
								log.Println("upd1", err)
								dialog.ShowInformation("Error", err.Error(), myWindow)
							} else {
								// Update material in the new location
								var newMaterial MaterialInfo
								rows, err := db.Query(`
								UPDATE materials
								SET quantity = (quantity + $1)
								WHERE stock_id = $2 and location_id = $3
								RETURNING material_id;
							`, quantity, stockId, newLocationId,
								)

								if err != nil {
									log.Println("upd2", err)
									dialog.ShowInformation("Error", err.Error(), myWindow)
								}
								for rows.Next() {
									err := rows.Scan(&newMaterial.materialId)
									if err != nil {
										log.Println("scan", err)
									}
								}

								// If there is no the material in the destination location
								// Then add the material in there
								if newMaterial.materialId == 0 {
									err := db.QueryRow(`
									INSERT INTO materials
										(stock_id, location_id,
										customer_id, material_type, description, notes, quantity, updated_at)
										VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
										RETURNING material_id;`,
										stockId, newLocationId,
										currMaterial.customerId, currMaterial.materialType, currMaterial.description,
										currMaterial.notes, quantity, time.Now()).Scan(&newMaterial.materialId)

									if err != nil {
										log.Println("upd3", err)
										dialog.ShowInformation("Error", err.Error(), myWindow)
									}
								}

								addTranscation(&TransactionInfo{
									materialId: currMaterial.materialId,
									stockId:    stockId,
									quantity:   -quantity,
									notes:      notes,
									cost:       cost,
									updatedAt:  time.Now(),
								}, db)

								addTranscation(&TransactionInfo{
									materialId: newMaterial.materialId,
									stockId:    stockId,
									quantity:   quantity,
									notes:      notes,
									cost:       cost,
									updatedAt:  time.Now(),
								}, db)

								dialog.ShowInformation("Success", strconv.Itoa(quantity)+" of "+
									stockId+" has been moved from "+strings.Split(stockIDEntrySelect.Text, "|")[1]+
									" to "+locationSelect.Selected, myWindow)
							}
						}
					}, myWindow)

				dialogMaterial.Resize(fyne.NewSize(600, 400))
				dialogMaterial.Show()
			}
		}, myWindow)

	dialogCustomer.Resize(fyne.NewSize(300, 100))
	dialogCustomer.Show()
}
