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

type warehouseTransferReadAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		FromWarehouseID uint64 `json:"from_warehouse_id"`
		ToWarehouseID   uint64 `json:"to_warehouse_id"`
		Status          string `json:"status"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		FromWarehouse   struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"from_warehouse"`
		ToWarehouse struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"to_warehouse"`
		Items []struct{} `json:"items"`
	} `json:"data"`
}

type warehouseTransferDetailReadAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              uint64 `json:"id"`
		ReferenceNumber string `json:"reference_number"`
		FromWarehouseID uint64 `json:"from_warehouse_id"`
		ToWarehouseID   uint64 `json:"to_warehouse_id"`
		Status          string `json:"status"`
		Note            string `json:"note"`
		CreatedBy       uint64 `json:"created_by"`
		PostedAt        any    `json:"posted_at"`
		FromWarehouse   struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"from_warehouse"`
		ToWarehouse struct {
			ID   uint64 `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"to_warehouse"`
		CreatedByUser struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"created_by_user"`
		Items []struct {
			ID         uint64 `json:"id"`
			TransferID uint64 `json:"transfer_id"`
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

func warehouseTransferReadAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestWarehouseTransfersReadRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/warehouse-transfers")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransfersReadRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/warehouse-transfers")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransfersReadRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/warehouse-transfers")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransfersReadListEmpty(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseTransferReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestWarehouseTransfersReadListReturnsHeadersSorted(t *testing.T) {
	f := newInventoryFixture(t)
	first := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID))
	second := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID))

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseTransferReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 2)
	require.Less(t, payload.Data[0].ID, payload.Data[1].ID)
	require.Equal(t, uint64(first.ID), payload.Data[0].ID)
	require.Equal(t, uint64(second.ID), payload.Data[1].ID)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data[0].FromWarehouseID)
	require.Equal(t, f.warehouse1.Code, payload.Data[0].FromWarehouse.Code)
	require.Equal(t, f.warehouse1.Name, payload.Data[0].FromWarehouse.Name)
	require.Equal(t, uint64(f.warehouse2.ID), payload.Data[0].ToWarehouseID)
	require.Equal(t, f.warehouse2.Code, payload.Data[0].ToWarehouse.Code)
	require.Equal(t, f.warehouse2.Name, payload.Data[0].ToWarehouse.Name)
	require.Empty(t, payload.Data[0].Items)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "\"items\"")
	require.NotContains(t, content, "\"password\"")
}

func TestWarehouseTransferReadDetailRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/warehouse-transfers/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferReadDetailRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/warehouse-transfers/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferReadDetailRejectsInvalidID(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers/0")
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestWarehouseTransferReadDetailNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers/999999999")
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestWarehouseTransferReadDetailDraftIncludesRelationsAndItems(t *testing.T) {
	f := newInventoryFixture(t)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID),
		models.WarehouseTransferItem{ProductID: uint64(f.product2.ID), Quantity: 2},
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 1},
	)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers/" + formatUint64(uint64(transfer.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseTransferDetailReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Equal(t, uint64(transfer.ID), payload.Data.ID)
	require.Equal(t, transfer.ReferenceNumber, payload.Data.ReferenceNumber)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data.FromWarehouseID)
	require.Equal(t, f.warehouse1.Name, payload.Data.FromWarehouse.Name)
	require.Equal(t, uint64(f.warehouse2.ID), payload.Data.ToWarehouseID)
	require.Equal(t, f.warehouse2.Name, payload.Data.ToWarehouse.Name)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedBy)
	require.Equal(t, uint64(f.user.ID), payload.Data.CreatedByUser.ID)
	require.Equal(t, f.user.Name, payload.Data.CreatedByUser.Name)
	require.Equal(t, services.StockStatusDraft, payload.Data.Status)
	require.Nil(t, payload.Data.PostedAt)
	require.Len(t, payload.Data.Items, 2)
	require.Less(t, payload.Data.Items[0].ID, payload.Data.Items[1].ID)
	require.Equal(t, uint64(f.product2.ID), payload.Data.Items[0].ProductID)
	require.Equal(t, f.product2.Code, payload.Data.Items[0].Product.Code)
	require.Equal(t, uint64(f.product1.ID), payload.Data.Items[1].ProductID)
	require.Equal(t, f.product1.Code, payload.Data.Items[1].Product.Code)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "\"password\"")
}

func TestWarehouseTransferReadDetailPostedHasPostedAt(t *testing.T) {
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
		WithHeader("Authorization", "Bearer "+warehouseTransferReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouse-transfers/" + formatUint64(uint64(transfer.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseTransferDetailReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Equal(t, services.StockStatusPosted, payload.Data.Status)
	require.NotNil(t, payload.Data.PostedAt)
}
