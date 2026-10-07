package feature

import (
	"testing"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/tests"
)

type warehouseStockAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID          uint64 `json:"id"`
		WarehouseID uint64 `json:"warehouse_id"`
		ProductID   uint64 `json:"product_id"`
		Quantity    int64  `json:"quantity"`
		Product     struct {
			ID           uint64 `json:"id"`
			Code         string `json:"code"`
			Name         string `json:"name"`
			Unit         string `json:"unit"`
			MinimumStock int64  `json:"minimum_stock"`
			IsActive     bool   `json:"is_active"`
		} `json:"product"`
	} `json:"data"`
}

func warehouseStockAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestWarehouseStocksRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/warehouses/1/stocks")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseStocksRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/warehouses/1/stocks")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseStocksRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/warehouses/1/stocks")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseStocksRejectsInvalidWarehouseID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseStockAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses/0/stocks")
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestWarehouseStocksReturnsNotFoundForMissingWarehouse(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseStockAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses/999999999/stocks")
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestWarehouseStocksReturnsEmptyDataForExistingWarehouseWithoutStock(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseStockAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses/" + formatUint64(uint64(f.warehouse1.ID)) + "/stocks")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseStockAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestWarehouseStocksReturnsProductsAndOrderedStocks(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 20)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	// This stock belongs to another warehouse and must not be returned.
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 99)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseStockAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses/" + formatUint64(uint64(f.warehouse1.ID)) + "/stocks")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseStockAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Len(t, payload.Data, 2)

	first := payload.Data[0]
	second := payload.Data[1]
	require.Equal(t, uint64(f.product1.ID), first.ProductID)
	require.Equal(t, int64(10), first.Quantity)
	require.Equal(t, uint64(f.product1.ID), first.Product.ID)
	require.Equal(t, f.product1.Code, first.Product.Code)
	require.Equal(t, f.product1.Name, first.Product.Name)
	require.Equal(t, f.product1.Unit, first.Product.Unit)
	require.Equal(t, f.product1.MinimumStock, first.Product.MinimumStock)
	require.Equal(t, f.product1.IsActive, first.Product.IsActive)

	require.Equal(t, uint64(f.product2.ID), second.ProductID)
	require.Equal(t, int64(20), second.Quantity)
	require.Equal(t, uint64(f.product2.ID), second.Product.ID)
	require.Equal(t, f.product2.Code, second.Product.Code)
}
