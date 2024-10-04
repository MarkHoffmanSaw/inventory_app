CREATE DATABASE tag_db;

CREATE TYPE customer_type AS ENUM ('TAG Owned', 'Customer Owned');

CREATE TABLE customers (
	customer_id serial PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	customer_code VARCHAR(100),
	customer_type CUSTOMER_TYPE  NOT NULL
);

CREATE TABLE warehouses (
	warehouse_id serial PRIMARY KEY,
	name VARCHAR(100) UNIQUE  NOT NULL
);

CREATE TABLE locations (
	location_id serial PRIMARY KEY,
	name VARCHAR(100)  NOT NULL,
	warehouse_id int REFERENCES warehouses(warehouse_id),
	CONSTRAINT locations_name_warehouse_id UNIQUE(name, warehouse_id)
);

CREATE TYPE material_type AS ENUM ('Carrier','Card','Envelope','Insert', 'Consumables');

CREATE TABLE materials (
	material_id serial,
	stock_id VARCHAR(100)  NOT NULL,
	location_id int REFERENCES locations(location_id),
	customer_id int REFERENCES customers(customer_id),
	material_type MATERIAL_TYPE  NOT NULL,
	description TEXT,
	notes TEXT,
	quantity int  NOT NULL,
	min_required_quantity int,
	max_required_quantity int,
	updated_at TIMESTAMP,
	CONSTRAINT pk_stock_id_location_id PRIMARY KEY (stock_id, location_id)
);

CREATE TABLE transactions_log (
	transaction_id serial PRIMARY KEY,
	material_id int NOT NULL,
	stock_id VARCHAR(100) NOT NULL,
	quantity_change int NOT NULL,
	notes text,
	cost decimal,
	job_ticket VARCHAR(100),
	updated_at timestamp
);