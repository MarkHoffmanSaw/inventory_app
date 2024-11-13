package main

import (
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lib/pq"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.LightTheme())
	myWindow := myApp.NewWindow("Tag Systems USA Inventory Management v1.2")

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

		customerLabel := widget.NewLabel("Customer Service")
		customerLabel.TextStyle.Bold = true
		customerLabel.Alignment = fyne.TextAlignCenter

		customerContainer := container.New(layout.NewCustomPaddedVBoxLayout(10),
			customerLabel,
			widget.NewButton("Add Customer", func() { addCustomer(myWindow, db) }),
			widget.NewButton("Send Material", func() { sendMaterial(myWindow, db) }),
		)

		// Incoming materials data
		incomingMaterialsData := binding.NewString()
		str := "Incoming Materials: "
		incomingMaterialsData.Set(str + strconv.Itoa(getIncomingMaterialsNumber(db)))
		incomingMaterialsLabel := widget.NewLabelWithData(incomingMaterialsData)
		incomingMaterialsLabel.Alignment = fyne.TextAlignCenter
		incomingMaterialsLabel.TextStyle.Bold = true

		warehouseLabel := widget.NewLabel("Warehouse")
		warehouseLabel.TextStyle.Bold = true
		warehouseLabel.Alignment = fyne.TextAlignCenter

		materialContainer := container.New(layout.NewCustomPaddedVBoxLayout(10),
			warehouseLabel,
			widget.NewButton("Add Location", func() { addWarehouse(myWindow, db) }),
			widget.NewSeparator(),
			incomingMaterialsLabel,
			widget.NewButton("Refresh Data", func() {
				incomingMaterialsQty := getIncomingMaterialsNumber(db)
				incomingMaterialsData.Set((str + strconv.Itoa(incomingMaterialsQty)))
			}),
			widget.NewSeparator(),
			widget.NewButton("Accept Materials", func() { acceptIncomingMaterials(myApp, db) }),
			widget.NewButton("Use Material", func() { removeMaterial(myWindow, db) }),
			widget.NewButton("Move Material to Location", func() { moveMaterial(myWindow, db) }),
		)

		reportsLabel := widget.NewLabel("Reports")
		reportsLabel.TextStyle.Bold = true
		reportsLabel.Alignment = fyne.TextAlignCenter

		report := Report{app: myApp, db: db, window: myWindow}
		inv := InventoryReport{Report: report}
		trx := TransactionReport{Report: report}
		blc := BalanceReport{Report: report}

		infoContainer := container.New(layout.NewCustomPaddedVBoxLayout(10),
			reportsLabel,
			widget.NewButton("Inventory List", func() { getReport(inv) }),
			widget.NewButton("Transactions Report", func() { getReport(trx) }),
			widget.NewButton("Balance Report", func() { getReport(blc) }),
		)

		actionsContainer := container.New(layout.NewGridLayoutWithColumns(3),
			customerContainer,
			materialContainer,
			infoContainer,
		)

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			actionsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(800, 600))
		myWindow.ShowAndRun()
	}
}
