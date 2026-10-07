package feature

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/tests"
)

type masterDataCreatedResponse struct {
	Data struct {
		ID uint `json:"id"`
	} `json:"data"`
}

func TestCategoryCRUDAPIAndSoftDelete(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	name := "API Category " + f.suffix
	header := map[string]string{"Authorization": "Bearer " + token, "Content-Type": "application/json"}

	response, err := (&tests.TestCase{}).Http(t).WithHeaders(header).Post("/api/categories", jsonReader(map[string]any{"name": name, "description": "API category", "is_active": true}))
	require.NoError(t, err)
	response.AssertCreated()
	var created masterDataCreatedResponse
	require.NoError(t, response.Bind(&created))
	require.NotZero(t, created.Data.ID)
	defer func() { _, _ = f.db.Query().Where("id = ?", created.Data.ID).ForceDelete(&models.Category{}) }()

	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Get(fmt.Sprintf("/api/categories/%d", created.Data.ID))
	require.NoError(t, err)
	response.AssertOk()

	response, err = (&tests.TestCase{}).Http(t).WithHeaders(header).Put(fmt.Sprintf("/api/categories/%d", created.Data.ID), jsonReader(map[string]any{"name": name + " Updated", "description": "Updated", "is_active": false}))
	require.NoError(t, err)
	response.AssertOk()

	response, err = (&tests.TestCase{}).Http(t).WithHeaders(header).Post("/api/categories", jsonReader(map[string]any{"name": name + " Updated"}))
	require.NoError(t, err)
	response.AssertStatus(409)

	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/categories/%d", created.Data.ID), nil)
	require.NoError(t, err)
	response.AssertOk()
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Get(fmt.Sprintf("/api/categories/%d", created.Data.ID))
	require.NoError(t, err)
	response.AssertNotFound()

	var probe struct {
		ID        uint    `gorm:"column:id"`
		DeletedAt *string `gorm:"column:deleted_at"`
	}
	require.NoError(t, f.db.Query().Table("categories").Where("id = ?", created.Data.ID).Find(&probe))
	require.Equal(t, created.Data.ID, probe.ID)
	require.NotNil(t, probe.DeletedAt)
}

func TestSupplierCRUDAPIAndSoftDelete(t *testing.T) {
	f := newInventoryFixture(t)
	token := phase3Token(t, f.user.ID)
	code := "API-SUP-" + f.suffix
	header := map[string]string{"Authorization": "Bearer " + token, "Content-Type": "application/json"}

	response, err := (&tests.TestCase{}).Http(t).WithHeaders(header).Post("/api/suppliers", jsonReader(map[string]any{"code": code, "name": "API Supplier", "phone": "0800", "email": "supplier@example.test", "address": "Warehouse road", "is_active": true}))
	require.NoError(t, err)
	response.AssertCreated()
	var created masterDataCreatedResponse
	require.NoError(t, response.Bind(&created))
	require.NotZero(t, created.Data.ID)
	defer func() { _, _ = f.db.Query().Where("id = ?", created.Data.ID).ForceDelete(&models.Supplier{}) }()

	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Get("/api/suppliers")
	require.NoError(t, err)
	response.AssertOk()
	response, err = (&tests.TestCase{}).Http(t).WithHeaders(header).Put(fmt.Sprintf("/api/suppliers/%d", created.Data.ID), jsonReader(map[string]any{"code": code, "name": "Updated Supplier", "is_active": false}))
	require.NoError(t, err)
	response.AssertOk()
	response, err = (&tests.TestCase{}).Http(t).WithHeaders(header).Post("/api/suppliers", jsonReader(map[string]any{"code": code, "name": "Duplicate Supplier"}))
	require.NoError(t, err)
	response.AssertStatus(409)

	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Delete(fmt.Sprintf("/api/suppliers/%d", created.Data.ID), nil)
	require.NoError(t, err)
	response.AssertOk()
	response, err = (&tests.TestCase{}).Http(t).WithHeader("Authorization", "Bearer "+token).Get(fmt.Sprintf("/api/suppliers/%d", created.Data.ID))
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestMasterDataAPIRequiresAuthentication(t *testing.T) {
	for _, path := range []string{"/api/categories", "/api/suppliers"} {
		response, err := (&tests.TestCase{}).Http(t).Get(path)
		require.NoError(t, err)
		response.AssertUnauthorized()
	}
}
