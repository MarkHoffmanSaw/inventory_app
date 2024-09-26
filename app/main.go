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
		myWindow.SetContent(widget.NewLabel("Database error: " + err.Error() + "/n Call to the IT department for help"))
		myWindow.Resize(fyne.NewSize(100, 100))
		myWindow.ShowAndRun()
	} else {
		mainLabel := widget.NewLabel("Main Menu")
		mainLabel.TextStyle.Bold = true
		mainLabel.Alignment = fyne.TextAlignCenter

		buttonsContainer := container.New(layout.NewVBoxLayout(),
			widget.NewLabel("Customer settings"),
			widget.NewButton("Add customer", func() { addCustomer(myWindow, db) }),
			widget.NewLabel("Warehouse settings"),
			widget.NewButton("Add warehouse", func() { addWarehouse(myWindow, db) }),
			widget.NewLabel("Material settings"),
			widget.NewButton("Create a material", func() { createMaterial(myWindow, db) }),
			widget.NewButton("Add material", func() { addMaterial(myWindow, db) }),
			widget.NewButton("Remove material", func() { removeMaterial(myWindow, db) }),
			widget.NewButton("Move material", func() { moveMaterial(myWindow, db) }),
			widget.NewButton("Show inventory list", func() { showInventory(myApp, db) }),
		)

		content := container.New(layout.NewVBoxLayout(),
			mainLabel,
			buttonsContainer,
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(450, 500))
		myWindow.ShowAndRun()
	}
}
