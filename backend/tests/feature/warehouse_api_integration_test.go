package feature

import (
	"testing"
	"time"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type warehouseSoftDeleteProbe struct {
	ID        uint       `gorm:"column:id"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

type warehouseAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID          uint64 `json:"id"`
		Code        string `json:"code"`
		Name        string `json:"name"`
		Location    string `json:"location"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	} `json:"data"`
}

func warehouseAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestWarehousesRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehousesRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehousesRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestWarehousesReturnsEmptyDataWhenNoWarehousesExist(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Where("id IN ?", []uint{f.warehouse1.ID, f.warehouse2.ID}).Delete(&models.Warehouse{})
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestWarehousesReturnsFieldsAndIDAscending(t *testing.T) {
	f := newInventoryFixture(t)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Len(t, payload.Data, 2)
	require.Less(t, payload.Data[0].ID, payload.Data[1].ID)

	first := payload.Data[0]
	require.Equal(t, uint64(f.warehouse1.ID), first.ID)
	require.Equal(t, f.warehouse1.Code, first.Code)
	require.Equal(t, f.warehouse1.Name, first.Name)
	require.Equal(t, f.warehouse1.Location, first.Location)
	require.Equal(t, f.warehouse1.Description, first.Description)
	require.Equal(t, f.warehouse1.IsActive, first.IsActive)

	second := payload.Data[1]
	require.Equal(t, uint64(f.warehouse2.ID), second.ID)
}

func TestWarehousesIncludeInactiveWarehouse(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Model(&models.Warehouse{}).
		Where("id = ?", f.warehouse1.ID).
		Update("is_active", false)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 2)
	require.False(t, payload.Data[0].IsActive)
}

func TestWarehousesExcludeSoftDeletedWarehouse(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Where("id = ?", f.warehouse2.ID).Delete(&models.Warehouse{})
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertOk()

	var payload warehouseAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 1)
	require.Equal(t, uint64(f.warehouse1.ID), payload.Data[0].ID)

	var probe warehouseSoftDeleteProbe
	require.NoError(t, f.db.Query().Table("warehouses").Where("id = ?", f.warehouse2.ID).Find(&probe))
	require.Equal(t, f.warehouse2.ID, probe.ID)
	require.NotNil(t, probe.DeletedAt)
}

func TestWarehousesResponseDoesNotIncludeRelationships(t *testing.T) {
	f := newInventoryFixture(t)
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+warehouseAPIToken(t, uint64(f.user.ID))).
		Get("/api/warehouses")
	require.NoError(t, err)
	response.AssertOk()

	content, err := response.Content()
	require.NoError(t, err)
	for _, field := range []string{"\"stocks\"", "\"stock_ins\"", "\"stock_outs\"", "\"transfers_from\"", "\"transfers_to\"", "\"stock_movements\""} {
		require.NotContains(t, content, field)
	}
}
