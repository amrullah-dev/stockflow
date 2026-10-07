package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261007120000AlterStockInItemMoneyPrecision struct{}

func (r *M20261007120000AlterStockInItemMoneyPrecision) Signature() string {
	return "20261007120000_alter_stock_in_item_money_precision"
}

func (r *M20261007120000AlterStockInItemMoneyPrecision) Up() error {
	return facades.Schema().Table("stock_in_items", func(table schema.Blueprint) {
		table.Decimal("unit_price").Total(18).Places(2).Change()
		table.Decimal("subtotal").Total(18).Places(2).Change()
	})
}

func (r *M20261007120000AlterStockInItemMoneyPrecision) Down() error {
	return facades.Schema().Table("stock_in_items", func(table schema.Blueprint) {
		table.Decimal("unit_price").Total(8).Places(2).Change()
		table.Decimal("subtotal").Total(8).Places(2).Change()
	})
}
