package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923054124CreateAuditLogsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923054124CreateAuditLogsTable) Signature() string {
	return "20260923054124_create_audit_logs_table"
}

// Up Run the migrations.
func (r *M20260923054124CreateAuditLogsTable) Up() error {
	if !facades.Schema().HasTable("audit_logs") {
		return facades.Schema().Create("audit_logs", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("user_id").Unsigned().Nullable()
			table.String("action", 50)
			table.String("subject_type", 50)
			table.BigInteger("subject_id").Unsigned()
			table.Text("old_values").Nullable()
			table.Text("new_values").Nullable()
			table.String("ip_address", 45).Nullable()
			table.String("user_agent", 255).Nullable()
			table.TimestampsTz()

			table.Index("subject_type", "subject_id")
			table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260923054124CreateAuditLogsTable) Down() error {
	return facades.Schema().DropIfExists("audit_logs")
}
