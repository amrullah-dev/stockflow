package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054057CreateSuppliersTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054057CreateSuppliersTable) Signature() string {
	return "20260923054057_create_suppliers_table"
}

// Up Run the migrations.
func (r *M20260923054057CreateSuppliersTable) Up() error {
	if !facades.Schema().HasTable("suppliers") {
		return facades.Schema().Create("suppliers", func(table schema.Blueprint) {
			table.ID()
			table.String("code", 50)
			table.String("name", 150)
			table.String("phone", 30).Nullable()
			table.String("email", 150).Nullable()
			table.String("address", 255).Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.SoftDeletes()

			table.Unique("code")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054057CreateSuppliersTable) Down() error {
	return facades.Schema().DropIfExists("suppliers")
}
