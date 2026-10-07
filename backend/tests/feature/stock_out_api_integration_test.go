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

func stockOutAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestStockOutPostRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Post("/api/stock-outs/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutPostRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Post("/api/stock-outs/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutPostRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Post("/api/stock-outs/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutPostRejectsInvalidRouteID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutAPIToken(t, uint64(f.user.ID))).
		Post("/api/stock-outs/0/post", nil)
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestStockOutPostAuthenticatedNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutAPIToken(t, uint64(f.user.ID))).
		Post("/api/stock-outs/999999999/post", nil)
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestStockOutPostAuthenticatedAlreadyPosted(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	require.NoError(t, services.NewStockOutService(f.db).Post(t.Context(), uint64(stockOut.ID), uint64(f.user.ID)))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutAPIToken(t, uint64(f.user.ID))).
		Post("/api/stock-outs/"+formatUint64(uint64(stockOut.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(409)
	require.Equal(t, int64(5), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 1)
}

func TestStockOutPostAuthenticatedInsufficientStock(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 3)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutAPIToken(t, uint64(f.user.ID))).
		Post("/api/stock-outs/"+formatUint64(uint64(stockOut.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(409)

	require.Equal(t, int64(3), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)
	stillDraft, err := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
}

func TestStockOutPostAuthenticatedSuccess(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutAPIToken(t, uint64(f.user.ID))).
		Post("/api/stock-outs/"+formatUint64(uint64(stockOut.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertOk()

	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)

	movements := f.movements(t, "StockOut", stockOut.ID)
	require.Len(t, movements, 1)
	require.Equal(t, "OUT", movements[0].Type)
	require.Equal(t, int64(4), movements[0].Quantity)
	require.Equal(t, int64(10), movements[0].BalanceBefore)
	require.Equal(t, int64(6), movements[0].BalanceAfter)
	require.Equal(t, uint64(f.user.ID), *movements[0].CreatedBy)

	auditLogs, err := f.auditLogs(t, "StockOut", stockOut.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_OUT", auditLogs[0].Action)
	require.Equal(t, "StockOut", auditLogs[0].SubjectType)
	require.Equal(t, uint64(stockOut.ID), auditLogs[0].SubjectID)
	require.Equal(t, uint64(f.user.ID), *auditLogs[0].UserID)
}
