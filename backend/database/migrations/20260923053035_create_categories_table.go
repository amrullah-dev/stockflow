package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"goravel/app/facades"
)

type M20260923053035CreateCategoriesTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923053035CreateCategoriesTable) Signature() string {
	return "20260923053035_create_categories_table"
}

// Up Run the migrations.
func (r *M20260923053035CreateCategoriesTable) Up() error {
	if !facades.Schema().HasTable("categories") {
		return facades.Schema().Create("categories", func(table schema.Blueprint) {
			table.ID()
			table.String("name", 100)
			table.String("description", 255).Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.SoftDeletes()

			table.Unique("name")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923053035CreateCategoriesTable) Down() error {
	return facades.Schema().DropIfExists("categories")
}
