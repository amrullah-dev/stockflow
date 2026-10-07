package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260923051200CreateRolesTable{},
		&migrations.M20260923052149CreateUsersTable{},
		&migrations.M20260923053035CreateCategoriesTable{},
		&migrations.M20260923053608CreateProductsTable{},
		&migrations.M20260923054057CreateSuppliersTable{},
		&migrations.M20260923054059CreateWarehousesTable{},
		&migrations.M20260923054102CreateWarehouseStocksTable{},
		&migrations.M20260923054105CreateStockInsTable{},
		&migrations.M20260923054108CreateStockInItemsTable{},
		&migrations.M20260923054110CreateStockOutsTable{},
		&migrations.M20260923054113CreateStockOutItemsTable{},
		&migrations.M20260923054116CreateWarehouseTransfersTable{},
		&migrations.M20260923054119CreateWarehouseTransferItemsTable{},
		&migrations.M20260923054121CreateStockMovementsTable{},
		&migrations.M20260923054124CreateAuditLogsTable{},
		&migrations.M20260923060000AlterInventoryQuantitiesToBigintTable{},
		&migrations.M20261007120000AlterStockInItemMoneyPrecision{},
	}
}
