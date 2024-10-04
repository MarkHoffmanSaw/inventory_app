package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lib/pq"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.LightTheme())
	myWindow := myApp.NewWindow("Tag Systems USA Inventory Management v1.0")

	// Database connection
	db, err := connectToDB()
	if err != nil {
		log.Println(err.Error())
		myWindow.SetContent(widget.NewLabel("Database error: " + err.Error() + "\nCall to the IT department for help"))
		myWindow.Resize(fyne.NewSize(100, 100))
		myWindow.ShowAndRun()
	} else {
		mainLabel := widget.NewLabel("Main Menu")
		mainLabel.TextStyle.Bold = true
		mainLabel.Alignment = fyne.TextAlignCenter

		customerContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Customer settings"),
			widget.NewButton("Add customer", func() { addCustomer(myWindow, db) }))

		warehouseContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Warehouse settings"),
			widget.NewButton("Add warehouse", func() { addWarehouse(myWindow, db) }),
		)

		materialContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Material settings"),
			widget.NewButton("Create material", func() { createMaterial(myWindow, db) }),
			widget.NewButton("Replenish material", func() { addMaterial(myWindow, db) }),
			widget.NewButton("Remove material", func() { removeMaterial(myWindow, db) }),
			widget.NewButton("Move material", func() { moveMaterial(myWindow, db) }),
			widget.NewButton("Handle a Material (CSR)", func() { createMaterial(myWindow, db) }),
		)

		infoContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Info Tables"),
			widget.NewButton("Show inventory list", func() { showInventory(myApp, db, myWindow) }),
			widget.NewButton("Show transactions", func() { showTransactions(myApp, db) }),
		)

		buttonsContainer := container.New(layout.NewGridLayoutWithColumns(4),
			customerContainer,
			warehouseContainer,
			materialContainer,
			infoContainer,
		)

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			buttonsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(800, 700))
		myWindow.ShowAndRun()
	}
}
