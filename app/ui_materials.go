package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var materialTypes = []string{"Envelope", "Card", "Carrier", "Insert", "Consumables"}

type Location struct {
	id          int    `field:"location_id"`
	name        string `field:"name"`
	warehouseID int    `field:"warehouse_id"`
}

type Customer struct {
	id   int    `field:"customer_id"`
	name string `field:"name"`
	code string `field:"customer_code"`
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
	updatedAt    time.Time `field:"updated_at"`
	isActive     bool      `field:"is_active"`
	cost         float64   `field:"cost"`
	minQty       int       `field:"min_required_quantity"`
	maxQty       int       `field:"max_required_quantity"`
	owner        string    `field:"onwer"`
}

type IncomingMaterial struct {
	ShippingID   int     `field:"shipping_id"`
	CustomerName string  `field:"customer_name"`
	StockID      string  `field:"stock_id"`
	Cost         float64 `field:"cost"`
	Quantity     int     `field:"quantity"`
	MinQty       int     `field:"min_required_quantity"`
	MaxQty       int     `field:"max_required_quantity"`
	Notes        string  `field:"notes"`
	IsActive     bool    `field:"is_active"`
	MaterialType string  `field:"type"`
	Owner        string  `field:"owner"`
}

type TransactionInfo struct {
	materialId    int       `field:"material_id"`
	stockId       string    `field:"stock_id"`
	quantity      int       `field:"quantity_change"`
	notes         string    `field:"notes"`
	cost          float64   `field:"cost"`
	updatedAt     time.Time `field:"updated_at"`
	jobTicket     string    `field:"job_ticket"`
	isMove        bool      // opts
	newMaterialId int       // opts
}

type MaterialOpts struct {
	shippingId   int
	customerName string
	stockID      string
	quantity     int
	minQty       int
	maxQty       int
	cost         float64
	materialType string
	isActive     bool
	notes        string
	owner        string
}

type LocationOpts struct {
	customerId int
	stockId    string
}

//////////////////////////////////////////
// FETCH DATA FROM THE DB
//////////////////////////////////////////

func getIncomingMaterialsNumber(db *sql.DB) int {
	var incomingMaterialsQty int
	row := db.QueryRow("SELECT count(shipping_id) FROM incoming_materials;")
	row.Scan(&incomingMaterialsQty)
	return incomingMaterialsQty
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
		if err := rows.Scan(&customer.id, &customer.name, &customer.code); err != nil {
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

func fetchAvailableLocations(db *sql.DB, locOpts *LocationOpts) ([]Location, error) {
	rows, err := db.Query(`
		SELECT l.location_id, l.name, l.warehouse_id 
		FROM locations l
		LEFT JOIN materials m ON m.location_id = l.location_id
		WHERE
			(m.customer_id = $1 AND m.stock_id = $2)
			OR m.material_id IS NULL`,
		locOpts.customerId, locOpts.stockId)
	if err != nil {
		log.Println("Error fetchAvailableLocations1: ", err)
		return nil, err
	}
	defer rows.Close()

	var locations []Location

	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.id, &location.name, &location.warehouseID); err != nil {
			log.Println("Error fetchAvailableLocations2: ", err)
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
							m.notes, m.quantity, m.updated_at, c.name, m.material_type,
							m.cost, m.owner
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
			&material.Cost,
			&material.Owner,
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
	if trx.quantity < 0 {
		removingQty := int(math.Abs(float64(trx.quantity)))

		emptyCost := []string{"0"}

		for removingQty > 0 {
			var transactionId int
			var cost float64
			var remainingQty int

			// Find a last deduction
			db.QueryRow(`
				SELECT transaction_id, cost, remaining_quantity FROM transactions_log
				WHERE material_id = $1 AND stock_id = $2 AND quantity_change < 0
					AND cost NOT IN (`+strings.Join(emptyCost, ",")+`)
				ORDER BY transaction_id DESC LIMIT 1;
						`,
				trx.materialId,
				trx.stockId).Scan(&transactionId, &cost, &remainingQty)

			// First deduction is NOT found
			if transactionId == 0 {
				db.QueryRow(`
					SELECT transaction_id, cost, remaining_quantity FROM transactions_log
					WHERE material_id = $1 AND stock_id = $2  AND quantity_change > 0
						AND cost NOT IN (`+strings.Join(emptyCost, ",")+`)
					ORDER BY transaction_id LIMIT 1;
							`,
					trx.materialId,
					trx.stockId,
				).Scan(&transactionId, &cost, &remainingQty)

				// When neither positive nor negative calculations found
				if transactionId == 0 {
					return errors.New("no remains found")
				}

				// First deduction is found, but remains are zero
			} else if transactionId != 0 && remainingQty == 0 {
				emptyCost = append(emptyCost, strconv.FormatFloat(cost, 'f', -1, 64))
				continue
			}

			// Deduct from the balance
			if remainingQty < removingQty {
				removingQty -= remainingQty

				_, errInsert := db.Exec(
					`INSERT INTO transactions_log
							(material_id, stock_id, quantity_change, notes,
							cost, job_ticket, updated_at, remaining_quantity)
							 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
							 `, trx.materialId, trx.stockId, -remainingQty, trx.notes,
					cost, trx.jobTicket, trx.updatedAt, 0)

				if errInsert != nil {
					log.Println("err1", errInsert)
					return errInsert
				}

				emptyCost = append(emptyCost, strconv.FormatFloat(cost, 'f', -1, 64))

				if trx.isMove {
					addTranscation(&TransactionInfo{
						materialId: trx.newMaterialId,
						stockId:    trx.stockId,
						quantity:   remainingQty,
						notes:      trx.notes,
						cost:       cost,
						updatedAt:  trx.updatedAt,
						jobTicket:  trx.jobTicket,
					}, db)
				}
			} else if remainingQty >= removingQty {
				remainingQty -= removingQty

				_, errInsert := db.Exec(
					`INSERT INTO transactions_log
							(material_id, stock_id, quantity_change, notes,
							cost, job_ticket, updated_at, remaining_quantity)
							 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
							 `, trx.materialId, trx.stockId, -removingQty, trx.notes,
					cost, trx.jobTicket, trx.updatedAt, remainingQty)

				if errInsert != nil {
					log.Println("err2", errInsert)
					return errInsert
				}

				if trx.isMove {
					addTranscation(&TransactionInfo{
						materialId: trx.newMaterialId,
						stockId:    trx.stockId,
						quantity:   removingQty,
						notes:      trx.notes,
						cost:       cost,
						updatedAt:  trx.updatedAt,
						jobTicket:  trx.jobTicket,
					}, db)
				}

				removingQty = 0
			}
		}
	} else {
		// Check if an ID with the same cost exists
		var transactionId int
		db.QueryRow(`
				SELECT transaction_id FROM transactions_log
				WHERE
					material_id = $1 AND
					stock_id = $2 AND
					quantity_change > 0 AND
					cost = $3
				ORDER BY transaction_id DESC LIMIT 1;
						`,
			trx.materialId, trx.stockId, trx.cost).Scan(&transactionId)

		// If the ID exists then update it
		if transactionId > 0 {
			_, e := db.Query(`
				UPDATE transactions_log
				SET quantity_change = quantity_change + $2,
					remaining_quantity = remaining_quantity + $2,
					updated_at = NOW()
				WHERE transaction_id = $1

		`, transactionId, trx.quantity)

			if e != nil {
				return e
			}
		} else {
			// If an ID doesn't exist then add a new one
			_, e := db.Exec(
				`INSERT INTO transactions_log
			(material_id, stock_id, quantity_change, notes,
			cost, job_ticket, updated_at, remaining_quantity)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 `, trx.materialId, trx.stockId, trx.quantity, trx.notes,
				trx.cost, trx.jobTicket, trx.updatedAt, trx.quantity)

			if e != nil {
				return e
			}
		}
	}

	return nil
}

func deleteIncomingMaterial(db *sql.DB, shippingId int) error {
	if _, err := db.Exec(`
			DELETE FROM incoming_materials WHERE shipping_id = $1;`,
		shippingId); err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////
// ACTIONS
// Update the current materials quantity within locations
//////////////////////////////////////////////////////////

// Send a material for warehouse handling
func sendMaterial(myWindow fyne.Window, db *sql.DB) {
	customers, _ := fetchCustomers(db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	customerInputSelector := widget.NewSelect(customersStr, func(s string) {})
	customerInputSelector.SetSelected(customersStr[0])

	stockIDInput := widget.NewEntry()
	stockIDInput.Validator = validation.NewRegexp(".+", "At least one character")

	typeSelector := widget.NewSelect(materialTypes, func(s string) {})
	typeSelector.SetSelected(materialTypes[0])

	quantityInput := widget.NewEntry()
	quantityInput.Validator = validation.NewRegexp(
		`^[1-9][0-9]*$`,
		"Positive numbers greater than 0 only",
	)

	costInput := widget.NewEntry()
	costInput.Validator = validation.NewRegexp(
		`^(0*[1-9][0-9]*(\.[0-9]+)?|0+\.[0-9]*[1-9][0-9]*)$`,
		"Positive numbers greater than 0 only",
	)

	minRequiredQtyInput := widget.NewEntry()
	maxRequiredQtyInput := widget.NewEntry()
	descrInput := widget.NewEntry()
	ownerChkBox := widget.NewCheck("", func(b bool) {})
	isActiveChkBox := widget.NewCheck("", func(b bool) {})

	dialog := dialog.NewForm("Sending a Material to the Warehouse", "Send", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Customer *", customerInputSelector),
			widget.NewFormItem("Stock ID *", stockIDInput),
			widget.NewFormItem("Type *", typeSelector),
			widget.NewFormItem("Quantity *", quantityInput),
			widget.NewFormItem("Unit Cost, USD *", costInput),
			widget.NewFormItem("Min Quantity", minRequiredQtyInput),
			widget.NewFormItem("Max Quantity", maxRequiredQtyInput),
			widget.NewFormItem("Description", descrInput),
			widget.NewFormItem("TAG ownership", ownerChkBox),
			widget.NewFormItem("Allow for use", isActiveChkBox),
		}, func(confirm bool) {
			if confirm {
				floatCost, _ := strconv.ParseFloat((strings.Replace(costInput.Text, ",", "", -1)), 32)
				cost := math.Round(floatCost*100) / 100

				owner := "Tag"
				if !ownerChkBox.Checked {
					owner = "Customer"
				}

				_, err := db.Query(`
				INSERT INTO incoming_materials
					(customer_name, stock_id, cost, quantity,
					max_required_quantity, min_required_quantity,
					notes, is_active, type, owner)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
					customerInputSelector.Selected, stockIDInput.Text, cost,
					quantityInput.Text, maxRequiredQtyInput.Text, minRequiredQtyInput.Text,
					descrInput.Text, isActiveChkBox.Checked, typeSelector.Selected,
					owner,
				)

				if err != nil {
					dialog.ShowInformation("Error", "Sending a material error: "+err.Error(), myWindow)
				} else {
					dialog.ShowInformation("Success", "The material has been sent to the Warehouse ", myWindow)
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(600, 400))
	dialog.Show()
}

func getIncomingMaterials(db *sql.DB) []IncomingMaterial {
	rows, err := db.Query(`SELECT * FROM incoming_materials;`)
	if err != nil {
		fmt.Printf("Error acceptIncomingMaterials1: %e", err)
	}

	materialsArr := []IncomingMaterial{}

	for rows.Next() {
		material := IncomingMaterial{}

		s := reflect.ValueOf(&material).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Printf("Error acceptIncomingMaterials2: %e", err)
		}

		materialsArr = append(materialsArr, IncomingMaterial{
			material.ShippingID,
			material.CustomerName,
			material.StockID,
			material.Cost,
			material.Quantity,
			material.MinQty,
			material.MaxQty,
			material.Notes,
			material.IsActive,
			material.MaterialType,
			material.Owner,
		})
	}

	return materialsArr
}

func acceptIncomingMaterials(app fyne.App, db *sql.DB) {
	window := app.NewWindow("Incoming Materials")

	materialsArr := getIncomingMaterials(db)

	materialWidgets := []fyne.CanvasObject{}

	for i := 0; i < len(materialsArr); i++ {
		material := materialsArr[i]

		materialWidgets = append(materialWidgets,
			container.New(layout.NewGridLayoutWithColumns(4),
				widget.NewLabel("Customer: "+material.CustomerName),
				widget.NewLabel("Stock ID: "+material.StockID),
				widget.NewLabel("Quantity: "+strconv.Itoa(material.Quantity)),
				widget.NewButton("Add", func() {

					var materialOpts = MaterialOpts{
						shippingId:   material.ShippingID,
						customerName: material.CustomerName,
						stockID:      material.StockID,
						quantity:     material.Quantity,
						maxQty:       material.MaxQty,
						minQty:       material.MinQty,
						cost:         material.Cost,
						materialType: material.MaterialType,
						isActive:     material.IsActive,
						notes:        material.Notes,
						owner:        material.Owner,
					}

					createMaterial(app, window, db, &materialOpts)
				}),
			),
			widget.NewSeparator(),
		)
	}

	vBox := container.New(layout.NewVBoxLayout(), materialWidgets...)
	window.SetContent(vBox)
	window.Resize(fyne.NewSize(800, 700))
	window.Show()
}

// Create a new material in a location
func createMaterial(app fyne.App, myWindow fyne.Window, db *sql.DB, materialOpts *MaterialOpts) {
	customers, _ := fetchCustomers(db)
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersMap[customer.name] = customer.id
	}

	locations, _ := fetchAvailableLocations(
		db,
		&LocationOpts{
			customerId: customersMap[materialOpts.customerName],
			stockId:    materialOpts.stockID,
		},
	)

	var locationsStr []string
	locationsMap := make(map[string]int)
	for _, location := range locations {
		locationsStr = append(locationsStr, location.name)
		locationsMap[location.name] = location.id
	}

	customerLabel := widget.NewLabel(materialOpts.customerName)
	locationSelector := widget.NewSelect(locationsStr, func(s string) {})
	typeLabel := widget.NewLabel(materialOpts.materialType)
	stockIDLabel := widget.NewLabel(materialOpts.stockID)
	descrLabel := widget.NewLabel(materialOpts.notes)
	quantityInput := widget.NewEntry()
	notesInput := widget.NewEntry()
	ownerLabel := widget.NewLabel(materialOpts.owner)

	isActive := "Yes"
	if !materialOpts.isActive {
		isActive = "No"
	}
	isActiveLabel := widget.NewLabel(isActive)

	quantityInput.SetText(strconv.Itoa(materialOpts.quantity))

	dialog := dialog.NewForm("Create Material", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Customer", customerLabel),
			widget.NewFormItem("Stock ID", stockIDLabel),
			widget.NewFormItem("Type", typeLabel),
			widget.NewFormItem("Ownership", ownerLabel),
			widget.NewFormItem("Allow for use", isActiveLabel),
			widget.NewFormItem("Description", descrLabel),
			widget.NewFormItem("Quantity *", quantityInput),
			widget.NewFormItem("Location *", locationSelector),
			widget.NewFormItem("Notes", notesInput),
		}, func(confirm bool) {
			if confirm {
				var material Material
				quantity, _ := strconv.Atoi(strings.Replace(quantityInput.Text, ",", "", -1))

				// Update material in the current location
				rows, err := db.Query(`
				UPDATE materials
				SET quantity = (quantity + $1)
				WHERE stock_id = $2
					AND location_id = $3
					AND owner = $4
				RETURNING material_id;
				`, quantity,
					materialOpts.stockID,
					locationsMap[locationSelector.Selected],
					materialOpts.owner,
				)

				if err != nil {
					log.Println("createMaterial upd2", err)
					dialog.ShowInformation("Error", err.Error(), myWindow)
				} else {
					for rows.Next() {
						err := rows.Scan(&material.MaterialID)
						if err != nil {
							log.Println("createMaterial scan", err)
						}
					}

					// If there is no the same material in the current location
					// Then add the material in the chosen one
					if material.MaterialID == 0 {
						err := db.QueryRow(`
					INSERT INTO materials
					(
						stock_id,
						location_id,
						customer_id,
						material_type,
						description,
						notes,
						quantity,
						updated_at,
						min_required_quantity,
						max_required_quantity,
						is_active,
						cost,
						owner
					)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING material_id;`,
							materialOpts.stockID,
							locationsMap[locationSelector.Selected],
							customersMap[materialOpts.customerName],
							materialOpts.materialType,
							materialOpts.notes,
							notesInput.Text,
							quantity,
							time.Now(),
							materialOpts.minQty,
							materialOpts.maxQty,
							materialOpts.isActive,
							materialOpts.cost,
							materialOpts.owner,
						).Scan(&material.MaterialID)

						if err != nil {
							log.Println("Error saving material:", err)
							dialog.ShowInformation("Error", "Unable to save data: "+err.Error(), myWindow)
							panic(err)
						}
					}

					// Remove the material from incoming
					err := deleteIncomingMaterial(db, materialOpts.shippingId)

					if err != nil {
						dialog.ShowInformation("Error", "Deleting incoming material: "+err.Error(), myWindow)
					} else {
						err := addTranscation(&TransactionInfo{
							materialId: material.MaterialID,
							stockId:    materialOpts.stockID,
							quantity:   quantity,
							notes:      notesInput.Text,
							updatedAt:  time.Now(),
							cost:       materialOpts.cost,
						}, db)

						if err != nil {
							dialog.ShowInformation("Error", "Updating transactions error: "+err.Error(), myWindow)
						} else {
							dialog.ShowConfirm("Material accepted", "Refresh the list?", func(confirm bool) {
								if confirm {
									myWindow.Close()
									acceptIncomingMaterials(app, db)
								}
							}, myWindow)
						}
					}
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(600, 400))
	dialog.Show()
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
	materialsMap := make(map[int][3]string)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			description := material.Notes
			if description == "" {
				description = "No description"
			}
			materialsStr = append(materialsStr,
				material.LocationName+" | "+
					material.StockID+" | "+
					material.Owner)

			materialsMap[material.MaterialID] = [3]string{
				material.StockID,
				material.LocationName,
				material.Owner}
		}
	})

	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "", customerSelector,
		func(confirm bool) {
			if confirm && customerSelector.Selected != "" {
				stockIDSelect := widget.NewSelect(materialsStr, func(s string) {})
				quantityInput := widget.NewEntry()
				notesInput := widget.NewEntry()
				jobTicketInput := widget.NewEntry()

				dialogMaterial := dialog.NewForm("Remove material", "Remove", "Cancel",
					[]*widget.FormItem{
						widget.NewFormItem("Stock ID *", stockIDSelect),
						widget.NewFormItem("Remove Quantity *", quantityInput),
						widget.NewFormItem("Job Ticket *", jobTicketInput),
						widget.NewFormItem("Notes", notesInput),
					},
					func(confirm bool) {
						if confirm {
							quantity, _ := strconv.Atoi(strings.Replace(quantityInput.Text, ",", "", -1))
							locationName := strings.Split(stockIDSelect.Selected, " | ")[0]
							stockId := strings.Split(stockIDSelect.Selected, " | ")[1]
							owner := strings.Split(stockIDSelect.Selected, " | ")[2]

							var materialId int
							valuesArr := [3]string{stockId, locationName, owner}
							for k, v := range materialsMap {
								if reflect.DeepEqual(v, valuesArr) {
									materialId = k
								}
							}

							jobTicket := jobTicketInput.Text
							notes := notesInput.Text

							// Verify that we have the remaining materials
							var actualQuantity int
							db.QueryRow(`SELECT quantity FROM materials WHERE material_id = $1`, materialId).Scan(&actualQuantity)

							if actualQuantity < quantity {
								dialog.ShowInformation(
									"Error",
									`The removing quantity (`+strconv.Itoa(quantity)+`) is more than the actual one (`+strconv.Itoa(actualQuantity)+`)`,
									myWindow)
							} else {
								var err error

								if actualQuantity == quantity {
									_, err = db.Exec(`
										DELETE FROM materials
										WHERE material_id = $1;
								`, materialId)
								} else {
									// Update the material quantity
									_, err = db.Exec(`
										UPDATE materials
										SET quantity = (quantity - $1),
											notes = $2
										WHERE material_id = $3;
								`, quantity, notes, materialId,
									)
								}

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
										dialog.ShowInformation("Success", "Material has been removed. The remaining quantity: "+strconv.Itoa(actualQuantity-quantity), myWindow)
									}
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

	var materials []Material
	var materialsStr []string
	materialsMap := make(map[int][3]string)

	customerSelector := widget.NewSelect(customersStr, func(customerName string) {
		customerId := customersMap[customerName]
		materials, _ = fetchMaterialsByCustomer(db, customerId)
		for _, material := range materials {
			description := material.Notes
			if description == "" {
				description = "No description"
			}
			materialsStr = append(materialsStr,
				material.LocationName+" | "+
					material.StockID+" | "+
					material.Owner)
			materialsMap[material.MaterialID] = [3]string{
				material.StockID,
				material.LocationName,
				material.Owner,
			}
		}
	})

	// Customer Dialog
	dialogCustomer := dialog.NewCustomConfirm("Choose customer", "OK", "Cancel", customerSelector,
		func(confirm bool) {
			if confirm && customerSelector.Selected != "" {
				stockIDSelector := widget.NewSelect(materialsStr, func(s string) {})

				// Stock ID dialog
				dialogStockID := dialog.NewCustomConfirm("Choose Stock ID to move", "OK", "Cancel", stockIDSelector, func(confirm bool) {
					if confirm && stockIDSelector.Selected != "" {
						stockId := strings.Split(stockIDSelector.Selected, " | ")[1]

						// Get empty OR the same stock ID locations
						locations, _ := fetchAvailableLocations(
							db,
							&LocationOpts{
								customerId: customersMap[customerSelector.Selected],
								stockId:    stockId,
							},
						)
						var locationsStr []string
						locationsMap := make(map[string]int)
						for _, location := range locations {
							locationsStr = append(locationsStr, location.name)
							locationsMap[location.name] = location.id
						}

						locationSelector := widget.NewSelect(locationsStr, func(s string) {})
						quantityInput := widget.NewEntry()
						notesInput := widget.NewEntry()

						// Material move dialog
						dialogMaterial := dialog.NewForm(stockIDSelector.Selected, "Move", "Cancel",
							[]*widget.FormItem{
								widget.NewFormItem("New Location *", locationSelector),
								widget.NewFormItem("Move Quantity *", quantityInput),
								widget.NewFormItem("Notes", notesInput),
							},
							func(confirm bool) {
								if confirm {
									currLocationName := strings.Split(stockIDSelector.Selected, " | ")[0]
									owner := strings.Split(stockIDSelector.Selected, " | ")[2]

									var currMaterialId int
									valuesArr := [3]string{stockId, currLocationName, owner}
									for k, v := range materialsMap {
										if reflect.DeepEqual(v, valuesArr) {
											currMaterialId = k
										}
									}

									currentLocationId := locationsMap[currLocationName]
									newLocationId := locationsMap[locationSelector.Selected]
									quantity, _ := strconv.Atoi(strings.Replace(quantityInput.Text, ",", "", -1))
									notes := notesInput.Text

									var currMaterial MaterialInfo

									var actualQuantity int
									db.QueryRow(`SELECT quantity FROM materials WHERE material_id = $1`, currMaterialId).Scan(&actualQuantity)

									// Check whether remaining quantity exists
									if actualQuantity < quantity {
										dialog.ShowInformation(
											"Error",
											`The moving quantity (`+strconv.Itoa(quantity)+`) is more than the actual one (`+strconv.Itoa(actualQuantity)+`)`,
											myWindow)
									} else {

										// Update material in the current location
										err := db.QueryRow(`
												UPDATE materials
												SET quantity = (quantity - $1),
													notes = $2
												WHERE material_id = $3 AND location_id = $4
												RETURNING material_id, stock_id, location_id, customer_id, material_type,
														description, notes, quantity, updated_at, is_active, cost,
														min_required_quantity, max_required_quantity, owner;
													`, quantity, notes, currMaterialId, currentLocationId,
										).Scan(
											&currMaterial.materialId,
											&currMaterial.stockId,
											&currMaterial.locationId,
											&currMaterial.customerId,
											&currMaterial.materialType,
											&currMaterial.description,
											&currMaterial.notes,
											&currMaterial.quantity,
											&currMaterial.updatedAt,
											&currMaterial.isActive,
											&currMaterial.cost,
											&currMaterial.minQty,
											&currMaterial.maxQty,
											&currMaterial.owner,
										)

										if err != nil {
											log.Println("moveMaterial upd1", err)
											dialog.ShowInformation("Error", err.Error(), myWindow)
										} else {
											// Update material in the new location
											var newMaterial MaterialInfo
											rows, err := db.Query(`
												UPDATE materials
												SET quantity = (quantity + $1)
												WHERE
													stock_id = $2 AND
													location_id = $3 AND
													owner = $4
												RETURNING material_id;
											`, quantity, stockId, newLocationId, owner,
											)

											if err != nil {
												log.Println("moveMaterial upd2", err)
												dialog.ShowInformation("Error", err.Error(), myWindow)
											}
											for rows.Next() {
												err := rows.Scan(&newMaterial.materialId)
												if err != nil {
													log.Println("moveMaterial scan", err)
												}
											}

											// If there is no the material in the destination location
											// Then add the material in there
											if newMaterial.materialId == 0 {
												err := db.QueryRow(`
														INSERT INTO materials
															(stock_id, location_id,
															customer_id, material_type, description, notes, quantity, updated_at,
															cost, is_active, min_required_quantity, max_required_quantity, owner)
															VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
															RETURNING material_id;`,
													stockId, newLocationId,
													currMaterial.customerId, currMaterial.materialType, currMaterial.description,
													currMaterial.notes, quantity, time.Now(), currMaterial.cost, currMaterial.isActive,
													currMaterial.minQty, currMaterial.maxQty, currMaterial.owner).
													Scan(&newMaterial.materialId)

												if err != nil {
													log.Println("upd3", err)
													dialog.ShowInformation("Error", err.Error(), myWindow)
													panic(err)
												}
											}

											addTranscation(&TransactionInfo{
												materialId:    currMaterial.materialId,
												stockId:       stockId,
												quantity:      -quantity,
												notes:         notes,
												cost:          currMaterial.cost,
												updatedAt:     time.Now(),
												isMove:        true,
												newMaterialId: newMaterial.materialId,
											}, db)

											dialog.ShowInformation("Success", strconv.Itoa(quantity)+" of "+
												stockId+" has been moved from "+currLocationName+
												" to "+locationSelector.Selected, myWindow)
										}
									}
								}
							}, myWindow)

						dialogMaterial.Resize(fyne.NewSize(600, 300))
						dialogMaterial.Show()
					}
				}, myWindow)

				dialogStockID.Resize(fyne.NewSize(300, 100))
				dialogStockID.Show()
			}
		}, myWindow)

	dialogCustomer.Resize(fyne.NewSize(300, 100))
	dialogCustomer.Show()
}
