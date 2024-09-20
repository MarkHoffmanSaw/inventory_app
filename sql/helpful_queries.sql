-- look up by locations
select stock_id,
		sum(quantity),
		l.name as "LOCATION",
		w.name as "WAREHOUSE",
		c.name as "CUSTOMER" 
from materials m
left join locations l on l.location_id = m.location_id
left join warehouses w on w.warehouse_id = l.warehouse_id
left join customers c on c.customer_id = m.customer_id
group by stock_id, l.name, w.name, c.name

