package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054102CreateWarehouseStocksTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054102CreateWarehouseStocksTable) Signature() string {
	return "20260923054102_create_warehouse_stocks_table"
}

// Up Run the migrations.
func (r *M20260923054102CreateWarehouseStocksTable) Up() error {
	if !facades.Schema().HasTable("warehouse_stocks") {
		return facades.Schema().Create("warehouse_stocks", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("warehouse_id").Unsigned()
			table.BigInteger("product_id").Unsigned()
			table.Decimal("quantity").Default(0)
			table.TimestampsTz()

			table.Unique("warehouse_id", "product_id")
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054102CreateWarehouseStocksTable) Down() error {
	return facades.Schema().DropIfExists("warehouse_stocks")
}
