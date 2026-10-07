package feature

import (
	"testing"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	"goravel/tests"
)

func warehouseTransferAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestWarehouseTransferPostRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Post("/api/warehouse-transfers/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferPostRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Post("/api/warehouse-transfers/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferPostRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Post("/api/warehouse-transfers/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferPostRejectsInvalidRouteID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/0/post", nil)
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestWarehouseTransferPostAuthenticatedNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/999999999/post", nil)
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestWarehouseTransferPostAuthenticatedAlreadyPosted(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 2)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})
	require.NoError(t, services.NewWarehouseTransferService(f.db).Post(t.Context(), uint64(transfer.ID), uint64(f.user.ID)))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/"+formatUint64(uint64(transfer.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(409)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 2)
}

func TestWarehouseTransferPostAuthenticatedInsufficientSourceStock(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 3)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 2)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/"+formatUint64(uint64(transfer.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(409)

	require.Equal(t, int64(3), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(2), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
}

func TestWarehouseTransferPostAuthenticatedSameWarehouse(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse1.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/"+formatUint64(uint64(transfer.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(422)
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
}

func TestWarehouseTransferPostAuthenticatedSuccess(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferAPIToken(t, uint64(f.user.ID))).
		Post("/api/warehouse-transfers/"+formatUint64(uint64(transfer.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertOk()

	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(24), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)

	movements := f.movements(t, "WarehouseTransfer", transfer.ID)
	require.Len(t, movements, 2)
	for _, movement := range movements {
		require.Equal(t, int64(4), movement.Quantity)
		require.Equal(t, uint64(f.user.ID), *movement.CreatedBy)
	}

	var sourceMovement, destinationMovement *models.StockMovement
	for i := range movements {
		movement := &movements[i]
		switch movement.Type {
		case "OUT":
			sourceMovement = movement
		case "IN":
			destinationMovement = movement
		}
	}
	require.NotNil(t, sourceMovement)
	require.NotNil(t, destinationMovement)
	require.Equal(t, int64(10), sourceMovement.BalanceBefore)
	require.Equal(t, int64(6), sourceMovement.BalanceAfter)
	require.Equal(t, int64(20), destinationMovement.BalanceBefore)
	require.Equal(t, int64(24), destinationMovement.BalanceAfter)

	auditLogs, err := f.auditLogs(t, "WarehouseTransfer", transfer.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_TRANSFER", auditLogs[0].Action)
	require.Equal(t, "WarehouseTransfer", auditLogs[0].SubjectType)
	require.Equal(t, uint64(transfer.ID), auditLogs[0].SubjectID)
	require.Equal(t, uint64(f.user.ID), *auditLogs[0].UserID)
}
