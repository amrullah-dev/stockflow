package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"goravel/app/facades"
)

type M20260923053608CreateProductsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923053608CreateProductsTable) Signature() string {
	return "20260923053608_create_products_table"
}

// Up Run the migrations.
func (r *M20260923053608CreateProductsTable) Up() error {
	if !facades.Schema().HasTable("products") {
		return facades.Schema().Create("products", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("category_id").Unsigned()
			table.String("code", 50)
			table.String("name", 150)
			table.String("description", 255).Nullable()
			table.String("unit", 30)
			table.Decimal("minimum_stock").Default(0)
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.SoftDeletes()

			table.Unique("code")
			table.Foreign("category_id").References("id").On("categories").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923053608CreateProductsTable) Down() error {
	return facades.Schema().DropIfExists("products")
}
