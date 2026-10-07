package feature

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/stretchr/testify/require"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/repositories"
	"goravel/app/services"
	_ "goravel/tests"
)

type inventoryFixture struct {
	db         contractsorm.Orm
	suffix     string
	role       models.Role
	user       models.User
	category   models.Category
	product1   models.Product
	product2   models.Product
	supplier   models.Supplier
	warehouse1 models.Warehouse
	warehouse2 models.Warehouse
}

func newInventoryFixture(t *testing.T) *inventoryFixture {
	t.Helper()

	f := &inventoryFixture{
		db:     facades.Orm(),
		suffix: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	f.role = models.Role{
		Name:        "test-role-" + f.suffix,
		Description: "StockFlow integration test role",
	}
	f.create(t, &f.role)
	f.user = models.User{
		RoleID:   uint64(f.role.ID),
		Name:     "StockFlow Integration User",
		Email:    "stockflow-" + f.suffix + "@example.test",
		Password: "test-password",
		IsActive: true,
	}
	f.create(t, &f.user)
	f.category = models.Category{
		Name:     "test-category-" + f.suffix,
		IsActive: true,
	}
	f.create(t, &f.category)
	f.product1 = models.Product{
		CategoryID: uint64(f.category.ID),
		Code:       "TEST-P1-" + f.suffix,
		Name:       "StockFlow Test Product 1",
		Unit:       "pcs",
		IsActive:   true,
	}
	f.create(t, &f.product1)
	f.product2 = models.Product{
		CategoryID: uint64(f.category.ID),
		Code:       "TEST-P2-" + f.suffix,
		Name:       "StockFlow Test Product 2",
		Unit:       "pcs",
		IsActive:   true,
	}
	f.create(t, &f.product2)
	f.supplier = models.Supplier{
		Code:     "TEST-SUP-" + f.suffix,
		Name:     "StockFlow Test Supplier",
		IsActive: true,
	}
	f.create(t, &f.supplier)
	f.warehouse1 = models.Warehouse{
		Code:     "TEST-W1-" + f.suffix,
		Name:     "StockFlow Test Warehouse 1",
		IsActive: true,
	}
	f.create(t, &f.warehouse1)
	f.warehouse2 = models.Warehouse{
		Code:     "TEST-W2-" + f.suffix,
		Name:     "StockFlow Test Warehouse 2",
		IsActive: true,
	}
	f.create(t, &f.warehouse2)

	t.Cleanup(f.cleanup)
	return f
}

func (f *inventoryFixture) create(t *testing.T, value any) {
	t.Helper()
	require.NoError(t, f.db.Query().Create(value))
}

func (f *inventoryFixture) cleanup() {
	q := f.db.Query()

	for _, id := range []uint{f.product1.ID, f.product2.ID} {
		if id == 0 {
			continue
		}
		_, _ = q.Where("product_id = ?", uint64(id)).Delete(&models.StockMovement{})
		_, _ = q.Where("product_id = ?", uint64(id)).Delete(&models.WarehouseStock{})
	}

	if f.user.ID != 0 {
		_, _ = q.Where("created_by = ?", uint64(f.user.ID)).Delete(&models.StockMovement{})
		_, _ = q.Where("user_id = ?", uint64(f.user.ID)).Delete(&models.AuditLog{})
	}

	if f.product1.ID != 0 || f.product2.ID != 0 {
		productIDs := []any{uint64(f.product1.ID), uint64(f.product2.ID)}
		_, _ = q.Where("product_id IN ?", productIDs).Delete(&models.StockInItem{})
		_, _ = q.Where("product_id IN ?", productIDs).Delete(&models.StockOutItem{})
		_, _ = q.Where("product_id IN ?", productIDs).Delete(&models.WarehouseTransferItem{})
	}

	if f.user.ID != 0 {
		_, _ = q.Where("created_by = ?", uint64(f.user.ID)).Delete(&models.StockIn{})
		_, _ = q.Where("created_by = ?", uint64(f.user.ID)).Delete(&models.StockOut{})
		_, _ = q.Where("created_by = ?", uint64(f.user.ID)).Delete(&models.WarehouseTransfer{})
	}

	for _, id := range []uint{f.warehouse1.ID, f.warehouse2.ID} {
		if id != 0 {
			_, _ = q.Where("warehouse_id = ?", uint64(id)).Delete(&models.WarehouseStock{})
		}
	}

	for _, id := range []uint{f.product1.ID, f.product2.ID} {
		if id != 0 {
			_, _ = q.Where("id = ?", id).ForceDelete(&models.Product{})
		}
	}
	if f.supplier.ID != 0 {
		_, _ = q.Where("id = ?", f.supplier.ID).ForceDelete(&models.Supplier{})
	}
	for _, id := range []uint{f.warehouse1.ID, f.warehouse2.ID} {
		if id != 0 {
			_, _ = q.Where("id = ?", id).ForceDelete(&models.Warehouse{})
		}
	}
	if f.user.ID != 0 {
		_, _ = q.Where("id = ?", f.user.ID).ForceDelete(&models.User{})
	}
	if f.category.ID != 0 {
		_, _ = q.Where("id = ?", f.category.ID).ForceDelete(&models.Category{})
	}
	if f.role.ID != 0 {
		_, _ = q.Where("id = ?", f.role.ID).Delete(&models.Role{})
	}
}

func (f *inventoryFixture) addStock(t *testing.T, warehouseID, productID uint64, quantity int64) *models.WarehouseStock {
	t.Helper()
	stock := &models.WarehouseStock{
		WarehouseID: warehouseID,
		ProductID:   productID,
		Quantity:    quantity,
	}
	f.create(t, stock)
	return stock
}

func (f *inventoryFixture) createStockIn(t *testing.T, warehouseID uint64, items ...models.StockInItem) *models.StockIn {
	t.Helper()
	stockIn := &models.StockIn{
		ReferenceNumber: "TEST-SI-" + f.suffix + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		SupplierID:      uint64(f.supplier.ID),
		WarehouseID:     warehouseID,
		TransactionDate: time.Now(),
		Status:          services.StockStatusDraft,
		CreatedBy:       uint64(f.user.ID),
	}
	f.create(t, stockIn)
	for i := range items {
		items[i].StockInID = uint64(stockIn.ID)
		f.create(t, &items[i])
	}
	return stockIn
}

func (f *inventoryFixture) createStockOut(t *testing.T, warehouseID uint64, items ...models.StockOutItem) *models.StockOut {
	t.Helper()
	stockOut := &models.StockOut{
		ReferenceNumber: "TEST-SO-" + f.suffix + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		WarehouseID:     warehouseID,
		TransactionDate: time.Now(),
		Status:          services.StockStatusDraft,
		CreatedBy:       uint64(f.user.ID),
	}
	f.create(t, stockOut)
	for i := range items {
		items[i].StockOutID = uint64(stockOut.ID)
		f.create(t, &items[i])
	}
	return stockOut
}

func (f *inventoryFixture) createTransfer(t *testing.T, from, to uint64, items ...models.WarehouseTransferItem) *models.WarehouseTransfer {
	t.Helper()
	transfer := &models.WarehouseTransfer{
		ReferenceNumber: "TEST-WT-" + f.suffix + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		FromWarehouseID: from,
		ToWarehouseID:   to,
		TransactionDate: time.Now(),
		Status:          services.StockStatusDraft,
		CreatedBy:       uint64(f.user.ID),
	}
	f.create(t, transfer)
	for i := range items {
		items[i].TransferID = uint64(transfer.ID)
		f.create(t, &items[i])
	}
	return transfer
}

func (f *inventoryFixture) stock(t *testing.T, warehouseID, productID uint64) *models.WarehouseStock {
	t.Helper()
	stock, err := repositories.NewWarehouseStockRepository(f.db.Query()).FindByWarehouseAndProduct(warehouseID, productID)
	require.NoError(t, err)
	return stock
}

func requireNoAuditLogs(t *testing.T, f *inventoryFixture, subjectType string, subjectID uint) {
	t.Helper()
	logs, err := f.auditLogs(t, subjectType, subjectID)
	require.NoError(t, err)
	require.Empty(t, logs)
}

func (f *inventoryFixture) movements(t *testing.T, referenceType string, referenceID uint) []models.StockMovement {
	t.Helper()
	movements, err := repositories.NewStockMovementRepository(f.db.Query()).FindByReference(referenceType, uint64(referenceID))
	require.NoError(t, err)
	return movements
}

func TestStockInPostingSuccess(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})

	err := services.NewStockInService(f.db).Post(context.Background(), uint64(stockIn.ID), uint64(f.user.ID))
	require.NoError(t, err)

	require.Equal(t, int64(15), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)

	movements := f.movements(t, "StockIn", stockIn.ID)
	require.Len(t, movements, 1)
	require.Equal(t, "IN", movements[0].Type)
	require.Equal(t, int64(5), movements[0].Quantity)
	require.Equal(t, int64(10), movements[0].BalanceBefore)
	require.Equal(t, int64(15), movements[0].BalanceAfter)
}

func TestStockInMultiItemNonSortedOrder(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 20)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID),
		models.StockInItem{ProductID: uint64(f.product2.ID), Quantity: 3},
		models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 2},
	)

	err := services.NewStockInService(f.db).Post(context.Background(), uint64(stockIn.ID), uint64(f.user.ID))
	require.NoError(t, err)
	require.Equal(t, int64(12), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(23), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID)).Quantity)

	movements := f.movements(t, "StockIn", stockIn.ID)
	require.Len(t, movements, 2)
	for _, movement := range movements {
		require.Equal(t, "IN", movement.Type)
	}
	auditLogs, err := f.auditLogs(t, "StockIn", stockIn.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_IN", auditLogs[0].Action)
}

func TestStockInDuplicatePostingDoesNotRepeatChanges(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID), models.StockInItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  5,
	})
	service := services.NewStockInService(f.db)

	require.NoError(t, service.Post(context.Background(), uint64(stockIn.ID), uint64(f.user.ID)))
	err := service.Post(context.Background(), uint64(stockIn.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrStockInAlreadyPosted)
	require.Equal(t, int64(15), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockIn", stockIn.ID), 1)
}

func TestStockInMissingLaterStockRollsBackEarlierChanges(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID),
		models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 2},
		models.StockInItem{ProductID: uint64(f.product2.ID), Quantity: 3},
	)

	err := services.NewStockInService(f.db).Post(context.Background(), uint64(stockIn.ID), uint64(f.user.ID))
	require.Error(t, err)
	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockIn", stockIn.ID), 0)
	stillDraft, findErr := repositories.NewStockInRepository(f.db.Query()).FindByID(uint64(stockIn.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
}

func TestStockOutPostingSuccess(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.NoError(t, err)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)
	movements := f.movements(t, "StockOut", stockOut.ID)
	require.Len(t, movements, 1)
	require.Equal(t, "OUT", movements[0].Type)
	require.Equal(t, int64(10), movements[0].BalanceBefore)
	require.Equal(t, int64(6), movements[0].BalanceAfter)
	auditLogs, err := repositories.NewAuditLogRepository(f.db.Query()).FindBySubject("StockOut", uint64(stockOut.ID))
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_OUT", auditLogs[0].Action)
	require.Equal(t, "StockOut", auditLogs[0].SubjectType)
	require.Equal(t, uint64(stockOut.ID), auditLogs[0].SubjectID)
	require.NotNil(t, auditLogs[0].UserID)
	require.Equal(t, uint64(f.user.ID), *auditLogs[0].UserID)
	require.Equal(t, "status=DRAFT", auditLogs[0].OldValues)
	require.Equal(t, "status=POSTED", auditLogs[0].NewValues)
	require.Empty(t, auditLogs[0].IPAddress)
	require.Empty(t, auditLogs[0].UserAgent)
}

func TestStockOutMultiItemNonSortedOrder(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 20)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product2.ID), Quantity: 3},
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 2},
	)

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.NoError(t, err)
	require.Equal(t, int64(8), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(17), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID)).Quantity)

	movements := f.movements(t, "StockOut", stockOut.ID)
	require.Len(t, movements, 2)
	for _, movement := range movements {
		require.Equal(t, "OUT", movement.Type)
	}
	auditLogs, err := f.auditLogs(t, "StockOut", stockOut.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_OUT", auditLogs[0].Action)
}

func TestStockOutInsufficientStockRollsBack(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 3)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrStockOutInsufficientStock)
	require.Equal(t, int64(3), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)
	stillDraft, findErr := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
}

func TestStockOutMultiItemInsufficientStockValidatesBeforeMutation(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 1)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 20)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product2.ID), Quantity: 3},
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 2},
	)

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrStockOutInsufficientStock)
	require.Equal(t, int64(1), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)
	stillDraft, findErr := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}

func TestStockOutRejectsDuplicateProductItem(t *testing.T) {
	f := newInventoryFixture(t)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 1},
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 2},
	)

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrStockOutDuplicateItem)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)
}

func TestStockOutMissingWarehouseStockFailsWithoutCreatingIt(t *testing.T) {
	f := newInventoryFixture(t)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  1,
	})

	err := services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.Error(t, err)
	exists, existsErr := f.db.Query().Model(&models.WarehouseStock{}).
		Where("warehouse_id = ? AND product_id = ?", f.warehouse1.ID, f.product1.ID).
		Exists()
	require.NoError(t, existsErr)
	require.False(t, exists)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)
}

func TestStockOutDuplicatePostingDoesNotRepeatChanges(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})
	service := services.NewStockOutService(f.db)

	require.NoError(t, service.Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID)))
	err := service.Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrStockOutAlreadyPosted)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 1)
	auditLogs, auditErr := f.auditLogs(t, "StockOut", stockOut.ID)
	require.NoError(t, auditErr)
	require.Len(t, auditLogs, 1)
}

func TestStockOutAuditFailureRollsBackTransactionScopedWork(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	actorID := uint64(f.user.ID)
	invalidAuditUserID := uint64(1<<63 - 1)
	err := f.db.Transaction(func(tx contractsorm.Query) error {
		stockOutRepository := repositories.NewStockOutRepository(tx)
		warehouseStockRepository := repositories.NewWarehouseStockRepository(tx)
		stockMovementRepository := repositories.NewStockMovementRepository(tx)
		txAuditService := services.NewAuditLogService(repositories.NewAuditLogRepository(tx))

		lockedStockOut, err := stockOutRepository.FindByIDForUpdate(uint64(stockOut.ID))
		if err != nil {
			return err
		}
		stock, err := warehouseStockRepository.FindByWarehouseAndProductForUpdate(
			lockedStockOut.WarehouseID,
			lockedStockOut.Items[0].ProductID,
		)
		if err != nil {
			return err
		}
		balanceBefore := stock.Quantity
		balanceAfter := balanceBefore - lockedStockOut.Items[0].Quantity
		if err := warehouseStockRepository.UpdateQuantity(stock, balanceAfter); err != nil {
			return err
		}

		referenceID := uint64(lockedStockOut.ID)
		movement := &models.StockMovement{
			WarehouseID:   lockedStockOut.WarehouseID,
			ProductID:     lockedStockOut.Items[0].ProductID,
			Type:          "OUT",
			Quantity:      lockedStockOut.Items[0].Quantity,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			ReferenceType: "StockOut",
			ReferenceID:   &referenceID,
			CreatedBy:     &actorID,
		}
		if err := stockMovementRepository.Create(movement); err != nil {
			return err
		}

		return txAuditService.Create(
			&invalidAuditUserID,
			"POST_STOCK_OUT",
			"StockOut",
			uint64(lockedStockOut.ID),
			"status=DRAFT",
			"status=POSTED",
			"",
			"",
		)
	})
	require.Error(t, err)

	require.Equal(t, int64(10), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 0)
	requireNoAuditLogs(t, f, "StockOut", stockOut.ID)

	stillDraft, findErr := repositories.NewStockOutRepository(f.db.Query()).FindByID(uint64(stockOut.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}

func TestWarehouseTransferPostingSuccess(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 100)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  30,
	})

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.NoError(t, err)
	require.Equal(t, int64(70), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(50), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	posted, err := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, err)
	require.Equal(t, services.StockStatusPosted, posted.Status)
	require.NotNil(t, posted.PostedAt)

	movements := f.movements(t, "WarehouseTransfer", transfer.ID)
	require.Len(t, movements, 2)
	var sourceMovement, destinationMovement *models.StockMovement
	for i := range movements {
		if movements[i].WarehouseID == uint64(f.warehouse1.ID) {
			sourceMovement = &movements[i]
		} else {
			destinationMovement = &movements[i]
		}
	}
	require.NotNil(t, sourceMovement)
	require.NotNil(t, destinationMovement)
	require.Equal(t, "OUT", sourceMovement.Type)
	require.Equal(t, "IN", destinationMovement.Type)
	require.Equal(t, int64(100), sourceMovement.BalanceBefore)
	require.Equal(t, int64(70), sourceMovement.BalanceAfter)
	require.Equal(t, int64(20), destinationMovement.BalanceBefore)
	require.Equal(t, int64(50), destinationMovement.BalanceAfter)
	auditLogs, err := f.auditLogs(t, "WarehouseTransfer", transfer.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	require.Equal(t, "POST_STOCK_TRANSFER", auditLogs[0].Action)
	require.Equal(t, "WarehouseTransfer", auditLogs[0].SubjectType)
	require.Equal(t, uint64(transfer.ID), auditLogs[0].SubjectID)
	require.NotNil(t, auditLogs[0].UserID)
	require.Equal(t, uint64(f.user.ID), *auditLogs[0].UserID)
	require.Equal(t, "status=DRAFT", auditLogs[0].OldValues)
	require.Equal(t, "status=POSTED", auditLogs[0].NewValues)
	require.Empty(t, auditLogs[0].IPAddress)
	require.Empty(t, auditLogs[0].UserAgent)
}

func TestWarehouseTransferInsufficientSourceRollsBack(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 5)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  6,
	})

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrWarehouseTransferInsufficientStock)
	require.Equal(t, int64(5), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
	stillDraft, findErr := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}

func TestWarehouseTransferRejectsSameWarehouse(t *testing.T) {
	f := newInventoryFixture(t)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse1.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  1,
	})

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrWarehouseTransferSameWarehouse)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
	stillDraft, findErr := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
}

func TestWarehouseTransferRejectsDuplicateProductItem(t *testing.T) {
	f := newInventoryFixture(t)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID),
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 1},
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 2},
	)

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrWarehouseTransferDuplicateItem)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
	stillDraft, findErr := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
}

func TestWarehouseTransferDuplicatePostingDoesNotRepeatChanges(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 100)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  30,
	})
	service := services.NewWarehouseTransferService(f.db)

	require.NoError(t, service.Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID)))
	err := service.Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.ErrorIs(t, err, services.ErrWarehouseTransferAlreadyPosted)
	require.Equal(t, int64(70), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(50), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 2)
	auditLogs, auditErr := f.auditLogs(t, "WarehouseTransfer", transfer.ID)
	require.NoError(t, auditErr)
	require.Len(t, auditLogs, 1)
}

func TestWarehouseTransferMissingLaterStockLeavesEarlierStockUnchanged(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 100)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID),
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 30},
		models.WarehouseTransferItem{ProductID: uint64(f.product2.ID), Quantity: 10},
	)

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.Error(t, err)
	require.Equal(t, int64(100), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)
	stillDraft, findErr := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}

func TestWarehouseTransferMultipleItemsCreatesOneAudit(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 100)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 40)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product2.ID), 5)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID),
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 30},
		models.WarehouseTransferItem{ProductID: uint64(f.product2.ID), Quantity: 10},
	)

	err := services.NewWarehouseTransferService(f.db).Post(context.Background(), uint64(transfer.ID), uint64(f.user.ID))
	require.NoError(t, err)
	require.Equal(t, int64(70), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(50), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(30), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID)).Quantity)
	require.Equal(t, int64(15), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product2.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 4)
	auditLogs, err := f.auditLogs(t, "WarehouseTransfer", transfer.ID)
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
}

func TestWarehouseTransferAuditFailureRollsBackTransactionScopedWork(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 100)
	f.addStock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID), 20)
	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID), models.WarehouseTransferItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  30,
	})

	actorID := uint64(f.user.ID)
	invalidAuditUserID := uint64(1<<63 - 1)
	err := f.db.Transaction(func(tx contractsorm.Query) error {
		transferRepository := repositories.NewWarehouseTransferRepository(tx)
		stockRepository := repositories.NewWarehouseStockRepository(tx)
		movementRepository := repositories.NewStockMovementRepository(tx)
		txAuditService := services.NewAuditLogService(repositories.NewAuditLogRepository(tx))

		lockedTransfer, err := transferRepository.FindByIDForUpdate(uint64(transfer.ID))
		if err != nil {
			return err
		}
		sourceStock, err := stockRepository.FindByWarehouseAndProductForUpdate(
			lockedTransfer.FromWarehouseID,
			lockedTransfer.Items[0].ProductID,
		)
		if err != nil {
			return err
		}
		destinationStock, err := stockRepository.FindByWarehouseAndProductForUpdate(
			lockedTransfer.ToWarehouseID,
			lockedTransfer.Items[0].ProductID,
		)
		if err != nil {
			return err
		}

		quantity := lockedTransfer.Items[0].Quantity
		sourceBefore := sourceStock.Quantity
		destinationBefore := destinationStock.Quantity
		sourceAfter := sourceBefore - quantity
		destinationAfter := destinationBefore + quantity
		if err := stockRepository.UpdateQuantity(sourceStock, sourceAfter); err != nil {
			return err
		}
		if err := stockRepository.UpdateQuantity(destinationStock, destinationAfter); err != nil {
			return err
		}

		referenceID := uint64(lockedTransfer.ID)
		sourceMovement := &models.StockMovement{
			WarehouseID:   lockedTransfer.FromWarehouseID,
			ProductID:     lockedTransfer.Items[0].ProductID,
			Type:          "OUT",
			Quantity:      quantity,
			BalanceBefore: sourceBefore,
			BalanceAfter:  sourceAfter,
			ReferenceType: "WarehouseTransfer",
			ReferenceID:   &referenceID,
			CreatedBy:     &actorID,
		}
		if err := movementRepository.Create(sourceMovement); err != nil {
			return err
		}
		destinationMovement := &models.StockMovement{
			WarehouseID:   lockedTransfer.ToWarehouseID,
			ProductID:     lockedTransfer.Items[0].ProductID,
			Type:          "IN",
			Quantity:      quantity,
			BalanceBefore: destinationBefore,
			BalanceAfter:  destinationAfter,
			ReferenceType: "WarehouseTransfer",
			ReferenceID:   &referenceID,
			CreatedBy:     &actorID,
		}
		if err := movementRepository.Create(destinationMovement); err != nil {
			return err
		}

		return txAuditService.Create(
			&invalidAuditUserID,
			"POST_STOCK_TRANSFER",
			"WarehouseTransfer",
			uint64(lockedTransfer.ID),
			"status=DRAFT",
			"status=POSTED",
			"",
			"",
		)
	})
	require.Error(t, err)

	require.Equal(t, int64(100), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(20), f.stock(t, uint64(f.warehouse2.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "WarehouseTransfer", transfer.ID), 0)
	requireNoAuditLogs(t, f, "WarehouseTransfer", transfer.ID)

	stillDraft, findErr := repositories.NewWarehouseTransferRepository(f.db.Query()).FindByID(uint64(transfer.ID))
	require.NoError(t, findErr)
	require.Equal(t, services.StockStatusDraft, stillDraft.Status)
	require.Nil(t, stillDraft.PostedAt)
}

func TestConcurrentMultiItemStockOutsUseDeterministicLockOrder(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 20)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID), 20)
	stockOutA := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 1},
		models.StockOutItem{ProductID: uint64(f.product2.ID), Quantity: 1},
	)
	stockOutB := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product2.ID), Quantity: 1},
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 1},
	)

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, stockOutID := range []uint64{uint64(stockOutA.ID), uint64(stockOutB.ID)} {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			<-start
			errs <- services.NewStockOutService(f.db).Post(context.Background(), id, uint64(f.user.ID))
		}(stockOutID)
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, int64(18), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Equal(t, int64(18), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product2.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOutA.ID), 2)
	require.Len(t, f.movements(t, "StockOut", stockOutB.ID), 2)
	auditA, err := f.auditLogs(t, "StockOut", stockOutA.ID)
	require.NoError(t, err)
	require.Len(t, auditA, 1)
	auditB, err := f.auditLogs(t, "StockOut", stockOutB.ID)
	require.NoError(t, err)
	require.Len(t, auditB, 1)
}

func TestConcurrentStockOutPostingOnlySucceedsOnce(t *testing.T) {
	f := newInventoryFixture(t)
	f.addStock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID), 10)
	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID), models.StockOutItem{
		ProductID: uint64(f.product1.ID),
		Quantity:  4,
	})

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- services.NewStockOutService(f.db).Post(context.Background(), uint64(stockOut.ID), uint64(f.user.ID))
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	var secondError error
	for err := range errs {
		if err == nil {
			successes++
		} else {
			secondError = err
		}
	}
	require.Equal(t, 1, successes)
	require.ErrorIs(t, secondError, services.ErrStockOutAlreadyPosted)
	require.Equal(t, int64(6), f.stock(t, uint64(f.warehouse1.ID), uint64(f.product1.ID)).Quantity)
	require.Len(t, f.movements(t, "StockOut", stockOut.ID), 1)
	auditLogs, auditErr := f.auditLogs(t, "StockOut", stockOut.ID)
	require.NoError(t, auditErr)
	require.Len(t, auditLogs, 1)
}
