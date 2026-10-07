package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054113CreateStockOutItemsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054113CreateStockOutItemsTable) Signature() string {
	return "20260923054113_create_stock_out_items_table"
}

// Up Run the migrations.
func (r *M20260923054113CreateStockOutItemsTable) Up() error {
	if !facades.Schema().HasTable("stock_out_items") {
		return facades.Schema().Create("stock_out_items", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("stock_out_id").Unsigned()
			table.BigInteger("product_id").Unsigned()
			table.Decimal("quantity").Default(0)
			table.TimestampsTz()

			table.Foreign("stock_out_id").References("id").On("stock_outs").RestrictOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054113CreateStockOutItemsTable) Down() error {
	return facades.Schema().DropIfExists("stock_out_items")
}
