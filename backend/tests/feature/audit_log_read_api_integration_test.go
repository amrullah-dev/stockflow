package feature

import (
	"testing"
	"time"

	frameworkhttp "github.com/goravel/framework/http"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	"goravel/tests"
)

type auditLogReadAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID          uint64  `json:"id"`
		Action      string  `json:"action"`
		SubjectType string  `json:"subject_type"`
		SubjectID   uint64  `json:"subject_id"`
		UserID      *uint64 `json:"user_id"`
		User        *struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
		CreatedAt any `json:"created_at"`
	} `json:"data"`
}

type auditLogDetailReadAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID          uint64  `json:"id"`
		Action      string  `json:"action"`
		SubjectType string  `json:"subject_type"`
		SubjectID   uint64  `json:"subject_id"`
		OldValues   string  `json:"old_values"`
		NewValues   string  `json:"new_values"`
		IPAddress   string  `json:"ip_address"`
		UserAgent   string  `json:"user_agent"`
		UserID      *uint64 `json:"user_id"`
		User        *struct {
			ID       uint64 `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
		CreatedAt any `json:"created_at"`
		UpdatedAt any `json:"updated_at"`
	} `json:"data"`
}

func auditLogReadAPIToken(t *testing.T, userID uint64) string {
	t.Helper()
	token, err := facades.Auth(frameworkhttp.NewContext()).Guard("user").LoginUsingID(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestAuditLogsReadRequiresAuthentication(t *testing.T) {
	response, err := (&tests.TestCase{}).Http(t).Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestAuditLogsReadRejectsMalformedAuthentication(t *testing.T) {
	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Basic invalid-token").
		Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestAuditLogsReadRejectsInvalidAuthentication(t *testing.T) {
	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer invalid-token").
		Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestAuditLogsReadReturnsEmptyData(t *testing.T) {
	f := newInventoryFixture(t)
	var existing []models.AuditLog
	require.NoError(t, f.db.Query().Find(&existing))
	if len(existing) != 0 {
		t.Skip("integration database already contains audit logs")
	}

	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertOk()

	var payload auditLogReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Empty(t, payload.Data)
}

func TestAuditLogsReadListReturnsSortedMappedData(t *testing.T) {
	f := newInventoryFixture(t)
	userID := uint64(f.user.ID)
	repository := repositories.NewAuditLogRepository(f.db.Query())

	first := models.AuditLog{
		UserID:      &userID,
		Action:      "READ_FIRST",
		SubjectType: "AuditRead",
		SubjectID:   uint64(f.product1.ID),
		OldValues:   "old-first",
		NewValues:   "new-first",
		IPAddress:   "127.0.0.1",
		UserAgent:   "audit-test",
	}
	second := models.AuditLog{
		UserID:      &userID,
		Action:      "READ_SECOND",
		SubjectType: "AuditRead",
		SubjectID:   uint64(f.product2.ID),
	}
	require.NoError(t, repository.Create(&first))
	require.NoError(t, repository.Create(&second))

	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := f.db.Query().Model(&models.AuditLog{}).
		Where("id IN ?", []uint{first.ID, second.ID}).
		Update("created_at", fixedTime)
	require.NoError(t, err)

	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, userID)).
		Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertOk()

	var payload auditLogReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.GreaterOrEqual(t, len(payload.Data), 2)
	firstIndex, secondIndex := -1, -1
	for index, item := range payload.Data {
		if item.ID == uint64(first.ID) {
			firstIndex = index
		}
		if item.ID == uint64(second.ID) {
			secondIndex = index
			require.Equal(t, "READ_SECOND", item.Action)
			require.Equal(t, "AuditRead", item.SubjectType)
			require.Equal(t, uint64(f.product2.ID), item.SubjectID)
			require.NotNil(t, item.UserID)
			require.Equal(t, userID, *item.UserID)
			require.NotNil(t, item.User)
			require.Equal(t, f.user.Name, item.User.Name)
			require.Equal(t, f.user.Email, item.User.Email)
			require.Empty(t, item.User.Password)
		}
	}
	require.NotEqual(t, -1, firstIndex)
	require.NotEqual(t, -1, secondIndex)
	require.Less(t, secondIndex, firstIndex)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "password")
	require.Contains(t, content, "old_values")
	require.Contains(t, content, "new_values")
	require.NotContains(t, content, "ip_address")
	require.NotContains(t, content, "user_agent")
}

func TestAuditLogsReadListAllowsEmptyAndNullableUser(t *testing.T) {
	f := newInventoryFixture(t)
	_, err := f.db.Query().Where("user_id = ?", uint64(f.user.ID)).Delete(&models.AuditLog{})
	require.NoError(t, err)

	service := services.NewAuditLogService(repositories.NewAuditLogRepository(f.db.Query()))
	require.NoError(t, service.Create(nil, "SYSTEM_READ", "AuditRead", uint64(f.product1.ID), "", "", "", ""))
	t.Cleanup(func() {
		_, _ = f.db.Query().Where("subject_type = ?", "AuditRead").Delete(&models.AuditLog{})
	})

	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/audit-logs")
	require.NoError(t, err)
	response.AssertOk()

	var payload auditLogReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.NotEmpty(t, payload.Data)
	last := payload.Data[0]
	require.Nil(t, last.UserID)
	require.Nil(t, last.User)
}

func TestAuditLogReadDetailRequiresAuthentication(t *testing.T) {
	response, err := (&tests.TestCase{}).Http(t).Get("/api/audit-logs/1")
	require.NoError(t, err)
	response.AssertUnauthorized()
}

func TestAuditLogReadDetailRejectsInvalidID(t *testing.T) {
	f := newInventoryFixture(t)
	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/audit-logs/0")
	require.NoError(t, err)
	response.AssertStatus(422)
}

func TestAuditLogReadDetailNotFound(t *testing.T) {
	f := newInventoryFixture(t)
	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/audit-logs/999999999")
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestAuditLogReadDetailReturnsCompleteDataAndNullableUser(t *testing.T) {
	f := newInventoryFixture(t)
	service := services.NewAuditLogService(repositories.NewAuditLogRepository(f.db.Query()))
	require.NoError(t, service.Create(nil, "DETAIL_READ", "AuditRead", uint64(f.product1.ID), "before", "after", "192.0.2.1", "Audit-Client"))
	t.Cleanup(func() {
		_, _ = f.db.Query().Where("subject_type = ?", "AuditRead").Delete(&models.AuditLog{})
	})

	var log models.AuditLog
	require.NoError(t, f.db.Query().Where("subject_type = ? AND action = ?", "AuditRead", "DETAIL_READ").Find(&log))

	response, err := (&tests.TestCase{}).Http(t).
		WithHeader("Authorization", "Bearer "+auditLogReadAPIToken(t, uint64(f.user.ID))).
		Get("/api/audit-logs/" + formatUint64(uint64(log.ID)))
	require.NoError(t, err)
	response.AssertOk()

	var payload auditLogDetailReadAPIResponse
	require.NoError(t, response.Bind(&payload))
	require.True(t, payload.Success)
	require.Equal(t, uint64(log.ID), payload.Data.ID)
	require.Equal(t, "DETAIL_READ", payload.Data.Action)
	require.Equal(t, "AuditRead", payload.Data.SubjectType)
	require.Equal(t, uint64(f.product1.ID), payload.Data.SubjectID)
	require.Equal(t, "before", payload.Data.OldValues)
	require.Equal(t, "after", payload.Data.NewValues)
	require.Equal(t, "192.0.2.1", payload.Data.IPAddress)
	require.Equal(t, "Audit-Client", payload.Data.UserAgent)
	require.Nil(t, payload.Data.UserID)
	require.Nil(t, payload.Data.User)
	require.NotNil(t, payload.Data.CreatedAt)
	require.NotNil(t, payload.Data.UpdatedAt)

	content, err := response.Content()
	require.NoError(t, err)
	require.NotContains(t, content, "password")
}
