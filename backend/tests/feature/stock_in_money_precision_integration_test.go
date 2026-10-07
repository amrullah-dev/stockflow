package feature

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/repositories"
	"goravel/app/services"
)

func TestStockInMoneyPrecisionRoundTrip(t *testing.T) {
	f := newInventoryFixture(t)
	service := services.NewStockInService(f.db)

	cases := []struct {
		name      string
		quantity  int64
		unitPrice float64
		subtotal  float64
	}{
		{name: "large whole amount", quantity: 10, unitPrice: 100000, subtotal: 1000000},
		{name: "ordinary amount", quantity: 3, unitPrice: 12500, subtotal: 37500},
		{name: "two decimal amount", quantity: 1, unitPrice: 999999.99, subtotal: 999999.99},
		{name: "zero amount", quantity: 10, unitPrice: 0, subtotal: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			created, err := service.CreateDraft(t.Context(), services.CreateStockInDraftInput{
				SupplierID:      uint64(f.supplier.ID),
				WarehouseID:     uint64(f.warehouse1.ID),
				TransactionDate: "2026-10-07",
				Items: []services.CreateStockInItemInput{
					{ProductID: uint64(f.product1.ID), Quantity: tc.quantity, UnitPrice: tc.unitPrice},
				},
			}, uint64(f.user.ID))
			require.NoError(t, err)

			stockIn, err := repositories.NewStockInRepository(f.db.Query()).FindByIDForRead(created.ID)
			require.NoError(t, err)
			require.Len(t, stockIn.Items, 1)
			require.InDelta(t, tc.unitPrice, stockIn.Items[0].UnitPrice, 0.000001)
			require.InDelta(t, tc.subtotal, stockIn.Items[0].Subtotal, 0.000001)
		})
	}
}
