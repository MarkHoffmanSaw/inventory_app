package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

func importToDB(db *sql.DB) {
	file, err := os.Open("./import_data.csv")
	if err != nil {
		log.Fatal("Error while opening the file", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	db.Query(`
		DELETE FROM transactions_log;
		DELETE FROM materials;
		DELETE FROM locations;
		DELETE FROM customers;
		DELETE FROM warehouses;
	`)

	for _, record := range records {
		customerName := record[0]
		customerCode := record[1]
		warehouseName := record[2]
		locationName := record[3]
		stockID := record[4]
		materialType := record[5]
		description := record[6]
		notes := record[7]
		qty, _ := strconv.Atoi(record[8])
		minQty, _ := strconv.Atoi(record[9])
		maxQty, _ := strconv.Atoi(record[10])
		isActive, _ := strconv.ParseBool(record[11])
		owner := record[12]
		unitCost, _ := strconv.ParseFloat(record[13], 64)

		// Check for a customer
		var customerId int
		db.QueryRow(`SELECT customer_id FROM customers
						WHERE name = $1
						AND customer_code = $2`,
			customerName, customerCode).
			Scan(&customerId)

		if customerId == 0 {
			db.QueryRow(`
			INSERT INTO customers(name,customer_code) VALUES($1,$2) RETURNING customer_id`,
				customerName, customerCode).
				Scan(&customerId)
		}

		// Check for a warehouse
		var warehouseId int
		db.QueryRow(`SELECT warehouse_id FROM warehouses
						WHERE name = $1
						`,
			warehouseName).
			Scan(&warehouseId)

		if warehouseId == 0 {
			db.QueryRow(`
					INSERT INTO warehouses(name) VALUES($1) RETURNING warehouse_id`,
				warehouseName).
				Scan(&warehouseId)
		}

		// Check for a location
		var locationId int
		db.QueryRow(`SELECT location_id FROM locations
						WHERE location_name = $1
						AND warehouse_id = $2
						`,
			locationName, warehouseId).
			Scan(&locationId)

		if locationId == 0 {
			db.QueryRow(`
			INSERT INTO locations(name,warehouse_id) VALUES($1,$2) RETURNING location_id`,
				locationName, warehouseId).
				Scan(&locationId)
		}

		var materialId int
		db.QueryRow(`
			INSERT INTO materials(
					stock_id,location_id,customer_id,material_type,
					description,notes,quantity,min_required_quantity,
					max_required_quantity,updated_at,is_active,cost,owner)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10,$11,$12)
			RETURNING material_id`,
			stockID, locationId, customerId, materialType,
			description, notes, qty, minQty, maxQty, isActive,
			unitCost, owner).
			Scan(&materialId)

		db.Query(`
			INSERT INTO transactions_log(
									 material_id,stock_id,quantity_change,
									 notes,cost,job_ticket,updated_at,remaining_quantity
									 	)
			VALUES($1,$2,$3,$4,$5,$6,NOW(),$7)`,
			materialId, stockID, qty, notes, unitCost, "job_ticket", qty,
		)

		log.Println("job done for material id", materialId)
	}
}
