package main

import (
	"database/sql"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Warehouse struct {
	warehouseId int    `field:"warehouse_id"`
	name        string `field:"name"`
}

func fetchWarehouses(db *sql.DB) ([]Warehouse, error) {
	rows, err := db.Query("SELECT * FROM warehouses;")
	if err != nil {
		log.Println("Error fetchWarehouses1: ", err)
		return nil, err
	}
	defer rows.Close()

	var warehouses []Warehouse

	for rows.Next() {
		var warehouse Warehouse
		if err := rows.Scan(&warehouse.warehouseId, &warehouse.name); err != nil {
			log.Println("Error fetchWarehouses2: ", err)
			return warehouses, err
		}
		warehouses = append(warehouses, warehouse)
	}
	if err = rows.Err(); err != nil {
		return warehouses, err
	}

	return warehouses, nil
}

func addWarehouse(myWindow fyne.Window, db *sql.DB) {
	warehouses, err := fetchWarehouses(db)
	if err != nil {
		log.Println("Error fetching warehouses:", err)
		dialog.ShowInformation("Error", err.Error(), myWindow)
	}

	var warehousesStr []string
	warehousesMap := make(map[string]int)
	for _, warehouse := range warehouses {
		warehousesStr = append(warehousesStr, warehouse.name)
		warehousesMap[warehouse.name] = warehouse.warehouseId
	}

	nameSelectInput := widget.NewSelectEntry(warehousesStr)
	locNameInput := widget.NewEntry()

	dialog := dialog.NewForm("Add Warehouse", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Warehouse", nameSelectInput),
			widget.NewFormItem("Location", locNameInput),
		}, func(confirm bool) {
			if confirm {
				var warehouseId int

				if strings.TrimSpace(nameSelectInput.Text) == "" ||
					strings.TrimSpace(locNameInput.Text) == "" {
					dialog.ShowInformation("Error", "All fields must be filled", myWindow)
				} else {
					id, ok := warehousesMap[nameSelectInput.Text]

					if !ok {
						err := db.QueryRow(`INSERT INTO warehouses(name) VALUES($1)
											RETURNING warehouse_id;`,
							nameSelectInput.Text).Scan(&warehouseId)
						if err != nil {
							log.Println("Error adding warehouse:", err)
							dialog.ShowInformation("Error", err.Error()+"\nChoose existing warehouse from the dropdown list", myWindow)
						}
					} else {
						warehouseId = id
					}

					if warehouseId != 0 {
						_, err := db.Exec("INSERT INTO locations(name, warehouse_id) VALUES ($1,$2);",
							locNameInput.Text, warehouseId)

						if err != nil {
							log.Println("Error adding location:", err)
							dialog.ShowInformation("Error", err.Error(), myWindow)
						} else {
							dialog.ShowInformation("Success", "Location \""+locNameInput.Text+"\" has been attached to Warehouse \""+nameSelectInput.Text+"\"", myWindow)
						}
					}
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(300, 200))
	dialog.Show()
}
