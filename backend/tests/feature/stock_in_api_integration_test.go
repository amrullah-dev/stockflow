package feature

import (
	"strconv"
	"testing"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	"goravel/tests"
)

func TestStockInPostRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Post("/api/stock-ins/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInPostRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Post("/api/stock-ins/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInPostRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Post("/api/stock-ins/1/post", nil)
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInPostAuthenticatedEndToEnd(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post("/api/stock-ins/"+formatUint64(uint64(stockIn.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertOk()

	require.Equal(t, int64(15), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)

	movements := f.movements(t, "StockIn", stockIn.ID)
	require.Len(t, movements, 1)
	require.Equal(t, uint64(f.user.ID), *movements[0].CreatedBy)

	auditLogs, err := f.auditLogs(t, "StockIn", stockIn.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_IN", auditLogs[0].Action)
	require.Equal(t, uint64(f.user.ID), *auditLogs[0].UserID)
}

func TestStockInPostAuthenticatedNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post("/api/stock-ins/999999999/post", nil)
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestStockInPostAuthenticatedAlreadyPosted(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	require.NoError(t, services.NewStockInService(f.db).Post(t.Context(), uint64(stockIn.ID), uint64(f.user.ID)))

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post("/api/stock-ins/"+formatUint64(uint64(stockIn.ID))+"/post", nil)
	require.NoError(t, err)
	response.AssertStatus(409)
}

func formatUint64(value uint64) string {
	return strconv.FormatUint(value, 10)
}
