package feature

import (
	"context"
	"errors"
	"fmt"
	"testing"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	frameworkerrors "github.com/goravel/framework/errors"
	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	"goravel/tests"
)

func phase3Token(t *testing.T, userID uint) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	return token
}

func TestDraftUpdateDeleteRequiresAuthentication(t *testing.T) {
	for _, path := range []string{"/api/stock-ins/1", "/api/stock-outs/1", "/api/warehouse-transfers/1"} {
		t.Run(path, func(t *testing.T) {
			response, err := (&tests.TestCase{}).Http(t).Put(path, nil)
			require.NoError(t, err)
			response.AssertUnauthorized()
			response, err = (&tests.TestCase{}).Http(t).Delete(path, nil)
			require.NoError(t, err)
			response.AssertUnauthorized()
		})
	}
}

func TestDraftUpdateDeleteNotFoundAndInvalidInput(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	for _, resource := range []string{"stock-ins", "stock-outs", "warehouse-transfers"} {
		t.Run(resource+" not found", func(t *testing.T) {
			path := "/api/" + resource + "/999999999"
			response, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(path, jsonReader(map[string]any{}))
			require.NoError(t, err)
			response.AssertNotFound()
			response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(path, nil)
			require.NoError(t, err)
			response.AssertNotFound()
		})
	}

	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 1})
	response, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/stock-ins/%d", stockIn.ID), jsonReader(map[string]any{
		"supplier_id": f.supplier.ID, "warehouse_id": f.warehouse1.ID, "transaction_date": "07-10-2026", "items": []map[string]any{{"product_id": f.product1.ID, "quantity": 1, "unit_price": 1}},
	}))
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestStockInDraftUpdateAndDelete(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 3, UnitPrice: 10, Subtotal: 30})
	ref := stockIn.ReferenceNumber

	response, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/stock-ins/%d", stockIn.ID), jsonReader(map[string]any{
		"supplier_id": f.supplier.ID, "warehouse_id": f.warehouse1.ID, "transaction_date": "2026-10-07", "note": "updated", "items": []map[string]any{{"product_id": f.product2.ID, "quantity": 7, "unit_price": 25, "subtotal": 999}},
	}))
	require.NoError(t, err)
	response.AssertOk()
	updated, err := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, ref, updated.ReferenceNumber)
	require.Equal(t, "updated", updated.Note)
	items, err := repositories.NewStockInRepository(f.db.Query()).FindItems(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(f.product2.ID), items[0].ProductID)
	require.Equal(t, int64(7), items[0].Quantity)
	require.Equal(t, float64(175), items[0].Subtotal)

	deleted := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 1})
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/stock-ins/%d", deleted.ID), nil)
	require.NoError(t, err)
	response.AssertOk()
	_, err = repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(deleted.ID))
	require.ErrorIs(t, err, frameworkerrors.OrmRecordNotFound)
}

func TestStockOutDraftUpdateDeleteAndPostedProtection(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 2})
	response, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/stock-outs/%d", stockOut.ID), jsonReader(map[string]any{
		"warehouse_id": f.warehouse1.ID, "transaction_date": "2026-10-07", "destination": "updated", "items": []map[string]any{{"product_id": f.product2.ID, "quantity": 4}},
	}))
	require.NoError(t, err)
	response.AssertOk()
	items, err := repositories.NewStockOutRepository(f.db.Query()).FindItems(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(f.product2.ID), items[0].ProductID)

	deleted := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 1})
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/stock-outs/%d", deleted.ID), nil)
	require.NoError(t, err)
	response.AssertOk()

	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	posted := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 1})
	require.NoError(t, services.NewStockOutService(f.db).Post(context.Background(), uint64(posted.ID), uint64(f.user.ID)))
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/stock-outs/%d", posted.ID), jsonReader(map[string]any{"warehouse_id": f.warehouse1.ID, "transaction_date": "2026-10-07", "items": []map[string]any{{"product_id": f.product1.ID, "quantity": 1}}}))
	require.NoError(t, err)
	response.AssertStatus(409)
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/stock-outs/%d", posted.ID), nil)
	require.NoError(t, err)
	response.AssertStatus(409)
}

func TestWarehouseTransferDraftUpdateDeleteAndValidation(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 2})
	response, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/warehouse-transfers/%d", transfer.ID), jsonReader(map[string]any{
		"from_warehouse_id": f.warehouse1.ID, "to_warehouse_id": f.warehouse2.ID, "transaction_date": "2026-10-07", "note": "updated", "items": []map[string]any{{"product_id": f.product2.ID, "quantity": 5}},
	}))
	require.NoError(t, err)
	response.AssertOk()
	items, err := repositories.NewWarehouseTransferRepository(f.db.Query()).FindItems(uint64(transfer.ID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(f.product2.ID), items[0].ProductID)

	bad, err := (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/warehouse-transfers/%d", transfer.ID), jsonReader(map[string]any{
		"from_warehouse_id": f.warehouse1.ID, "to_warehouse_id": f.warehouse1.ID, "transaction_date": "2026-10-07", "items": []map[string]any{{"product_id": f.product1.ID, "quantity": 1}},
	}))
	require.NoError(t, err)
	bad.AssertStatus(422)

	deleted := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 1})
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/warehouse-transfers/%d", deleted.ID), nil)
	require.NoError(t, err)
	response.AssertOk()
}

func TestDraftUpdateRollsBackAfterReplacementItemFailure(t *testing.T) {
	f := newInventoryFixture(t)
	fail := func(contractsorm.Query) error { return errors.New("simulated replacement item failure") }

	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 2, UnitPrice: 10})
	stockInService := services.NewStockInService(f.db)
	stockInService.SetOnItemCreatedHook(fail)
	_, err := stockInService.UpdateDraft(context.Background(), uint64(stockIn.ID), services.CreateStockInDraftInput{SupplierID: uint64(f.supplier.ID), WarehouseID: uint64(f.warehouse1.ID), TransactionDate: "2026-10-07", Note: "failed", Items: []services.CreateStockInItemInput{{ProductID: uint64(f.product2.ID), Quantity: 5, UnitPrice: 20}}})
	require.Error(t, err)
	stockIn, err = repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, "", stockIn.Note)
	items, err := repositories.NewStockInRepository(f.db.Query()).FindItems(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(f.product1.ID), items[0].ProductID)

	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 2})
	stockOutService := services.NewStockOutService(f.db)
	stockOutService.SetOnItemCreatedHook(fail)
	_, err = stockOutService.UpdateDraft(context.Background(), uint64(stockOut.ID), services.CreateStockOutDraftInput{WarehouseID: uint64(f.warehouse1.ID), TransactionDate: "2026-10-07", Destination: "failed", Items: []services.CreateStockOutItemInput{{ProductID: uint64(f.product2.ID), Quantity: 5}}})
	require.Error(t, err)
	stockOutItems, err := repositories.NewStockOutRepository(f.db.Query()).FindItems(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Len(t, stockOutItems, 1)
	require.Equal(t, uint64(f.product1.ID), stockOutItems[0].ProductID)

	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 2})
	transferService := services.NewWarehouseTransferService(f.db)
	transferService.SetOnItemCreatedHook(fail)
	_, err = transferService.UpdateDraft(context.Background(), uint64(transfer.ID), services.CreateWarehouseTransferDraftInput{FromWarehouseID: uint64(f.warehouse1.ID), ToWarehouseID: uint64(f.warehouse2.ID), TransactionDate: "2026-10-07", Note: "failed", Items: []services.CreateWarehouseTransferItemInput{{ProductID: uint64(f.product2.ID), Quantity: 5}}})
	require.Error(t, err)
	transferItems, err := repositories.NewWarehouseTransferRepository(f.db.Query()).FindItems(uint64(transfer.ID))
	require.NoError(t, err)
	require.Len(t, transferItems, 1)
	require.Equal(t, uint64(f.product1.ID), transferItems[0].ProductID)
}
