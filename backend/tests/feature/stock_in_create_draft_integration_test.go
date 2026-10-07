package feature

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	"goravel/tests"
)

func jsonReader(v any) io.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}

func TestStockInCreateDraftRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}

	// 1. Missing Authorization header
	response, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-ins", jsonReader(map[string]any{
			"supplier_id":      1,
			"warehouse_id":     1,
			"transaction_date": "2026-09-24",
			"items": []map[string]any{
				{"product_id": 1, "quantity": 10, "unit_price": 50000},
			},
		}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	// 2. Malformed Authorization
	response, err = testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-ins", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	// 3. Invalid JWT
	response, err = testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-ins", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockInCreateDraftValidationErrors(t *testing.T) {
	f := newInventoryFixture(t)
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	postDraft := func(payload any) {
		testCase := &tests.TestCase{}
		resp, err := testCase.Http(t).
			WithHeader("Authorization", "Bearer "+token).
			WithHeader("Content-Type", "application/json").
			Post("/api/stock-ins", jsonReader(payload))
		require.NoError(t, err)
		resp.AssertStatus(422)
	}

	postDraftStatus := func(payload any, expectedStatus int) {
		testCase := &tests.TestCase{}
		resp, err := testCase.Http(t).
			WithHeader("Authorization", "Bearer "+token).
			WithHeader("Content-Type", "application/json").
			Post("/api/stock-ins", jsonReader(payload))
		require.NoError(t, err)
		resp.AssertStatus(expectedStatus)
	}

	validDate := "2026-09-24"
	validItems := []map[string]any{
		{"product_id": uint64(f.product1.ID), "quantity": 5, "unit_price": 1000},
	}

	testCases := []struct {
		name         string
		payload      map[string]any
		expectStatus int
	}{
		{
			name: "Missing supplier ID",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Missing warehouse ID",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Supplier not found",
			payload: map[string]any{
				"supplier_id":      999999999,
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 404,
		},
		{
			name: "Warehouse not found",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     999999999,
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 404,
		},
		{
			name: "Invalid date format",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": "24-09-2026",
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Empty items",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items":            []map[string]any{},
			},
			expectStatus: 422,
		},
		{
			name: "Product not found",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": 999999999, "quantity": 5, "unit_price": 1000},
				},
			},
			expectStatus: 404,
		},
		{
			name: "Zero quantity",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 0, "unit_price": 1000},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Negative quantity",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": -5, "unit_price": 1000},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Negative unit price",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 5, "unit_price": -100},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Duplicate product IDs",
			payload: map[string]any{
				"supplier_id":      uint64(f.supplier.ID),
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 5, "unit_price": 1000},
					{"product_id": uint64(f.product1.ID), "quantity": 2, "unit_price": 1200},
				},
			},
			expectStatus: 422,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			postDraftStatus(tc.payload, tc.expectStatus)
		})
	}

	_, _ = f.db.Query().Where("code LIKE ?", "INACTIVE-%").ForceDelete(&models.Warehouse{})
	_, _ = f.db.Query().Where("code LIKE ?", "INACTIVE-%").ForceDelete(&models.Supplier{})
	_, _ = f.db.Query().Where("code LIKE ?", "INACTIVE-%").ForceDelete(&models.Product{})

	// Inactive master data tests
	t.Run("Inactive supplier rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Supplier{}).Where("id = ?", f.supplier.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Supplier{}).Where("id = ?", f.supplier.ID).Update("is_active", true)
		}()

		postDraft(map[string]any{
			"supplier_id":      uint64(f.supplier.ID),
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": validDate,
			"items":            validItems,
		})
	})

	t.Run("Inactive warehouse rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", true)
		}()

		postDraft(map[string]any{
			"supplier_id":      uint64(f.supplier.ID),
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": validDate,
			"items":            validItems,
		})
	})

	t.Run("Inactive product rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", true)
		}()

		postDraft(map[string]any{
			"supplier_id":      uint64(f.supplier.ID),
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": validDate,
			"items": []map[string]any{
				{"product_id": uint64(f.product2.ID), "quantity": 10, "unit_price": 500},
			},
		})
	})
}

func TestStockInCreateDraftSuccessAndInventoryUntouched(t *testing.T) {
	f := newInventoryFixture(t)
	// Initialize initial stock to 10
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-ins", jsonReader(map[string]any{
			"supplier_id":      uint64(f.supplier.ID),
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": "2026-09-24",
			"note":             "Sample draft stock-in",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 7, "unit_price": 2500.50},
			},
		}))
	require.NoError(t, err)
	response.AssertStatus(201)

	var body struct {
		Success bool                          `json:"success"`
		Message string                        `json:"message"`
		Data    services.StockInDraftResponse `json:"data"`
	}
	require.NoError(t, response.Bind(&body))
	require.True(t, body.Success)
	require.NotZero(t, body.Data.ID)
	require.True(t, strings.HasPrefix(body.Data.ReferenceNumber, "IN-"))
	require.Equal(t, "DRAFT", body.Data.Status)
	require.Equal(t, uint64(f.user.ID), body.Data.CreatedBy)
	require.Nil(t, body.Data.PostedAt)
	require.Equal(t, "2026-09-24", body.Data.TransactionDate)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, uint64(f.product1.ID), body.Data.Items[0].ProductID)
	require.Equal(t, int64(7), body.Data.Items[0].Quantity)
	require.Equal(t, 2500.50, body.Data.Items[0].UnitPrice)
	// Verify server-side subtotal calculation: 7 * 2500.50 = 17503.50
	require.Equal(t, 17503.50, body.Data.Items[0].Subtotal)

	// INVENTORY INTEGRITY CHECK:
	// Warehouse stock MUST remain exactly 10
	currentStock := f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID))
	require.Equal(t, int64(10), currentStock.Quantity)

	// Stock movements MUST remain empty (0 movements created)
	movements := f.movements(t, "StockIn", uint(body.Data.ID))
	require.Empty(t, movements)

	// Audit logs for POST_STOCK_IN must NOT exist
	auditLogs, err := f.auditLogs(t, "StockIn", uint(body.Data.ID))
	require.NoError(t, err)
	require.Empty(t, auditLogs)
}

func TestStockInCreateDraftAtomicRollback(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)

	stockInService := services.NewStockInService(f.db)
	// Inject hook simulating failure after header creation
	stockInService.SetOnHeaderCreatedHook(func(tx contractsorm.Query) error {
		return errors.New("simulated item creation failure")
	})

	input := services.CreateStockInDraftInput{
		SupplierID:      uint64(f.supplier.ID),
		WarehouseID:     uint64(f.warehouse1.ID),
		TransactionDate: "2026-09-24",
		Note:            "Should rollback",
		Items: []services.CreateStockInItemInput{
			{ProductID: uint64(f.product1.ID), Quantity: 5, UnitPrice: 100},
		},
	}

	result, err := stockInService.CreateDraft(t.Context(), input, uint64(f.user.ID))
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "simulated item creation failure")

	// Verify header was rolled back
	count, err := f.db.Query().Model(&models.StockIn{}).Where("note = ?", "Should rollback").Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Verify items were rolled back
	itemCount, err := f.db.Query().Model(&models.StockInItem{}).Where("product_id = ? AND unit_price = ?", uint64(f.product1.ID), 100).Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), itemCount)

	// Verify stock is untouched
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
}

func TestStockInCreateDraftAndPostRegression(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}

	// 1. Create draft via API
	createResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-ins", jsonReader(map[string]any{
			"supplier_id":      uint64(f.supplier.ID),
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": "2026-09-24",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 15, "unit_price": 100},
			},
		}))
	require.NoError(t, err)
	createResp.AssertStatus(201)

	var createBody struct {
		Data services.StockInDraftResponse `json:"data"`
	}
	require.NoError(t, createResp.Bind(&createBody))
	stockInID := createBody.Data.ID

	// 2. Verify stock untouched before post
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)

	// 3. Post the draft using existing POST /api/stock-ins/{id}/post
	postResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post(fmt.Sprintf("/api/stock-ins/%d/post", stockInID), nil)
	require.NoError(t, err)
	postResp.AssertOk()

	// 4. Verify status is POSTED and stock incremented: 10 + 15 = 25
	posted, err := repositories.NewStockInRepository(f.db.Query()).FindByID(stockInID)
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)
	require.Equal(t, int64(25), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)

	// 5. Verify movement created
	movements := f.movements(t, "StockIn", uint(stockInID))
	require.Len(t, movements, 1)
	require.Equal(t, int64(15), movements[0].Quantity)
	require.Equal(t, int64(10), movements[0].BalanceBefore)
	require.Equal(t, int64(25), movements[0].BalanceAfter)

	// 6. Verify audit log created
	auditLogs, err := f.auditLogs(t, "StockIn", uint(stockInID))
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_IN", auditLogs[0].Action)
}
