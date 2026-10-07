package feature

import (
	"testing"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/tests"
)

type stockInReadAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		SupplierID      uint64 `json:"supplier_id"`
		WarehouseID     uint64 `json:"warehouse_id"`
		TransactionDate string `json:"transaction_date"`
		Status          string `json:"status"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		Supplier        struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"supplier"`
		Warehouse struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"warehouse"`
		CreatedByUser struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"created_by_user"`
		Items []struct {
			ID        uint64  `json:"id"`
			StockInID uint64  `json:"stock_in_id"`
			ProductID uint64  `json:"product_id"`
			Quantity  int64   `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
			Subtotal  float64 `json:"subtotal"`
			Product   struct {
				ID   uint64 `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
				Unit string `json:"unit"`
			} `json:"product"`
		} `json:"items"`
	} `json:"data"`
}

type stockInDetailAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		SupplierID      uint64 `json:"supplier_id"`
		WarehouseID     uint64 `json:"warehouse_id"`
		Status          string `json:"status"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		Supplier        struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"supplier"`
		Warehouse struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"warehouse"`
		CreatedByUser struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"created_by_user"`
		Items []struct {
			ID        uint64  `json:"id"`
			StockInID uint64  `json:"stock_in_id"`
			ProductID uint64  `json:"product_id"`
			Quantity  int64   `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
			Subtotal  float64 `json:"subtotal"`
			Product   struct {
				ID   uint64 `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
				Unit string `json:"unit"`
			} `json:"product"`
		} `json:"items"`
	} `json:"data"`
}

func stockInReadAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestStockInsReadRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/stock-ins")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInsReadRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/stock-ins")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInsReadRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/stock-ins")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInsReadListEmpty(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins")
	require.NoError(t, err)
	response.AssertOk()

	var payload stockInReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestStockInsReadListReturnsHeadersSortedWithSummaries(t *testing.T) {
	f := newInventoryFixture(t)
	first := f.createStockIn(t, uint64(f.warehouse1.ID))
	second := f.createStockIn(t, uint64(f.warehouse1.ID))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins")
	require.NoError(t, err)
	response.AssertOk()

	var payload stockInReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 2)
	require.Less(t, payload.Data[0].ID, payload.Data[1].ID)
	require.Equal(t, uint64(first.ID), payload.Data[0].ID)
	require.Equal(t, uint64(second.ID), payload.Data[1].ID)
	require.Equal(t, uint64(f.supplier.ID), payload.Data[0].SupplierID)
	require.Equal(t, f.supplier.Code, payload.Data[0].Supplier.Code)
	require.Equal(t, f.supplier.Name, payload.Data[0].Supplier.Name)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data[0].WarehouseID)
	require.Equal(t, f.warehouse1.Code, payload.Data[0].Warehouse.Code)
	require.Empty(t, payload.Data[0].Items)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "\"password\"")
	require.NotContains(t, content, "\"items\"")
}

func TestStockInReadDetailRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/stock-ins/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInReadDetailRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/stock-ins/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInReadDetailRejectsInvalidID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins/0")
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestStockInReadDetailNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins/999999999")
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestStockInReadDetailDraftIncludesRelationsAndItems(t *testing.T) {
	f := newInventoryFixture(t)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
		UnitPrice: 12.5,
		Subtotal:  62.5,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins/" + formatUint64(uint64(stockIn.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload stockInDetailAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Equal(t, uint64(stockIn.ID), payload.Data.ID)
	require.Equal(t, stockIn.ReferenceNumber, payload.Data.ReferenceNumber)
	require.Equal(t, uint64(f.supplier.ID), payload.Data.SupplierID)
	require.Equal(t, f.supplier.Name, payload.Data.Supplier.Name)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data.WarehouseID)
	require.Equal(t, f.warehouse1.Name, payload.Data.Warehouse.Name)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedBy)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedByUser.ID)
	require.Equal(t, f.user.Name, payload.Data.CreatedByUser.Name)
	require.Equal(t, services.StockStatusDraft, payload.Data.Status)
	require.Nil(t, payload.Data.PostedAt)
	require.Len(t, payload.Data.Items, 1)
	require.Equal(t, int64(5), payload.Data.Items[0].Quantity)
	require.Equal(t, 12.5, payload.Data.Items[0].UnitPrice)
	require.Equal(t, 62.5, payload.Data.Items[0].Subtotal)
	require.Equal(t, uint64(f.product1.ID), payload.Data.Items[0].Product.ID)
	require.Equal(t, f.product1.Code, payload.Data.Items[0].Product.Code)
	require.Equal(t, f.product1.Name, payload.Data.Items[0].Product.Name)
	require.NotContains(t, mustStockInContent(t, response), "\"password\"")
}

func TestStockInReadDetailPostedHasPostedAt(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	require.NoError(t, services.NewStockInService(f.db).Post(t.Context(), uint64(stockIn.ID), uint64(f.user.ID)))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockInReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-ins/" + formatUint64(uint64(stockIn.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload stockInDetailAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Equal(t, services.StockStatusPosted, payload.Data.Status)
	require.NotNil(t, payload.Data.PostedAt)
}

func mustStockInContent(t *testing.T, response interface{ Content() (string, error) }) string {
	t.Helper()
	content, err := response.Content()
	require.NoError(t, err)
	return content
}
