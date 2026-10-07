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

type productSoftDeleteProbe struct {
	ID        uint       `gorm:"column:id"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

type productAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID           uint64 `json:"id"`
		CategoryID   uint64 `json:"category_id"`
		Code         string `json:"code"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		Unit         string `json:"unit"`
		MinimumStock int64  `json:"minimum_stock"`
		IsActive     bool   `json:"is_active"`
		Category     struct {
			ID          uint64 `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			IsActive    bool   `json:"is_active"`
		} `json:"category"`
	} `json:"data"`
}

func productAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestProductsRequiresAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).Get("/api/products")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestProductsRejectsMalformedAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/products")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestProductsRejectsInvalidAuthentication(t *testing.T) {
	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/products")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestProductsReturnsEmptyDataWhenNoProductsExist(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Where("id IN ?", []uint{f.product1.ID, f.product2.ID}).Delete(&models.Product{})
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+productAPIToken(t, uint64(f.user.ID))).
		Get("/api/products")
	require.NoError(t, err)
	response.AssertOk()

	var payload productAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestProductsReturnsFieldsCategoryAndIDAscending(t *testing.T) {
	f := newInventoryFixture(t)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+productAPIToken(t, uint64(f.user.ID))).
		Get("/api/products")
	require.NoError(t, err)
	response.AssertOk()

	var payload productAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Len(t, payload.Data, 2)
	require.Less(t, payload.Data[0].ID, payload.Data[1].ID)

	first := payload.Data[0]
	require.Equal(t, uint64(f.product1.ID), first.ID)
	require.Equal(t, uint64(f.category.ID), first.CategoryID)
	require.Equal(t, f.product1.Code, first.Code)
	require.Equal(t, f.product1.Name, first.Name)
	require.Equal(t, f.product1.Description, first.Description)
	require.Equal(t, f.product1.Unit, first.Unit)
	require.Equal(t, f.product1.MinimumStock, first.MinimumStock)
	require.Equal(t, f.product1.IsActive, first.IsActive)
	require.Equal(t, uint64(f.category.ID), first.Category.ID)
	require.Equal(t, f.category.Name, first.Category.Name)
	require.Equal(t, f.category.Description, first.Category.Description)
	require.Equal(t, f.category.IsActive, first.Category.IsActive)

	second := payload.Data[1]
	require.Equal(t, uint64(f.product2.ID), second.ID)
}

func TestProductsIncludeInactiveProductAndCategory(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Model(&models.Product{}).
		Where("id = ?", f.product1.ID).
		Update("is_active", false)
	require.NoError(t, err)
	_, err = f.db.Query().Model(&models.Category{}).
		Where("id = ?", f.category.ID).
		Update("is_active", false)
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+productAPIToken(t, uint64(f.user.ID))).
		Get("/api/products")
	require.NoError(t, err)
	response.AssertOk()

	var payload productAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 2)
	require.False(t, payload.Data[0].IsActive)
	require.False(t, payload.Data[0].Category.IsActive)
}

func TestProductsExcludeSoftDeletedProduct(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Where("id = ?", f.product2.ID).Delete(&models.Product{})
	require.NoError(t, err)

	testCase := &tests.TestCase{}
	response, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+productAPIToken(t, uint64(f.user.ID))).
		Get("/api/products")
	require.NoError(t, err)
	response.AssertOk()

	var payload productAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.Len(t, payload.Data, 1)
	require.Equal(t, uint64(f.product1.ID), payload.Data[0].ID)

	var probe productSoftDeleteProbe
	require.NoError(t, f.db.Query().Table("products").Where("id = ?", f.product2.ID).Find(&probe))
	require.Equal(t, f.product2.ID, probe.ID)
	require.NotNil(t, probe.DeletedAt)
}
