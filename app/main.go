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
			widget.NewLabel("Customers"),
			widget.NewButton("Add a Customer", func() { addCustomer(myWindow, db) }))

		warehouseContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Warehouses"),
			widget.NewButton("Add a Warehouse/Location", func() { addWarehouse(myWindow, db) }),
		)

		materialContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Materials"),
			widget.NewButton("Accept incoming materials", func() { acceptIncomingMaterials(myApp, db) }),
			// widget.NewButton("Add a Material", func() { createMaterial(myWindow, db, MaterialOpts{}) }),
			// widget.NewButton("Replenish material", func() { addMaterial(myWindow, db) }),
			widget.NewButton("Remove a Material", func() { removeMaterial(myWindow, db) }),
			widget.NewButton("Move a Material", func() { moveMaterial(myWindow, db) }),
		)

		infoContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Tables"),
			widget.NewButton("Inventory List", func() { showInventory(myApp, db, myWindow) }),
			widget.NewButton("Transactions", func() { showTransactions(myApp, db) }),
		)

		CSRContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("CSR management"),
			widget.NewButton("Send a Material to the Warehouse", func() { acceptMaterial(myWindow, db) }),
		)

		warehouseActionsContainer := container.New(layout.NewGridLayoutWithColumns(4),
			customerContainer,
			warehouseContainer,
			materialContainer,
			infoContainer,
		)

		CSRActionsContainer := container.New(layout.NewGridLayoutWithColumns(4),
			CSRContainer,
		)

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			warehouseActionsContainer,
			widget.NewSeparator(),
			CSRActionsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(800, 700))
		myWindow.ShowAndRun()
	}
}
