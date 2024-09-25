package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Materials struct {
	MaterialID   int       `field:"material_id"`
	StockID      string    `field:"stock_id"`
	LocationName string    `field:"location_id"`
	Description  string    `field:"description"`
	Notes        string    `field:"notes"`
	Quantity     int       `field:"quantity"`
	UpdatedAt    time.Time `field:"updated_at"`
	CustomerName string    `field:"customer_id"`
	MaterialType string    `field:"type"`
}

func showInventory(app fyne.App, db *sql.DB) {
	window := app.NewWindow("Inventory")
	materialsTable := getMaterialsTable(db)
	window.SetContent(materialsTable)
	window.Resize(fyne.NewSize(1600, 700))
	window.Show()
}

func getMaterialsTable(db *sql.DB) fyne.Widget {
	rows, err := db.Query(`SELECT m.material_id, m.stock_id, l.name, m.description,
							m.notes, m.quantity, m.updated_at, c.name, m.material_type
							FROM materials m
							LEFT JOIN locations l ON m.location_id = l.location_id
							LEFT JOIN customers c ON c.customer_id = m.customer_id`)
	if err != nil {
		fmt.Printf("Error getMaterialsTable1: %e", err)
	}

	invList := [][]string{
		{"Material ID", "Stock ID", "Location", "Material Type",
			"Description", "Notes", "Quantity", "Updated At", "Customer"},
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
			log.Printf("Error getMaterialsTable2: %e", err)
		}

		invList = append(invList, []string{
			strconv.Itoa(inv.MaterialID),
			inv.StockID,
			inv.LocationName,
			inv.MaterialType,
			inv.Description,
			inv.Notes,
			strconv.Itoa(inv.Quantity),
			inv.UpdatedAt.String(),
			inv.CustomerName,
		})
	}

	materialsTable := widget.NewTable(
		func() (int, int) {
			return len(invList), len(invList[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Materials")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(invList[i.Row][i.Col])
		})

	columnWidth := 150
	for col := 0; col < len(invList[0]); col++ {
		materialsTable.SetColumnWidth(col, float32(columnWidth))
	}
	materialsTable.SetColumnWidth(7, float32(300))

	return materialsTable
}
