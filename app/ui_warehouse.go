package main

import (
	"database/sql"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func addWarehouse(myWindow fyne.Window, db *sql.DB) {
	nameInput := widget.NewEntry()
	locNameInput := widget.NewEntry()

	dialog := dialog.NewForm("Add Warehouse", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Warehouse", nameInput),
			widget.NewFormItem("Location", locNameInput),
		}, func(confirm bool) {
			if confirm {
				var warehouseId int
				err := db.QueryRow("INSERT INTO warehouses(name) VALUES ($1) RETURNING warehouse_id;",
					nameInput.Text).Scan(&warehouseId)

				if err != nil {
					dialog.ShowInformation("Error", "Can't add new warehouse: "+err.Error(), myWindow)
					log.Println("Error adding warehouse:", err)
					dialog.ShowInformation("Error", err.Error(), myWindow)
				} else {
					_, err := db.Exec("INSERT INTO locations(name, warehouse_id) VALUES ($1,$2);",
						locNameInput.Text, warehouseId)

					if err != nil {
						dialog.ShowInformation("Error", "Can't add new location: "+err.Error(), myWindow)
						log.Println("Error adding location:", err)
						dialog.ShowInformation("Error", err.Error(), myWindow)
					} else {
						dialog.ShowInformation("Success", "Warehouse and location created", myWindow)
					}
				}

			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(300, 200))

	dialog.Show()
}
