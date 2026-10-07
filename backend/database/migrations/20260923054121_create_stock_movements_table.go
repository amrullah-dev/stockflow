package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054121CreateStockMovementsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054121CreateStockMovementsTable) Signature() string {
	return "20260923054121_create_stock_movements_table"
}

// Up Run the migrations.
func (r *M20260923054121CreateStockMovementsTable) Up() error {
	if !facades.Schema().HasTable("stock_movements") {
		return facades.Schema().Create("stock_movements", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("warehouse_id").Unsigned()
			table.BigInteger("product_id").Unsigned()
			table.String("type", 30)
			table.Decimal("quantity").Default(0)
			table.Decimal("balance_before").Default(0)
			table.Decimal("balance_after").Default(0)
			table.String("reference_type", 50).Nullable()
			table.BigInteger("reference_id").Unsigned().Nullable()
			table.String("note", 255).Nullable()
			table.BigInteger("created_by").Unsigned().Nullable()
			table.TimestampsTz()

			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
			table.Foreign("created_by").References("id").On("users").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054121CreateStockMovementsTable) Down() error {
	return facades.Schema().DropIfExists("stock_movements")
}
