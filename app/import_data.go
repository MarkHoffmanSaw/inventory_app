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
	`)

	var warehouseId int
	db.QueryRow(`
			INSERT INTO warehouses(name) VALUES($1) RETURNING warehouse_id`,
		"Tag").
		Scan(&warehouseId)

	for id, record := range records {
		customerName := record[0]
		customerCode := record[1]
		locationName := record[2]
		stockID := record[3]
		materialType := record[4]
		description := record[5]
		notes := record[6]
		qty, _ := strconv.Atoi(record[7])
		minQty, _ := strconv.Atoi(record[8])
		maxQty, _ := strconv.Atoi(record[9])
		isActive, _ := strconv.ParseBool(record[10])
		owner := record[11]
		unitCost, _ := strconv.ParseFloat(record[12], 64)

		var customerId int
		db.QueryRow(`
			INSERT INTO customers(name,customer_code) VALUES($1,$2) RETURNING customer_id`,
			customerName, customerCode).
			Scan(&customerId)

		var locationId int
		db.QueryRow(`
			INSERT INTO locations(name,warehouse_id) VALUES($1,$2) RETURNING location_id`,
			locationName, warehouseId).
			Scan(&locationId)

		var materialId int
		db.QueryRow(`
			INSERT INTO materials(
								stock_id,location_id,customer_id,material_type,
								description,notes,quantity,min_required_quantity,
								max_required_quantity,updated_at,is_active,cost,owner
								  )
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10,$11,$12)
			RETURNING material_id`,
			stockID, locationId, customerId, materialType, description, notes, qty, minQty, maxQty, isActive, unitCost, owner).
			Scan(&materialId)

		_, err := db.Query(`
			INSERT INTO transactions_log(
									 material_id,stock_id,quantity_change,
									 notes,cost,job_ticket,updated_at,remaining_quantity
									 	)
			VALUES($1,$2,$3,$4,$5,$6,NOW(),$7)`,
			materialId, stockID, qty, notes, unitCost, "job_ticket", qty,
		)
		log.Println(err)

		log.Println("job done", id)
	}
}
