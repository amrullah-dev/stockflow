package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260923060000AlterInventoryQuantitiesToBigintTable struct{}

// Signature The unique signature for the migration.
func (r *M20260923060000AlterInventoryQuantitiesToBigintTable) Signature() string {
	return "20260923060000_alter_inventory_quantities_to_bigint_table"
}

// Up changes inventory quantities and balances from DECIMAL to BIGINT.
func (r *M20260923060000AlterInventoryQuantitiesToBigintTable) Up() error {
	if err := facades.Schema().Table("warehouse_stocks", func(table schema.Blueprint) {
		table.BigInteger("quantity").Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_in_items", func(table schema.Blueprint) {
		table.BigInteger("quantity").Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_out_items", func(table schema.Blueprint) {
		table.BigInteger("quantity").Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("warehouse_transfer_items", func(table schema.Blueprint) {
		table.BigInteger("quantity").Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_movements", func(table schema.Blueprint) {
		table.BigInteger("quantity").Default(0).Change()
		table.BigInteger("balance_before").Default(0).Change()
		table.BigInteger("balance_after").Default(0).Change()
	}); err != nil {
		return err
	}

	return facades.Schema().Table("products", func(table schema.Blueprint) {
		table.BigInteger("minimum_stock").Default(0).Change()
	})
}

// Down restores the inventory columns to their verified previous DECIMAL(8,2) definition.
func (r *M20260923060000AlterInventoryQuantitiesToBigintTable) Down() error {
	if err := facades.Schema().Table("warehouse_stocks", func(table schema.Blueprint) {
		table.Decimal("quantity").Total(8).Places(2).Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_in_items", func(table schema.Blueprint) {
		table.Decimal("quantity").Total(8).Places(2).Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_out_items", func(table schema.Blueprint) {
		table.Decimal("quantity").Total(8).Places(2).Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("warehouse_transfer_items", func(table schema.Blueprint) {
		table.Decimal("quantity").Total(8).Places(2).Default(0).Change()
	}); err != nil {
		return err
	}

	if err := facades.Schema().Table("stock_movements", func(table schema.Blueprint) {
		table.Decimal("quantity").Total(8).Places(2).Default(0).Change()
		table.Decimal("balance_before").Total(8).Places(2).Default(0).Change()
		table.Decimal("balance_after").Total(8).Places(2).Default(0).Change()
	}); err != nil {
		return err
	}

	return facades.Schema().Table("products", func(table schema.Blueprint) {
		table.Decimal("minimum_stock").Total(8).Places(2).Default(0).Change()
	})
}
