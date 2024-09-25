package main

import (
	"database/sql"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func addCustomer(myWindow fyne.Window, db *sql.DB) {
	nameInput := widget.NewEntry()
	codeInput := widget.NewEntry()
	typeSelect := widget.NewRadioGroup([]string{"Internal", "External"}, func(s string) {})

	dialog := dialog.NewForm("Add Customer", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameInput),
			widget.NewFormItem("Code", codeInput),
			widget.NewFormItem("User type", typeSelect),
		}, func(confirm bool) {
			if confirm {
				if _, err := db.Exec("INSERT INTO customers (customer_name, customer_code, customer_type) VALUES ($1,$2,$3)",
					nameInput.Text, codeInput.Text, typeSelect.Selected); err != nil {
					log.Println("Error adding customer:", err)
				}
			}
		}, myWindow)

	dialog.Resize(fyne.NewSize(400, 300))

	dialog.Show()
}