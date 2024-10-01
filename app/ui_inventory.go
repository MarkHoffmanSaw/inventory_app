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

type Material struct {
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

type Transaction struct {
	TransactionId int       `field:"transaction_id"`
	MaterialId    int       `field:"material_id"`
	StockId       string    `field:"stock_id"`
	Quantity      int       `field:"quantity_change"`
	Notes         string    `field:"notes"`
	Cost          int       `field:"cost"`
	UpdatedAt     time.Time `field:"updated_at"`
	JobTicket     string    `field:"job_ticket"`

	// optional attributes are only for info
	LocationName  string `field:"location_name"`
	WarehouseName string `field:"warehouse_name"`
	CustomerName  string `field:"customer_name"`
}

func showInventory(app fyne.App, db *sql.DB) {
	window := app.NewWindow("Inventory")
	materialsTable := getMaterialsTable(db)
	window.SetContent(materialsTable)
	window.Resize(fyne.NewSize(1600, 700))
	window.Show()
}

func showTransactions(app fyne.App, db *sql.DB) {
	window := app.NewWindow("Transactions")
	transactionsTable := getTransactionsTable(db)
	window.SetContent(transactionsTable)
	window.Resize(fyne.NewSize(1700, 700))
	window.Show()
}

func getMaterialsTable(db *sql.DB) fyne.Widget {
	rows, err := db.Query(`SELECT m.material_id, m.stock_id, l.name, m.description,
							m.notes, m.quantity, m.updated_at, c.name, m.material_type
							FROM materials m
							LEFT JOIN locations l ON m.location_id = l.location_id
							LEFT JOIN customers c ON c.customer_id = m.customer_id
							ORDER BY m.updated_at ASC;`)
	if err != nil {
		fmt.Printf("Error getMaterialsTable1: %e", err)
	}

	invList := [][]string{
		{"Material ID", "Stock ID", "Location", "Material Type",
			"Description", "Notes", "Quantity", "Updated At", "Customer"},
	}

	for rows.Next() {
		inv := Material{}

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

func getTransactionsTable(db *sql.DB) fyne.Widget {
	rows, err := db.Query(`SELECT transaction_id, m.material_id as material_id,
							m.stock_id as stock_id, tl.quantity_change,
							tl.notes, tl.cost, tl.updated_at, tl.job_ticket,
							l.name as location_name, w.name as warehouse_name, c.name as customer_name
							FROM transactions_log tl
							LEFT JOIN materials m ON m.material_id = tl.material_id
							LEFT JOIN locations l ON m.location_id = l.location_id
							LEFT JOIN warehouses w ON l.warehouse_id = w.warehouse_id
							LEFT JOIN customers c ON m.customer_id = c.customer_id
							ORDER BY transaction_id`)
	if err != nil {
		fmt.Printf("Error getTransactionsTable1: %e", err)
	}

	trxList := [][]string{
		{
			"Transaction ID", "Stock ID",
			"Quantity Change", "Notes",
			"Cost", "Updated At",
			"Job Ticket", "Location",
			"Warehouse", "Customer",
		},
	}

	for rows.Next() {
		trx := Transaction{}

		s := reflect.ValueOf(&trx).Elem()
		numCols := s.NumField()

		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()

		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Printf("Error getTransactionsTable2: %e", err)
		}

		trxList = append(trxList, []string{
			strconv.Itoa(trx.TransactionId),
			trx.StockId,
			strconv.Itoa(trx.Quantity),
			trx.Notes,
			strconv.Itoa(trx.Cost),
			trx.UpdatedAt.String(),
			trx.JobTicket,
			trx.LocationName,
			trx.WarehouseName,
			trx.CustomerName,
		})
	}

	transactionsTable := widget.NewTable(
		func() (int, int) {
			return len(trxList), len(trxList[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Transactions")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(trxList[i.Row][i.Col])
		})

	columnWidth := 150
	for col := 0; col < len(trxList[0]); col++ {
		transactionsTable.SetColumnWidth(col, float32(columnWidth))
	}

	transactionsTable.SetColumnWidth(5, float32(300))

	return transactionsTable
}
