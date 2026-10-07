package feature

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type loginAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     struct {
				ID   uint64 `json:"id"`
				Name string `json:"name"`
			} `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

func prepareLoginUser(t *testing.T, f *inventoryFixture, password string) {
	t.Helper()
	hashed, err := facades.Hash().Make(password)
	require.NoError(t, err)
	_, err = f.db.Query().Model(&models.User{}).Where("id = ?", f.user.ID).Update("password", hashed)
	require.NoError(t, err)
}

func TestLoginReturnsJWTAndUpdatesLastLogin(t *testing.T) {
	f := newInventoryFixture(t)
	prepareLoginUser(t, f, "correct-password")

	response, err := (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": f.user.Email, "password": "correct-password",
	}))
	require.NoError(t, err)
	response.AssertOk()

	var payload loginAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.NotEmpty(t, payload.Data.Token)
	require.Equal(t, uint64(f.user.ID), payload.Data.User.ID)
	require.Equal(t, f.user.Email, payload.Data.User.Email)
	require.Equal(t, f.role.Name, payload.Data.User.Role.Name)
	require.Empty(t, payload.Data.User.Password)
	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, strings.ToLower(content), "password")

	var user models.User
	require.NoError(t, f.db.Query().Where("id = ?", f.user.ID).Find(&user))
	require.NotNil(t, user.LastLoginAt)
	require.WithinDuration(t, time.Now(), *user.LastLoginAt, 10*time.Second)

	protected, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+payload.Data.Token).
		Get("/api/products")
	require.NoError(t, err)
	protected.AssertOk()
}

func TestLoginValidationAndInvalidCredentials(t *testing.T) {
	f := newInventoryFixture(t)
	prepareLoginUser(t, f, "correct-password")

	cases := []map[string]any{
		{},
		{"email": f.user.Email},
		{"password": "correct-password"},
		{"email": "not-an-email", "password": "correct-password"},
	}
	for _, input := range cases {
		response, err := (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(input))
		require.NoError(t, err)
		response.AssertStatus(422)
	}

	response, err := (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": f.user.Email, "password": "wrong-password",
	}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	response, err = (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": "missing-" + f.user.Email, "password": "correct-password",
	}))
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestLoginRejectsInactiveAndSoftDeletedUsers(t *testing.T) {
	f := newInventoryFixture(t)
	prepareLoginUser(t, f, "correct-password")

	_, err := f.db.Query().Model(&models.User{}).Where("id = ?", f.user.ID).Update("is_active", false)
	require.NoError(t, err)
	response, err := (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": f.user.Email, "password": "correct-password",
	}))
	require.NoError(t, err)
	response.AssertUnauthorized()

	_, err = f.db.Query().Model(&models.User{}).Where("id = ?", f.user.ID).Update("is_active", true)
	require.NoError(t, err)
	_, err = f.db.Query().Where("id = ?", f.user.ID).Delete(&models.User{})
	require.NoError(t, err)
	response, err = (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": f.user.Email, "password": "correct-password",
	}))
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestLoginDoesNotRequireBearerToken(t *testing.T) {
	f := newInventoryFixture(t)
	prepareLoginUser(t, f, "correct-password")
	response, err := (&tests.TestCase{}).Http(t).Post("/api/login", jsonReader(map[string]any{
		"email": f.user.Email, "password": "correct-password",
	}))
	require.NoError(t, err)
	response.AssertOk()
}
