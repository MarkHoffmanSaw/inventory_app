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
	myWindow := myApp.NewWindow("Tag Systems USA Inventory Management v1.1")

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

		materialContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Materials"),
			widget.NewButton("Add Warehouse/Location", func() { addWarehouse(myWindow, db) }),
			widget.NewButton("[CSR] Send Material to Warehouse", func() { sendMaterial(myWindow, db) }),
			widget.NewButton("[Pete] Incoming Materials", func() { acceptIncomingMaterials(myApp, db) }),
			widget.NewButton("[Pete] Use Material", func() { removeMaterial(myWindow, db) }),
			widget.NewButton("[Pete] Move Material to Location", func() { moveMaterial(myWindow, db) }),
		)

		infoContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Tables"),
			widget.NewButton("[Pete] Inventory List", func() { showInventory(myApp, db, myWindow) }),
			widget.NewButton("[John] Get Transactions Report", func() { showTransactions(myApp, db, myWindow) }),
			widget.NewButton("[John] Get Transactions Balance", func() { showBalance(myApp, db, myWindow) }),
			widget.NewButton("[John] Download Report (beta)", func() { downloadFinancialReport(db, myWindow) }),
		)

		warehouseActionsContainer := container.New(layout.NewGridLayoutWithColumns(3),
			customerContainer,
			materialContainer,
			infoContainer,
		)

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			warehouseActionsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(800, 600))
		myWindow.ShowAndRun()
	}
}
