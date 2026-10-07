package feature

import (
	"context"
	"errors"
	"testing"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
)

func newAuditLogServiceFixture(t *testing.T) (*inventoryFixture, *services.AuditLogService, string) {
	t.Helper()

	f := newInventoryFixture(t)
	subjectType := "AuditTest-" + f.suffix
	t.Cleanup(func() {
		_, _ = f.db.Query().Where("subject_type = ?", subjectType).Delete(&models.AuditLog{})
	})

	service := services.NewAuditLogService(
		repositories.NewAuditLogRepository(f.db.Query()),
	)
	return f, service, subjectType
}

func (f *inventoryFixture) auditLogs(t *testing.T, subjectType string, subjectID uint) ([]models.AuditLog, error) {
	t.Helper()
	return repositories.NewAuditLogRepository(f.db.Query()).FindBySubject(subjectType, uint64(subjectID))
}

func TestAuditLogServiceCreate(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)
	userID := uint64(f.user.ID)

	err := service.Create(
		&userID,
		"CREATE",
		subjectType,
		uint64(f.product1.ID),
		`{"status":"DRAFT"}`,
		`{"status":"POSTED"}`,
		"127.0.0.1",
		"StockFlow-Test-Agent",
	)
	require.NoError(t, err)

	logs, err := service.FindBySubject(subjectType, uint64(f.product1.ID))
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.NotNil(t, logs[0].UserID)
	require.Equal(t, userID, *logs[0].UserID)
	require.Equal(t, "CREATE", logs[0].Action)
	require.Equal(t, `{"status":"DRAFT"}`, logs[0].OldValues)
	require.Equal(t, `{"status":"POSTED"}`, logs[0].NewValues)
	require.Equal(t, "127.0.0.1", logs[0].IPAddress)
	require.Equal(t, "StockFlow-Test-Agent", logs[0].UserAgent)
}

func TestAuditLogServiceAllowsNullableUser(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)

	require.NoError(t, service.Create(
		nil,
		"SYSTEM_ACTION",
		subjectType,
		uint64(f.product1.ID),
		"",
		"",
		"",
		"",
	))

	logs, err := service.FindBySubject(subjectType, uint64(f.product1.ID))
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Nil(t, logs[0].UserID)
}

func TestAuditLogServiceRejectsInvalidRequiredData(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)
	userID := uint64(f.user.ID)

	tests := []struct {
		name string
		err  error
		call func() error
	}{
		{
			name: "empty action",
			err:  services.ErrAuditLogInvalidAction,
			call: func() error {
				return service.Create(&userID, "", subjectType, uint64(f.product1.ID), "", "", "", "")
			},
		},
		{
			name: "empty subject type",
			err:  services.ErrAuditLogInvalidSubject,
			call: func() error {
				return service.Create(&userID, "CREATE", "", uint64(f.product1.ID), "", "", "", "")
			},
		},
		{
			name: "zero subject id",
			err:  services.ErrAuditLogInvalidSubjectID,
			call: func() error {
				return service.Create(&userID, "CREATE", subjectType, 0, "", "", "", "")
			},
		},
		{
			name: "zero user id",
			err:  services.ErrAuditLogInvalidUserID,
			call: func() error {
				zero := uint64(0)
				return service.Create(&zero, "CREATE", subjectType, uint64(f.product1.ID), "", "", "", "")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestAuditLogServiceQueries(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)
	userID := uint64(f.user.ID)

	require.NoError(t, service.Create(&userID, "CREATE", subjectType, uint64(f.product1.ID), "", "created", "", ""))
	require.NoError(t, service.Create(&userID, "UPDATE", subjectType, uint64(f.product1.ID), "created", "updated", "", ""))
	require.NoError(t, service.Create(nil, "CREATE", subjectType, uint64(f.product2.ID), "", "created", "", ""))

	bySubject, err := service.FindBySubject(subjectType, uint64(f.product1.ID))
	require.NoError(t, err)
	require.Len(t, bySubject, 2)

	byUser, err := service.FindByUser(userID)
	require.NoError(t, err)
	require.Len(t, byUser, 2)

	byAction, err := service.FindByAction("CREATE")
	require.NoError(t, err)
	require.Len(t, byAction, 2)

	byDate, err := service.FindByDateRange(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(byDate), 3)
}

func TestAuditLogServiceReturnsDatabaseErrors(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)
	invalidUserID := uint64(1<<63 - 1)

	err := service.Create(&invalidUserID, "CREATE", subjectType, uint64(f.product1.ID), "", "", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "create audit log")
}

func TestAuditLogServiceTransactionRollbackAndCommit(t *testing.T) {
	f, service, subjectType := newAuditLogServiceFixture(t)
	userID := uint64(f.user.ID)
	rollbackSubjectID := uint64(f.product1.ID)
	commitSubjectID := uint64(f.product2.ID)

	rollbackErr := f.db.Transaction(func(tx contractsorm.Query) error {
		txService := services.NewAuditLogService(repositories.NewAuditLogRepository(tx))
		if err := txService.Create(&userID, "ROLLBACK_TEST", subjectType, rollbackSubjectID, "", "", "", ""); err != nil {
			return err
		}
		return errors.New("force audit transaction rollback")
	})
	require.Error(t, rollbackErr)

	rolledBack, err := service.FindBySubject(subjectType, rollbackSubjectID)
	require.NoError(t, err)
	require.Empty(t, rolledBack)

	commitErr := f.db.Transaction(func(tx contractsorm.Query) error {
		txService := services.NewAuditLogService(repositories.NewAuditLogRepository(tx))
		return txService.Create(&userID, "COMMIT_TEST", subjectType, commitSubjectID, "", "", "", "")
	})
	require.NoError(t, commitErr)

	committed, err := service.FindBySubject(subjectType, commitSubjectID)
	require.NoError(t, err)
	require.Len(t, committed, 1)
}

func TestStockInPostingCreatesAuditLog(t *testing.T) {
	f, _, _ := newAuditLogServiceFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	t.Cleanup(func() {
		_, _ = f.db.Query().Where("subject_type = ? AND subject_id = ?", "StockIn", stockIn.ID).Delete(&models.AuditLog{})
	})

	err := services.NewStockInService(f.db).Post(context.TODO(), uint64(stockIn.ID), uint64(f.user.ID))
	require.NoError(t, err)

	posted, err := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)
	require.Equal(t, int64(15), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)

	movements := f.movements(t, "StockIn", stockIn.ID)
	require.Len(t, movements, 1)
	require.Equal(t, "IN", movements[0].Type)

	logs, err := f.auditLogs(t, "StockIn", stockIn.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "POST_STOCK_IN", logs[0].Action)
	require.Equal(t, "StockIn", logs[0].SubjectType)
	require.Equal(t, uint64(stockIn.ID), logs[0].SubjectID)
	require.NotNil(t, logs[0].UserID)
	require.Equal(t, uint64(f.user.ID), *logs[0].UserID)
	require.Equal(t, "status=DRAFT", logs[0].OldValues)
	require.Equal(t, "status=POSTED", logs[0].NewValues)
	require.Empty(t, logs[0].IPAddress)
	require.Empty(t, logs[0].UserAgent)
}

func TestStockInAuditFailureRollsBackTransactionScopedWork(t *testing.T) {
	f, auditService, subjectType := newAuditLogServiceFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	actorID := uint64(f.user.ID)
	invalidAuditUserID := uint64(1<<63 - 1)
	err := f.db.Transaction(func(tx contractsorm.Query) error {
		stockInRepository := repositories.NewStockInRepository(tx)
		warehouseStockRepository := repositories.NewWarehouseStockRepository(tx)
		stockMovementRepository := repositories.NewStockMovementRepository(tx)
		txAuditService := services.NewAuditLogService(repositories.NewAuditLogRepository(tx))

		lockedStockIn, err := stockInRepository.FindByIDForUpdate(uint64(stockIn.ID))
		if err != nil {
			return err
		}
		stock, err := warehouseStockRepository.FindByWarehouseAndProductForUpdate(
			lockedStockIn.WarehouseID,
			lockedStockIn.Items[0].ProductID,
		)
		if err != nil {
			return err
		}
		balanceBefore := stock.Quantity
		balanceAfter := balanceBefore + lockedStockIn.Items[0].Quantity
		if err := warehouseStockRepository.UpdateQuantity(stock, balanceAfter); err != nil {
			return err
		}

		referenceID := uint64(lockedStockIn.ID)
		movement := &models.StockMovement{
			WarehouseID:   lockedStockIn.WarehouseID,
			ProductID:     lockedStockIn.Items[0].ProductID,
			Type:          "IN",
			Quantity:      lockedStockIn.Items[0].Quantity,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			ReferenceType: "StockIn",
			ReferenceID:   &referenceID,
			CreatedBy:     &actorID,
		}
		if err := stockMovementRepository.Create(movement); err != nil {
			return err
		}

		return txAuditService.Create(
			&invalidAuditUserID,
			"POST_STOCK_IN",
			"StockIn",
			uint64(lockedStockIn.ID),
			"status=DRAFT",
			"status=POSTED",
			"",
			"",
		)
	})
	require.Error(t, err)

	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockIn", stockIn.ID), 0)
	logs, findErr := auditService.FindBySubject(subjectType, uint64(stockIn.ID))
	require.NoError(t, findErr)
	require.Empty(t, logs)

	stillDraft, findErr := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}
