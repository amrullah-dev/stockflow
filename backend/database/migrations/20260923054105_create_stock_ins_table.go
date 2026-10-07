package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054105CreateStockInsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054105CreateStockInsTable) Signature() string {
	return "20260923054105_create_stock_ins_table"
}

// Up Run the migrations.
func (r *M20260923054105CreateStockInsTable) Up() error {
	if !facades.Schema().HasTable("stock_ins") {
		return facades.Schema().Create("stock_ins", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number", 50)
			table.BigInteger("supplier_id").Unsigned()
			table.BigInteger("warehouse_id").Unsigned()
			table.Date("transaction_date")
			table.String("status", 20).Default("DRAFT")
			table.String("note", 255).Nullable()
			table.BigInteger("created_by").Unsigned()
			table.Timestamp("posted_at").Nullable()
			table.TimestampsTz()

			table.Unique("reference_number")
			table.Foreign("supplier_id").References("id").On("suppliers").RestrictOnDelete()
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("created_by").References("id").On("users").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054105CreateStockInsTable) Down() error {
	return facades.Schema().DropIfExists("stock_ins")
}
