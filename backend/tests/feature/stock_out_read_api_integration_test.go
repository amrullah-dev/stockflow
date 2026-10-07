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

type stockOutReadAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		WarehouseID     uint64 `json:"warehouse_id"`
		TransactionDate string `json:"transaction_date"`
		Status          string `json:"status"`
		Destination     string `json:"destination"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		Warehouse       struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"warehouse"`
		Items []struct{} `json:"items"`
	} `json:"data"`
}

type stockOutDetailAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		WarehouseID     uint64 `json:"warehouse_id"`
		Status          string `json:"status"`
		Destination     string `json:"destination"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		Warehouse       struct {
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
			ID         uint64 `json:"id"`
			StockOutID uint64 `json:"stock_out_id"`
			ProductID  uint64 `json:"product_id"`
			Quantity   int64  `json:"quantity"`
			Product    struct {
				ID   uint64 `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
				Unit string `json:"unit"`
			} `json:"product"`
		} `json:"items"`
	} `json:"data"`
}

func stockOutReadAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestStockOutsReadRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/stock-outs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutsReadRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/stock-outs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutsReadRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/stock-outs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutsReadListEmpty(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs")
	require.NoError(t, err)
	response.AssertOk()

	var payload stockOutReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestStockOutsReadListReturnsHeadersAndWarehouseSorted(t *testing.T) {
	f := newInventoryFixture(t)
	first := f.createStockOut(t, uint64(f.warehouse1.ID))
	second := f.createStockOut(t, uint64(f.warehouse1.ID))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs")
	require.NoError(t, err)
	response.AssertOk()

	var payload stockOutReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 2)
	require.Less(t, payload.Data[0].ID, payload.Data[1].ID)
	require.Equal(t, uint64(first.ID), payload.Data[0].ID)
	require.Equal(t, uint64(second.ID), payload.Data[1].ID)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data[0].WarehouseID)
	require.Equal(t, f.warehouse1.Code, payload.Data[0].Warehouse.Code)
	require.Equal(t, f.warehouse1.Name, payload.Data[0].Warehouse.Name)
	require.Empty(t, payload.Data[0].Items)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "\"password\"")
	require.NotContains(t, content, "\"items\"")
}

func TestStockOutReadDetailRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/stock-outs/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutReadDetailRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/stock-outs/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutReadDetailRejectsInvalidID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs/0")
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestStockOutReadDetailNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs/999999999")
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestStockOutReadDetailDraftIncludesRelationsAndItems(t *testing.T) {
	f := newInventoryFixture(t)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs/" + formatUint64(uint64(stockOut.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload stockOutDetailAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Equal(t, uint64(stockOut.ID), payload.Data.ID)
	require.Equal(t, stockOut.ReferenceNumber, payload.Data.ReferenceNumber)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data.WarehouseID)
	require.Equal(t, f.warehouse1.Name, payload.Data.Warehouse.Name)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedBy)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedByUser.ID)
	require.Equal(t, f.user.Name, payload.Data.CreatedByUser.Name)
	require.Equal(t, services.StockStatusDraft, payload.Data.Status)
	require.Nil(t, payload.Data.PostedAt)
	require.Len(t, payload.Data.Items, 1)
	require.Equal(t, int64(5), payload.Data.Items[0].Quantity)
	require.Equal(t, uint64(f.product1.ID), payload.Data.Items[0].Product.ID)
	require.Equal(t, f.product1.Code, payload.Data.Items[0].Product.Code)
	require.NotContains(t, mustStockOutContent(t, response), "\"password\"")
}

func TestStockOutReadDetailPostedHasPostedAt(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	require.NoError(t, services.NewStockOutService(f.db).Post(t.Context(), uint64(stockOut.ID), uint64(f.user.ID)))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+stockOutReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/stock-outs/" + formatUint64(uint64(stockOut.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload stockOutDetailAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Equal(t, services.StockStatusPosted, payload.Data.Status)
	require.NotNil(t, payload.Data.PostedAt)
}

func mustStockOutContent(t *testing.T, response interface{ Content() (string, error) }) string {
	t.Helper()
	content, err := response.Content()
	require.NoError(t, err)
	return content
}
