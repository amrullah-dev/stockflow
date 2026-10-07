package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054119CreateWarehouseTransferItemsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054119CreateWarehouseTransferItemsTable) Signature() string {
	return "20260923054119_create_warehouse_transfer_items_table"
}

// Up Run the migrations.
func (r *M20260923054119CreateWarehouseTransferItemsTable) Up() error {
	if !facades.Schema().HasTable("warehouse_transfer_items") {
		return facades.Schema().Create("warehouse_transfer_items", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("transfer_id").Unsigned()
			table.BigInteger("product_id").Unsigned()
			table.Decimal("quantity").Default(0)
			table.TimestampsTz()

			table.Foreign("transfer_id").References("id").On("warehouse_transfers").RestrictOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054119CreateWarehouseTransferItemsTable) Down() error {
	return facades.Schema().DropIfExists("warehouse_transfer_items")
}
