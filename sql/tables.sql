CREATE TYPE customer_type AS ENUM ('Internal', 'External');

CREATE TABLE customers (
	customer_id serial PRIMARY KEY,
	customer_name VARCHAR(55),
	customer_code VARCHAR(55),
	customer_type CUSTOMER_TYPE
);

CREATE TABLE warehouses (
	warehouse_id serial PRIMARY KEY,
	name VARCHAR(55)
);

CREATE TABLE locations (
	location_id serial PRIMARY KEY,
	name VARCHAR(55),
	warehouse_id int REFERENCES warehouses(warehouse_id)
);

CREATE TYPE material_type AS ENUM ('Carrier','Card','Envelope','Insert', 'Consumables');

CREATE TABLE materials (
	material_id serial,
	stock_id VARCHAR(55),
	location_id int REFERENCES locations(location_id),
	customer_id int REFERENCES customers(customer_id),
	material_type MATERIAL_TYPE,
	description TEXT,
	notes TEXT,
	quantity int,
	updated_at TIMESTAMP,
	CONSTRAINT pk_material_id_stock_id PRIMARY KEY (material_id, stock_id)
);

CREATE TABLE transactions_log (
	transaction_id serial PRIMARY KEY,
	material_id int,
	stock_id VARCHAR(55),
	quantity_change int,
	notes text,
	cost int,
	updated_at timestamp,
	FOREIGN KEY (material_id, stock_id) REFERENCES materials(material_id, stock_id)
);