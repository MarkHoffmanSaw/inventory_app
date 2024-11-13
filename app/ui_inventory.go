package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/leekchan/accounting"
)

// STRUCTS
/////////////////////////////////

type Material struct {
	MaterialID     int       `field:"material_id"`
	StockID        string    `field:"stock_id"`
	LocationName   string    `field:"location_id"`
	Description    string    `field:"description"`
	Notes          string    `field:"notes"`
	Quantity       int       `field:"quantity"`
	MinRequiredQty int       `field:"min_required_quantity"`
	MaxRequiredQty int       `field:"min_required_quantity"`
	UpdatedAt      time.Time `field:"updated_at"`
	CustomerName   string    `field:"customer_id"`
	MaterialType   string    `field:"type"`
	IsActive       string    `field:"is_active"`
	Cost           float64   `field:"cost"`
	Owner          string    `field:"owner"`
}

type Transaction struct {
	TransactionId int       `field:"transaction_id"`
	MaterialId    int       `field:"material_id"`
	StockId       string    `field:"stock_id"`
	Quantity      int       `field:"quantity_change"`
	Notes         string    `field:"notes"`
	Cost          float64   `field:"cost"`
	UpdatedAt     time.Time `field:"updated_at"`
	JobTicket     string    `field:"job_ticket"`
	LocationName  string    `field:"location_name"`
	WarehouseName string    `field:"warehouse_name"`
	CustomerName  string    `field:"customer_name"`
	RemainingQty  int       `field:"remaining_quantity"`
}

type TransactionRep struct {
	StockID      string    `field:"stock_id"`
	MaterialType string    `field:"material_type"`
	Qty          int       `field:"quantity"`
	UnitCost     float64   `field:"unit_cost"`
	Cost         float64   `field:"cost"`
	UpdatedAt    time.Time `field:"updated_at"`
	TotalValue   float64   `field:"total_value"`
}

type SearchFilter struct {
	stockID      string
	customerID   int
	locationID   int
	materialType string
	dateFrom     string
	dateTo       string
	dateAsOf     string
}

type Reporter interface {
	getReportList() [][]string
	showReport()
}

type Report struct {
	db     *sql.DB
	app    fyne.App
	window fyne.Window
}

type InventoryReport struct {
	Report
	invFilter SearchFilter
}

type TransactionReport struct {
	Report
	trxFilter SearchFilter
}

type BalanceReport struct {
	Report
	blcFilter SearchFilter
}

var accLib accounting.Accounting = accounting.Accounting{Symbol: "$", Precision: 2}

func fetchLocations(db *sql.DB) ([]Location, error) {
	rows, err := db.Query("SELECT * FROM locations;")
	if err != nil {
		log.Println("Error fetchLocations1: ", err)
		return nil, err
	}
	defer rows.Close()

	var locations []Location

	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.id, &location.name, &location.warehouseID); err != nil {
			log.Println("Error fetchLocations2: ", err)
			return locations, err
		}
		locations = append(locations, location)
	}
	if err = rows.Err(); err != nil {
		return locations, err
	}

	return locations, nil
}

func getReport(r Reporter) {
	r.showReport()
}

func getReportTable(list [][]string) fyne.Widget {
	reportTable := widget.NewTable(
		func() (int, int) {
			return len(list), len(list[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Transactions")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(list[i.Row][i.Col])
		})

	for col := 0; col < len(list[0]); col++ {
		reportTable.SetColumnWidth(col, float32(150))
	}

	return reportTable
}

func downloadReport(window fyne.Window, list [][]string, fileName string) {
	file, err := os.Create("../reports/ " + fileName + ".csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.WriteAll(list)

	dialog := dialog.NewInformation("Success", "File has been saved to the reports folder", window)
	dialog.Show()
}

func (i InventoryReport) getReportList() [][]string {
	rows, err := i.db.Query(`SELECT m.material_id, m.stock_id, l.name, m.description,
							m.notes, m.quantity, m.min_required_quantity, m.max_required_quantity,
							m.updated_at, c.name, m.material_type,
								CASE
									WHEN m.is_active THEN 'Yes'
									ELSE 'No'
								END AS is_active,
							m.cost, m.owner
							FROM materials m
							LEFT JOIN locations l ON m.location_id = l.location_id
							LEFT JOIN customers c ON c.customer_id = m.customer_id
							WHERE 
								($1 = '' OR m.stock_id = $1) AND
								($2 = 0 OR c.customer_id = $2) AND
								($3 = 0 OR l.location_id = $3)
							ORDER BY m.updated_at ASC;`,
		i.invFilter.stockID, i.invFilter.customerID, i.invFilter.locationID)
	if err != nil {
		fmt.Printf("Error getMaterialsTable1: %e", err)
	}

	invList := [][]string{
		{
			"Material ID", "Stock ID", "Location", "Material Type",
			"Description", "Notes", "Quantity", "Min Qty",
			"Max Qty", "Updated At", "Customer", "Is Active", "Owner",
		},
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

		year, month, day := inv.UpdatedAt.Date()
		strDate := strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(day) + "/" +
			strconv.Itoa(year)

		invList = append(invList, []string{
			strconv.Itoa(inv.MaterialID),
			inv.StockID,
			inv.LocationName,
			inv.MaterialType,
			inv.Description,
			inv.Notes,
			strconv.Itoa(inv.Quantity),
			strconv.Itoa(inv.MinRequiredQty),
			strconv.Itoa(inv.MaxRequiredQty),
			strDate,
			inv.CustomerName,
			inv.IsActive,
			inv.Owner,
		})
	}

	return invList
}

func (i InventoryReport) showReport() {
	window := i.app.NewWindow("Inventory")

	customers, _ := fetchCustomers(i.db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	locations, _ := fetchLocations(i.db)
	var locationsStr []string
	locationsMap := make(map[string]int)
	for _, location := range locations {
		locationsStr = append(locationsStr, location.name)
		locationsMap[location.name] = location.id
	}

	stockIDInput := widget.NewEntry()
	customerSelector := widget.NewSelect(customersStr, func(s string) {})
	locationSelector := widget.NewSelect(locationsStr, func(s string) {})

	// Filter Inventory List by options
	dialog := dialog.NewForm("Filter Options", "Show", "",
		[]*widget.FormItem{
			widget.NewFormItem("Stock ID", stockIDInput),
			widget.NewFormItem("Customer", customerSelector),
			widget.NewFormItem("Location", locationSelector),
		}, func(confirm bool) {
			if confirm {
				i.invFilter = SearchFilter{
					stockID:    stockIDInput.Text,
					customerID: customersMap[customerSelector.Selected],
					locationID: locationsMap[locationSelector.Selected],
				}

				invList := i.getReportList()
				inventoryTable := getReportTable(invList)
				window.SetContent(inventoryTable)
				window.Resize(fyne.NewSize(1600, 700))
				window.Show()
			}
		}, i.window)

	dialog.Resize(fyne.NewSize(600, 200))
	dialog.Show()

}

func (t TransactionReport) getReportList() [][]string {
	rows, err := t.db.Query(`SELECT m.stock_id, m.material_type, tl.quantity_change as "quantity",
								tl.cost as "unit_cost",
								(tl.quantity_change * tl.cost) as "cost",
								tl.updated_at
							 FROM transactions_log tl
							 LEFT JOIN materials m ON m.material_id = tl.material_id
							 LEFT JOIN customers c ON m.customer_id = c.customer_id
							 WHERE 
								($1 = 0 OR m.customer_id = $1) AND
								($2 = '' OR m.material_type::TEXT = $2) AND
								tl.updated_at::TEXT >= $3 AND
								tl.updated_at::TEXT <= $4
							 ORDER BY transaction_id;`,
		t.trxFilter.customerID, t.trxFilter.materialType, t.trxFilter.dateFrom, t.trxFilter.dateTo)
	if err != nil {
		fmt.Printf("Error getTransactionsTable1: %e", err)
	}

	trxList := [][]string{
		{
			"Stock ID", "Material Type", "Quantity (+/-)", "Unit Price, USD", "Price, USD", "Accepted Date",
		},
	}

	for rows.Next() {
		trx := TransactionRep{}

		err := rows.Scan(
			&trx.StockID,
			&trx.MaterialType,
			&trx.Qty,
			&trx.UnitCost,
			&trx.Cost,
			&trx.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error getTransactionsTable2: %e", err)
		}

		year, month, day := trx.UpdatedAt.Date()
		strDate := strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(day) + "/" +
			strconv.Itoa(year)

		unitCost := accLib.FormatMoney(trx.UnitCost)
		cost := accLib.FormatMoney(trx.Cost)

		trxList = append(trxList, []string{
			trx.StockID,
			trx.MaterialType,
			strconv.Itoa(trx.Qty),
			unitCost,
			cost,
			strDate,
		})
	}

	return trxList
}

func (t TransactionReport) showReport() {
	customers, _ := fetchCustomers(t.db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	customerSelector := widget.NewSelect(customersStr, func(s string) {})
	typeSelector := widget.NewSelect([]string{"Carrier", "Card", "Envelope", "Insert", "Consumables"}, func(s string) {})
	dateFromEntry := widget.NewEntry()
	dateFromEntry.SetText(
		padStart(strconv.Itoa(int(time.Now().Month())), 2, '0') + "/" +
			"01" + "/" +
			strconv.Itoa(time.Now().Year()),
	)
	dateToEntry := widget.NewEntry()
	dateToEntry.SetText(
		padStart(strconv.Itoa(int(time.Now().Month())), 2, '0') + "/" +
			padStart(strconv.Itoa(time.Now().Day()), 2, '0') + "/" +
			strconv.Itoa(time.Now().Year()),
	)

	// Filter Inventory List by options
	dialog := dialog.NewForm("Filter Options", "Show", "",
		[]*widget.FormItem{
			widget.NewFormItem("Customer", customerSelector),
			widget.NewFormItem("Material Type", typeSelector),
			widget.NewFormItem("Date From (MM/DD/YYYY)", dateFromEntry),
			widget.NewFormItem("Date To (MM/DD/YYYY)", dateToEntry),
		}, func(confirm bool) {
			if confirm {
				parsedDateFrom := strings.Split(dateFromEntry.Text, "/")
				monthFrom := parsedDateFrom[0]
				dayFrom := parsedDateFrom[1]
				yearFrom := parsedDateFrom[2]

				parsedDateTo := strings.Split(dateToEntry.Text, "/")
				monthTo := parsedDateTo[0]
				dayTo := parsedDateTo[1]
				yearTo := parsedDateTo[2]

				t.trxFilter = SearchFilter{
					customerID:   customersMap[customerSelector.Selected],
					materialType: typeSelector.Selected,
					dateFrom:     yearFrom + "-" + monthFrom + "-" + dayFrom + " 00:00:00.000000",
					dateTo:       yearTo + "-" + monthTo + "-" + dayTo + " 23:59:59.999999",
				}

				window := t.app.NewWindow("Transactions")

				trxList := t.getReportList()
				transactionsTable := getReportTable(trxList)

				fileMenu := fyne.NewMenu("File", fyne.NewMenuItem("Save as .csv", func() {
					fileNameEntry := widget.NewEntry()
					dialog := dialog.NewForm("Save the File", "Save", "Cancel", []*widget.FormItem{
						widget.NewFormItem("File name", fileNameEntry),
					}, func(confirm bool) {
						if confirm {
							downloadReport(window, trxList, fileNameEntry.Text)
						}
					}, window)
					dialog.Resize(fyne.NewSize(400, 50))
					dialog.Show()
				}))

				window.SetMainMenu(fyne.NewMainMenu(fileMenu))
				window.SetContent(transactionsTable)
				window.Resize(fyne.NewSize(1000, 700))
				window.Show()

			}
		}, t.window)

	dialog.Resize(fyne.NewSize(600, 200))
	dialog.Show()
}

func (b BalanceReport) getReportList() [][]string {
	rows, err := b.db.Query(`
	SELECT m.stock_id,
		   m.material_type,
		   SUM(tl.quantity_change) AS "quantity",
		   SUM(tl.quantity_change * tl.cost) AS "total_value"
	FROM transactions_log tl
	LEFT JOIN materials m ON m.material_id = tl.material_id
	WHERE
		($1 = 0 OR m.customer_id = $1) AND
		($2 = '' OR m.material_type::TEXT = $2) AND
		tl.updated_at::TEXT <= $3
	GROUP BY m.stock_id, m.material_type
`,
		b.blcFilter.customerID, b.blcFilter.materialType, b.blcFilter.dateAsOf,
	)
	if err != nil {
		fmt.Printf("Error getBalanceTable1: %e", err)
	}

	blcList := [][]string{
		{
			"Stock ID", "Material Type", "Quantity", "Total Value, USD",
		},
	}

	for rows.Next() {
		balance := TransactionRep{}

		err := rows.Scan(&balance.StockID, &balance.MaterialType, &balance.Qty, &balance.TotalValue)

		if err != nil {
			log.Printf("Error getBalanceTable2: %e", err)
		}

		totalValue := accLib.FormatMoney(balance.TotalValue)

		blcList = append(blcList, []string{
			balance.StockID,
			balance.MaterialType,
			strconv.Itoa(balance.Qty),
			totalValue,
		})
	}

	return blcList

}

func (b BalanceReport) showReport() {
	customers, _ := fetchCustomers(b.db)
	var customersStr []string
	customersMap := make(map[string]int)
	for _, customer := range customers {
		customersStr = append(customersStr, customer.name)
		customersMap[customer.name] = customer.id
	}

	customerSelector := widget.NewSelect(customersStr, func(s string) {})
	typeSelector := widget.NewSelect([]string{"Carrier", "Card", "Envelope", "Insert", "Consumables"}, func(s string) {})
	dateAsOf := widget.NewEntry()
	dateAsOf.SetText(
		padStart(strconv.Itoa(int(time.Now().Month())), 2, 0) + "/" +
			padStart(strconv.Itoa(time.Now().Day()), 2, '0') + "/" +
			strconv.Itoa(time.Now().Year()),
	)

	// Filter Inventory List by options
	dialog := dialog.NewForm("Filter Options", "Show", "",
		[]*widget.FormItem{
			widget.NewFormItem("Customer", customerSelector),
			widget.NewFormItem("Material Type", typeSelector),
			widget.NewFormItem("Date As of (MM/DD/YYYY)", dateAsOf),
		}, func(confirm bool) {
			if confirm {
				parsedDate := strings.Split(dateAsOf.Text, "/")
				month := parsedDate[0]
				day := parsedDate[1]
				year := parsedDate[2]

				b.blcFilter = SearchFilter{
					customerID:   customersMap[customerSelector.Selected],
					materialType: typeSelector.Selected,
					dateAsOf:     year + "-" + month + "-" + day + " 23:59:59.999999",
				}

				window := b.app.NewWindow("Transactions Balance")
				blcList := b.getReportList()
				balanceTable := getReportTable(blcList)

				fileMenu := fyne.NewMenu("File", fyne.NewMenuItem("Save as .csv", func() {
					fileNameEntry := widget.NewEntry()
					dialog := dialog.NewForm("Save the File", "Save", "Cancel", []*widget.FormItem{
						widget.NewFormItem("File name", fileNameEntry),
					}, func(confirm bool) {
						if confirm {
							downloadReport(window, blcList, fileNameEntry.Text)
						}
					}, window)
					dialog.Resize(fyne.NewSize(400, 50))
					dialog.Show()
				}))

				window.SetMainMenu(fyne.NewMainMenu(fileMenu))
				window.SetContent(balanceTable)
				window.Resize(fyne.NewSize(650, 400))
				window.Show()
			}
		}, b.window)

	dialog.Resize(fyne.NewSize(500, 200))
	dialog.Show()
}
