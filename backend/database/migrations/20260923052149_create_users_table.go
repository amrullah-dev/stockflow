package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923052149CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923052149CreateUsersTable) Signature() string {
	return "20260923052149_create_users_table"
}

// Up Run the migrations.
func (r *M20260923052149CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("role_id").Unsigned()
			table.String("name", 100)
			table.String("email", 150)
			table.String("password", 255)
			table.Boolean("is_active").Default(true)
			table.Timestamp("last_login_at").Nullable()
			table.TimestampsTz()
			table.SoftDeletes()

			table.Unique("email")
			table.Foreign("role_id").References("id").On("roles").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923052149CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
