package main

import (
	"database/sql"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func addCustomer(myWindow fyne.Window, db *sql.DB) {
	nameInput := widget.NewEntry()
	nameInput.Validator = validation.NewRegexp(`^\d*[a-zA-Z][a-zA-Z0-9]*$`, "Enter a name")
	codeInput := widget.NewEntry()
	codeInput.Validator = validation.NewRegexp(`^\d*[a-zA-Z][a-zA-Z0-9]*$`, "Enter a code")

	dialog := dialog.NewForm("Add Customer", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameInput),
			widget.NewFormItem("Code", codeInput),
		}, func(confirm bool) {
			if confirm {
				if _, err := db.Exec("INSERT INTO customers (name, customer_code) VALUES ($1,$2,$3)",
					nameInput.Text, codeInput.Text); err != nil {
					log.Println("Error adding customer:", err)
					dialog.ShowInformation("Error", err.Error(), myWindow)
				} else {
					dialog.ShowInformation("Success", "Customer added", myWindow)
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(600, 300))

	dialog.Show()
}
