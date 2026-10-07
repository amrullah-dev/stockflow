package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054059CreateWarehousesTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054059CreateWarehousesTable) Signature() string {
	return "20260923054059_create_warehouses_table"
}

// Up Run the migrations.
func (r *M20260923054059CreateWarehousesTable) Up() error {
	if !facades.Schema().HasTable("warehouses") {
		return facades.Schema().Create("warehouses", func(table schema.Blueprint) {
			table.ID()
			table.String("code", 50)
			table.String("name", 150)
			table.String("location", 255).Nullable()
			table.String("description", 255).Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.SoftDeletes()

			table.Unique("code")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054059CreateWarehousesTable) Down() error {
	return facades.Schema().DropIfExists("warehouses")
}
