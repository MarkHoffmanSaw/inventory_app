package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lib/pq"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Tag Systems USA Inventory Management v1.0")

	// Database connection
	db, err := connectToDB()
	if err != nil {
		log.Println(err.Error())
		myWindow.SetContent(widget.NewLabel("Database error: " + err.Error() + "/n Call to the IT department for help"))
		myWindow.Resize(fyne.NewSize(100, 100))
		myWindow.ShowAndRun()
	} else {
		buttonsContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Customer settings"),
			widget.NewButton("Add customer", func() { addCustomer(myWindow, db) }),
			widget.NewLabel("Warehouse settings"),
			widget.NewButton("Add warehouse", func() { addWarehouse(myWindow, db) }),
			widget.NewLabel("Material settings"),
			widget.NewButton("Create material", func() { addMaterial(myWindow, db) }),
			widget.NewButton("Add material", func() {}),
			widget.NewButton("Move material", func() {}),
			widget.NewButton("Remove material", func() {}),
			widget.NewButton("Show inventory", func() { showInventory(myApp, db) }),
		)

		mainLabel := widget.NewLabel("Main Menu")
		mainLabel.TextStyle.Bold = true
		mainLabel.Alignment = fyne.TextAlignCenter

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			buttonsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(500, 500))
		myWindow.ShowAndRun()
	}
}
