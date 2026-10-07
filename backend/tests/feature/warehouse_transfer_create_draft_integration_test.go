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

func TestWarehouseTransferCreateDraftRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}

	response, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/warehouse-transfers", jsonReader(map[string]any{
			"from_warehouse_id": 1,
			"to_warehouse_id":   2,
			"transaction_date":  "2026-09-24",
			"items": []map[string]any{
				{"product_id": 1, "quantity": 5},
			},
		}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	response, err = testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/warehouse-transfers", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	response, err = testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		WithHeader("Content-Type", "application/json").
		Post("/api/warehouse-transfers", jsonReader(map[string]any{}))
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehouseTransferCreateDraftValidationErrors(t *testing.T) {
	f := newInventoryFixture(t)
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	postDraftStatus := func(payload any, expectedStatus int) {
		testCase := &tests.TestCase{}
		resp, err := testCase.Http(t).
			WithHeader("Authorization", "Bearer "+token).
			WithHeader("Content-Type", "application/json").
			Post("/api/warehouse-transfers", jsonReader(payload))
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
			name: "Missing source warehouse ID",
			payload: map[string]any{
				"to_warehouse_id":  uint64(f.warehouse2.ID),
				"transaction_date": validDate,
				"items":            validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Missing destination warehouse ID",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"transaction_date":  validDate,
				"items":             validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Same source and destination warehouse",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse1.ID),
				"transaction_date":  validDate,
				"items":             validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Source warehouse not found",
			payload: map[string]any{
				"from_warehouse_id": 999999999,
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items":             validItems,
			},
			expectStatus: 404,
		},
		{
			name: "Destination warehouse not found",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   999999999,
				"transaction_date":  validDate,
				"items":             validItems,
			},
			expectStatus: 404,
		},
		{
			name: "Invalid date format",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  "2026/09/24",
				"items":             validItems,
			},
			expectStatus: 422,
		},
		{
			name: "Empty items",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items":             []map[string]any{},
			},
			expectStatus: 422,
		},
		{
			name: "Product not found",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items": []map[string]any{
					{"product_id": 999999999, "quantity": 5},
				},
			},
			expectStatus: 404,
		},
		{
			name: "Zero quantity",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 0},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Negative quantity",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": -5},
				},
			},
			expectStatus: 422,
		},
		{
			name: "Duplicate product IDs",
			payload: map[string]any{
				"from_warehouse_id": uint64(f.warehouse1.ID),
				"to_warehouse_id":   uint64(f.warehouse2.ID),
				"transaction_date":  validDate,
				"items": []map[string]any{
					{"product_id": uint64(f.product1.ID), "quantity": 4},
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

	t.Run("Inactive source warehouse rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse1.ID).Update("is_active", true)
		}()

		postDraftStatus(map[string]any{
			"from_warehouse_id": uint64(f.warehouse1.ID),
			"to_warehouse_id":   uint64(f.warehouse2.ID),
			"transaction_date":  validDate,
			"items":             validItems,
		}, 422)
	})

	t.Run("Inactive destination warehouse rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse2.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Warehouse{}).Where("id = ?", f.warehouse2.ID).Update("is_active", true)
		}()

		postDraftStatus(map[string]any{
			"from_warehouse_id": uint64(f.warehouse1.ID),
			"to_warehouse_id":   uint64(f.warehouse2.ID),
			"transaction_date":  validDate,
			"items":             validItems,
		}, 422)
	})

	t.Run("Inactive product rejected", func(t *testing.T) {
		_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", false)
		defer func() {
			_, _ = f.db.Query().Model(&models.Product{}).Where("id = ?", f.product2.ID).Update("is_active", true)
		}()

		postDraftStatus(map[string]any{
			"from_warehouse_id": uint64(f.warehouse1.ID),
			"to_warehouse_id":   uint64(f.warehouse2.ID),
			"transaction_date":  validDate,
			"items": []map[string]any{
				{"product_id": uint64(f.product2.ID), "quantity": 5},
			},
		}, 422)
	})
}

func TestWarehouseTransferCreateDraftSuccessAndInventoryUntouched(t *testing.T) {
	f := newInventoryFixture(t)
	// Initialize stock in source (10) and destination (20)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/warehouse-transfers", jsonReader(map[string]any{
			"from_warehouse_id": uint64(f.warehouse1.ID),
			"to_warehouse_id":   uint64(f.warehouse2.ID),
			"transaction_date":  "2026-09-24",
			"note":              "Sample draft warehouse transfer",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 4},
			},
		}))
	require.NoError(t, err)
	response.AssertStatus(201)

	var body struct {
		Success bool                                    `json:"success"`
		Message string                                  `json:"message"`
		Data    services.WarehouseTransferDraftResponse `json:"data"`
	}
	require.NoError(t, response.Bind(&body))
	require.True(t, body.Success)
	require.NotZero(t, body.Data.ID)
	require.True(t, strings.HasPrefix(body.Data.ReferenceNumber, "TR-"))
	require.Equal(t, "DRAFT", body.Data.Status)
	require.Equal(t, uint64(f.user.ID), body.Data.CreatedBy)
	require.Nil(t, body.Data.PostedAt)
	require.Equal(t, "2026-09-24", body.Data.TransactionDate)
	require.Equal(t, uint64(f.warehouse1.ID), body.Data.FromWarehouseID)
	require.Equal(t, uint64(f.warehouse2.ID), body.Data.ToWarehouseID)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, uint64(f.product1.ID), body.Data.Items[0].ProductID)
	require.Equal(t, int64(4), body.Data.Items[0].Quantity)

	// INVENTORY INTEGRITY CHECK:
	// Source warehouse stock MUST remain exactly 10
	sourceStock := f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID))
	require.Equal(t, int64(10), sourceStock.Quantity)

	// Destination warehouse stock MUST remain exactly 20
	destStock := f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID))
	require.Equal(t, int64(20), destStock.Quantity)

	// Stock movements MUST remain empty
	movements := f.movements(t, "WarehouseTransfer", uint(body.Data.ID))
	require.Empty(t, movements)

	// Audit logs must NOT exist
	auditLogs, err := f.auditLogs(t, "WarehouseTransfer", uint(body.Data.ID))
	require.NoError(t, err)
	require.Empty(t, auditLogs)
}

func TestWarehouseTransferCreateDraftAtomicRollback(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)

	transferService := services.NewWarehouseTransferService(f.db)
	transferService.SetOnHeaderCreatedHook(func(tx contractsorm.Query) error {
		return errors.New("simulated item creation failure for transfer")
	})

	input := services.CreateWarehouseTransferDraftInput{
		FromWarehouseID: uint64(f.warehouse1.ID),
		ToWarehouseID:   uint64(f.warehouse2.ID),
		TransactionDate: "2026-09-24",
		Note:            "Should rollback",
		Items: []services.CreateWarehouseTransferItemInput{
			{ProductID: uint64(f.product1.ID), Quantity: 5},
		},
	}

	result, err := transferService.CreateDraft(t.Context(), input, uint64(f.user.ID))
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "simulated item creation failure for transfer")

	// Verify header was rolled back
	count, err := f.db.Query().Model(&models.WarehouseTransfer{}).Where("note = ?", "Should rollback").Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Verify items were rolled back
	itemCount, err := f.db.Query().Model(&models.WarehouseTransferItem{}).Where("product_id = ? AND quantity = ?", uint64(f.product1.ID), 5).Count()
	require.NoError(t, err)
	require.Equal(t, int64(0), itemCount)

	// Verify stock is untouched in both warehouses
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
}

func TestWarehouseTransferCreateDraftAndPostRegression(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)

	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(f.user.ID)
	require.NoError(t, err)

	testCase := &tests.TestCase{}

	// 1. Create draft via API
	createResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		WithHeader("Content-Type", "application/json").
		Post("/api/warehouse-transfers", jsonReader(map[string]any{
			"from_warehouse_id": uint64(f.warehouse1.ID),
			"to_warehouse_id":   uint64(f.warehouse2.ID),
			"transaction_date":  "2026-09-24",
			"items": []map[string]any{
				{"product_id": uint64(f.product1.ID), "quantity": 4},
			},
		}))
	require.NoError(t, err)
	createResp.AssertStatus(201)

	var createBody struct {
		Data services.WarehouseTransferDraftResponse `json:"data"`
	}
	require.NoError(t, createResp.Bind(&createBody))
	transferID := createBody.Data.ID

	// 2. Verify stocks untouched before post
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)

	// 3. Post the draft using existing POST /api/warehouse-transfers/{id}/post
	postResp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Post(fmt.Sprintf("/api/warehouse-transfers/%d/post", transferID), nil)
	require.NoError(t, err)
	postResp.AssertOk()

	// 4. Verify status is POSTED, source decremented: 10 - 4 = 6, destination incremented: 20 + 4 = 24
	posted, err := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(transferID)
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(24), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)

	// 5. Verify movements created (OUT for source, IN for destination)
	movements := f.movements(t, "WarehouseTransfer", uint(transferID))
	require.Len(t, movements, 2)

	var sourceMovement, destMovement *models.StockMovement
	for i := range movements {
		switch movements[i].Type {
		case "OUT":
			sourceMovement = &movements[i]
		case "IN":
			destMovement = &movements[i]
		}
	}
	require.NotNil(t, sourceMovement)
	require.NotNil(t, destMovement)
	require.Equal(t, int64(4), sourceMovement.Quantity)
	require.Equal(t, int64(10), sourceMovement.BalanceBefore)
	require.Equal(t, int64(6), sourceMovement.BalanceAfter)
	require.Equal(t, int64(4), destMovement.Quantity)
	require.Equal(t, int64(20), destMovement.BalanceBefore)
	require.Equal(t, int64(24), destMovement.BalanceAfter)

	// 6. Verify audit log created
	auditLogs, err := f.auditLogs(t, "WarehouseTransfer", uint(transferID))
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_TRANSFER", auditLogs[0].Action)
}
