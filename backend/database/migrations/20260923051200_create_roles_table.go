package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"goravel/app/facades"
)

type M20260923051200CreateRolesTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923051200CreateRolesTable) Signature() string {
	return "20260923051200_create_roles_table"
}

// Up Run the migrations.
func (r *M20260923051200CreateRolesTable) Up() error {
	if !facades.Schema().HasTable("roles") {
		return facades.Schema().Create("roles", func(table schema.Blueprint) {
			table.ID()
			table.String("name", 50)
			table.String("description", 255).Nullable()
			table.Unique("name")
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923051200CreateRolesTable) Down() error {
	return facades.Schema().DropIfExists("roles")
}
