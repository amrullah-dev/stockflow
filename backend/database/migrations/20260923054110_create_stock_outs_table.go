package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054110CreateStockOutsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054110CreateStockOutsTable) Signature() string {
	return "20260923054110_create_stock_outs_table"
}

// Up Run the migrations.
func (r *M20260923054110CreateStockOutsTable) Up() error {
	if !facades.Schema().HasTable("stock_outs") {
		return facades.Schema().Create("stock_outs", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number", 50)
			table.BigInteger("warehouse_id").Unsigned()
			table.Date("transaction_date")
			table.String("status", 20).Default("DRAFT")
			table.String("destination", 150).Nullable()
			table.String("note", 255).Nullable()
			table.BigInteger("created_by").Unsigned()
			table.Timestamp("posted_at").Nullable()
			table.TimestampsTz()

			table.Unique("reference_number")
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("created_by").References("id").On("users").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054110CreateStockOutsTable) Down() error {
	return facades.Schema().DropIfExists("stock_outs")
}
