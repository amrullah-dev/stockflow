package feature

import (
	"errors"
	"fmt"
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

func TestStockOutCreateDraftRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}

	response, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-outs", jsonReader(map[string]any{
			"warehouse_id":     1,
			"transaction_date": "2026-09-24",
			"items": []map[string]any{
				{"product_id": 1, "quantity": 5},
			},
		}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	response, err = testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-outs", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	response, err = testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-outs", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestStockOutCreateDraftValidationErrors(t *testing.T) {
	f := newInventoryFixture(t)
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	postDraftStatus := func(payload any, expectedStatus int) {
		testCase := &tests.TestCase{}
		resp, err := testCase.Http(t).
			WithHeader("Authorization", "Bearer "+token).
			WithHeader("Content-Type", "application/json").
			Post("/api/stock-outs", jsonReader(payload))
		require.NoError(t, err)
		resp.AssertStatus(expectedStatus)
	}

	validDate := "2026-09-24"
	validItems := []map[string]any{
		{"product_id": uint64(f.product1.ID), "quantity": 5},
	}

	testCases := []struct {
		name         string
		payload      map[string]any
		expectStatus int
	}{
		{
			name: "Missing warehouse ID",
			payload: map[string]any{
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Warehouse not found",
			payload: map[string]any{
				"warehouse_id":     999999999,
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 404,
		},
		{
			name: "Invalid date format",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": "invalid-date",
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Empty items",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items":            []map[string]any{},
			},
			expectStatus: 422,
		},
		{
			name: "Product not found",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": 999999999, "quantity": 5},
				},
			},
			expectStatus: 404,
		},
		{
			name: "Zero quantity",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 0},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Negative quantity",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": -3},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Duplicate product IDs",
			payload: map[string]any{
				"warehouse_id":     uint64(f.warehouse1.ID),
				"transaction_date": validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 3},
					{"product_id": uint64(f.product1.ID), "quantity": 2},
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

	t.Run("Inactive warehouse rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", true)
		}()

		postDraftStatus(map[string]any{
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": validDate,
			"items":            validItems,
		}, 422)
	})

	t.Run("Inactive product rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", true)
		}()

		postDraftStatus(map[string]any{
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": validDate,
			"items": []map[string]any{
				{"product_id": uint64(f.product2.ID), "quantity": 5},
			},
		}, 422)
	})
}

func TestStockOutCreateDraftSuccessAndInventoryUntouched(t *testing.T) {
	f := newInventoryFixture(t)
	// Initialize stock to 50
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 50)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	// Request asking for 10 units (even if quantity was greater than stock, Draft creation must not check stock)
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-outs", jsonReader(map[string]any{
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": "2026-09-24",
			"destination":      "Distribution Hub East",
			"note":             "Sample draft stock-out",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 10},
			},
		}))
	require.NoError(t, err)
	response.AssertStatus(201)

	var body struct {
		Success bool                           `json:"success"`
		Message string                         `json:"message"`
		Data    services.StockOutDraftResponse `json:"data"`
	}
	require.NoError(t, response.Bind(&body))
	require.True(t, body.Success)
	require.NotZero(t, body.Data.ID)
	require.True(t, strings.HasPrefix(body.Data.ReferenceNumber, "OUT-"))
	require.Equal(t, "DRAFT", body.Data.Status)
	require.Equal(t, uint64(f.user.ID), body.Data.CreatedBy)
	require.Nil(t, body.Data.PostedAt)
	require.Equal(t, "2026-09-24", body.Data.TransactionDate)
	require.Equal(t, "Distribution Hub East", body.Data.Destination)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, uint64(f.product1.ID), body.Data.Items[0].ProductID)
	require.Equal(t, int64(10), body.Data.Items[0].Quantity)

	// INVENTORY INTEGRITY CHECK:
	// Warehouse stock MUST remain untouched at 50
	currentStock := f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID))
	require.Equal(t, int64(50), currentStock.Quantity)

	// Stock movements MUST remain empty
	movements := f.movements(t, "StockOut", uint(body.Data.ID))
	require.Empty(t, movements)

	// Audit logs must NOT exist
	auditLogs, err := f.auditLogs(t, "StockOut", uint(body.Data.ID))
	require.NoError(t, err)
	require.Empty(t, auditLogs)
}

func TestStockOutCreateDraftAtomicRollback(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 50)

	stockOutService := services.NewStockOutService(f.db)
	stockOutService.SetOnHeaderCreatedHook(func(tx contractsorm.Query) error {
		return errors.New("simulated item creation failure for stock-out")
	})

	input := services.CreateStockOutDraftInput{
		WarehouseID:     uint64(f.warehouse1.ID),
		TransactionDate: "2026-09-24",
		Destination:     "Store A",
		Note:            "Should rollback",
		Items: []services.CreateStockOutItemInput{
			{ProductID: uint64(f.product1.ID), Quantity: 5},
		},
	}

	result, err := stockOutService.CreateDraft(t.Context(), input, uint64(f.user.ID))
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "simulated item creation failure for stock-out")

	// Verify header was rolled back
	count, err := f.db.Query().Model(&models.StockOut{}).Where("note = ?", "Should rollback").Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Verify items were rolled back
	itemCount, err := f.db.Query().Model(&models.StockOutItem{}).Where("product_id = ? AND quantity = ?", uint64(f.product1.ID), 5).Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), itemCount)

	// Verify stock is untouched
	require.Equal(t, int64(50), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
}

func TestStockOutCreateDraftAndPostRegression(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 50)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}

	// 1. Create draft via API
	createResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/stock-outs", jsonReader(map[string]any{
			"warehouse_id":     uint64(f.warehouse1.ID),
			"transaction_date": "2026-09-24",
			"destination":      "Client Depot",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 15},
			},
		}))
	require.NoError(t, err)
	createResp.AssertStatus(201)

	var createBody struct {
		Data services.StockOutDraftResponse `json:"data"`
	}
	require.NoError(t, createResp.Bind(&createBody))
	stockOutID := createBody.Data.ID

	// 2. Verify stock untouched before post
	require.Equal(t, int64(50), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)

	// 3. Post the draft using existing POST /api/stock-outs/{id}/post
	postResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post(fmt.Sprintf("/api/stock-outs/%d/post", stockOutID), nil)
	require.NoError(t, err)
	postResp.AssertOk()

	// 4. Verify status is POSTED and stock decremented: 50 - 15 = 35
	posted, err := repositories.NewStockOutRepository(f.db.Query()).FindByID(stockOutID)
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)
	require.Equal(t, int64(35), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)

	// 5. Verify movement created
	movements := f.movements(t, "StockOut", uint(stockOutID))
	require.Len(t, movements, 1)
	require.Equal(t, int64(15), movements[0].Quantity)
	require.Equal(t, int64(50), movements[0].BalanceBefore)
	require.Equal(t, int64(35), movements[0].BalanceAfter)

	// 6. Verify audit log created
	auditLogs, err := f.auditLogs(t, "StockOut", uint(stockOutID))
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_OUT", auditLogs[0].Action)
}
