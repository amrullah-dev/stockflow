package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054116CreateWarehouseTransfersTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054116CreateWarehouseTransfersTable) Signature() string {
	return "20260923054116_create_warehouse_transfers_table"
}

// Up Run the migrations.
func (r *M20260923054116CreateWarehouseTransfersTable) Up() error {
	if !facades.Schema().HasTable("warehouse_transfers") {
		return facades.Schema().Create("warehouse_transfers", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number", 50)
			table.BigInteger("from_warehouse_id").Unsigned()
			table.BigInteger("to_warehouse_id").Unsigned()
			table.Date("transaction_date")
			table.String("status", 20).Default("DRAFT")
			table.String("note", 255).Nullable()
			table.BigInteger("created_by").Unsigned()
			table.Timestamp("posted_at").Nullable()
			table.TimestampsTz()

			table.Unique("reference_number")
			table.Foreign("from_warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("to_warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("created_by").References("id").On("users").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054116CreateWarehouseTransfersTable) Down() error {
	return facades.Schema().DropIfExists("warehouse_transfers")
}
