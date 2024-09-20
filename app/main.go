package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lib/pq"
)

type Materials struct {
	MaterialID   int       `field:"material_id"`
	StockID      string    `field:"stock_id"`
	LocationID   string    `field:"location_id"`
	MaterialType string    `field:"type"`
	Description  string    `field:"description"`
	Notes        string    `field:"notes"`
	Quantity     int       `field:"quantity"`
	UpdatedAt    time.Time `field:"updated_at"`
	CustomerID   int       `field:"customer_id"`
}

func main() {
	// Database connection
	db, err := connectToDB()
	if err != nil {
		fmt.Println(err.Error())
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("Inventory Management")

	invList := getInvList(db)

	materialsTable := widget.NewTable(
		func() (int, int) {
			return len(invList), len(invList[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(invList[i.Row][i.Col])
		})

	content := container.NewVBox(
		widget.NewLabel("Inventory Maanagement"),
		materialsTable,
		widget.NewButton("Add Customer", func() { addCustomer(myWindow, db) }),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1000, 400))
	myWindow.ShowAndRun()
}

func getInvList(db *sql.DB) [][]string {
	rows, err := db.Query("SELECT * FROM materials;")
	if err != nil {
		fmt.Printf("Error: %e", err)
	}

	invList := [][]string{
		{"Material ID", "Stock ID", "Location ID", "Material Type", "Description", "Notes", "Quantity", "Updated At", "Customer ID"},
	}

	for rows.Next() {
		inv := Materials{}

		s := reflect.ValueOf(&inv).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Printf("Error: %e", err)
		}

		invList = append(invList, []string{
			strconv.Itoa(inv.MaterialID),
			inv.StockID,
			inv.LocationID,
			inv.MaterialType,
			inv.Description,
			inv.Notes,
			strconv.Itoa(inv.Quantity),
			inv.UpdatedAt.String(),
			strconv.Itoa(inv.CustomerID),
		})
	}

	return invList
}

func addCustomer(myWindow fyne.Window, db *sql.DB) {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter Customer Name")

	dialog := dialog.NewForm("Add Customer", "Submit", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", input),
		}, func(confirm bool) {
			if confirm && input.Text != "" {
				if _, err := db.Exec("INSERT INTO customers (name) VALUES ($1)", input.Text); err != nil {
					log.Println("Error adding customer:", err)
				}
			}
		}, myWindow)

	dialog.Show()
}
