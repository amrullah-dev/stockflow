package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054108CreateStockInItemsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054108CreateStockInItemsTable) Signature() string {
	return "20260923054108_create_stock_in_items_table"
}

// Up Run the migrations.
func (r *M20260923054108CreateStockInItemsTable) Up() error {
	if !facades.Schema().HasTable("stock_in_items") {
		return facades.Schema().Create("stock_in_items", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("stock_in_id").Unsigned()
			table.BigInteger("product_id").Unsigned()
			table.Decimal("quantity").Default(0)
			table.Decimal("unit_price").Default(0)
			table.Decimal("subtotal").Default(0)
			table.TimestampsTz()

			table.Foreign("stock_in_id").References("id").On("stock_ins").RestrictOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054108CreateStockInItemsTable) Down() error {
	return facades.Schema().DropIfExists("stock_in_items")
}
