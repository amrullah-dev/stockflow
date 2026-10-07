package feature

import (
	"errors"
	"testing"
	"time"

	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/app/repositories"
)

func TestProductRepositoryFindByID(t *testing.T) {
	f := newInventoryFixture(t)
	repo := repositories.NewProductRepository(f.db.Query())

	// 1. Existing product
	product, err := repo.FindByID(uint64(f.product1.ID))
	require.NoError(t, err)
	require.NotNil(t, product)
	require.Equal(t, f.product1.ID, product.ID)
	require.Equal(t, f.product1.Code, product.Code)

	// 2. Missing product
	missing, err := repo.FindByID(999999999)
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, missing)

	// 3. Excludes soft-deleted product
	_, err = f.db.Query().Delete(&f.product2)
	require.NoError(t, err)

	softDeleted, err := repo.FindByID(uint64(f.product2.ID))
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, softDeleted)
}

func TestSupplierRepositoryFindByID(t *testing.T) {
	f := newInventoryFixture(t)
	repo := repositories.NewSupplierRepository(f.db.Query())

	// 1. Existing supplier
	supplier, err := repo.FindByID(uint64(f.supplier.ID))
	require.NoError(t, err)
	require.NotNil(t, supplier)
	require.Equal(t, f.supplier.ID, supplier.ID)
	require.Equal(t, f.supplier.Code, supplier.Code)

	// 2. Missing supplier
	missing, err := repo.FindByID(999999999)
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, missing)

	// 3. Excludes soft-deleted supplier
	_, err = f.db.Query().Delete(&f.supplier)
	require.NoError(t, err)

	softDeleted, err := repo.FindByID(uint64(f.supplier.ID))
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, softDeleted)
}

func TestStockInRepositoryDraftFoundationMethods(t *testing.T) {
	f := newInventoryFixture(t)
	repo := repositories.NewStockInRepository(f.db.Query())

	stockIn := f.createStockIn(t, uint64(f.warehouse1.ID),
		models.StockInItem{ProductID: uint64(f.product1.ID), Quantity: 5, UnitPrice: 10, Subtotal: 50},
		models.StockInItem{ProductID: uint64(f.product2.ID), Quantity: 2, UnitPrice: 20, Subtotal: 40},
	)

	originalRef := stockIn.ReferenceNumber
	originalCreatedBy := stockIn.CreatedBy
	originalStatus := stockIn.Status
	newDate := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	newNote := "Updated draft note"

	// 1. UpdateHeader
	err := repo.UpdateHeader(stockIn, repositories.UpdateStockInHeaderData{
		TransactionDate: newDate,
		SupplierID:      uint64(f.supplier.ID),
		WarehouseID:     uint64(f.warehouse2.ID),
		Note:            newNote,
	})
	require.NoError(t, err)

	// Reload from DB to confirm only mutable fields changed
	reloaded, err := repo.FindByID(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Equal(t, uint64(f.warehouse2.ID), reloaded.WarehouseID)
	require.Equal(t, newNote, reloaded.Note)
	require.Equal(t, originalRef, reloaded.ReferenceNumber)
	require.Equal(t, originalCreatedBy, reloaded.CreatedBy)
	require.Equal(t, originalStatus, reloaded.Status)
	require.Nil(t, reloaded.PostedAt)

	// 2. DeleteItemsByStockInID
	err = repo.DeleteItemsByStockInID(uint64(stockIn.ID))
	require.NoError(t, err)

	items, err := repo.FindItems(uint64(stockIn.ID))
	require.NoError(t, err)
	require.Empty(t, items)

	// 3. DeleteByID
	err = repo.DeleteByID(uint64(stockIn.ID))
	require.NoError(t, err)

	deleted, err := repo.FindByID(uint64(stockIn.ID))
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, deleted)
}

func TestStockOutRepositoryDraftFoundationMethods(t *testing.T) {
	f := newInventoryFixture(t)
	repo := repositories.NewStockOutRepository(f.db.Query())

	stockOut := f.createStockOut(t, uint64(f.warehouse1.ID),
		models.StockOutItem{ProductID: uint64(f.product1.ID), Quantity: 3},
	)

	originalRef := stockOut.ReferenceNumber
	originalCreatedBy := stockOut.CreatedBy
	originalStatus := stockOut.Status
	newDate := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	newDest := "Retail Store Branch A"
	newNote := "Updated stock-out note"

	// 1. UpdateHeader
	err := repo.UpdateHeader(stockOut, repositories.UpdateStockOutHeaderData{
		TransactionDate: newDate,
		WarehouseID:     uint64(f.warehouse2.ID),
		Destination:     newDest,
		Note:            newNote,
	})
	require.NoError(t, err)

	// Reload from DB
	reloaded, err := repo.FindByID(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Equal(t, uint64(f.warehouse2.ID), reloaded.WarehouseID)
	require.Equal(t, newDest, reloaded.Destination)
	require.Equal(t, newNote, reloaded.Note)
	require.Equal(t, originalRef, reloaded.ReferenceNumber)
	require.Equal(t, originalCreatedBy, reloaded.CreatedBy)
	require.Equal(t, originalStatus, reloaded.Status)
	require.Nil(t, reloaded.PostedAt)

	// 2. DeleteItemsByStockOutID
	err = repo.DeleteItemsByStockOutID(uint64(stockOut.ID))
	require.NoError(t, err)

	items, err := repo.FindItems(uint64(stockOut.ID))
	require.NoError(t, err)
	require.Empty(t, items)

	// 3. DeleteByID
	err = repo.DeleteByID(uint64(stockOut.ID))
	require.NoError(t, err)

	deleted, err := repo.FindByID(uint64(stockOut.ID))
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, deleted)
}

func TestWarehouseTransferRepositoryDraftFoundationMethods(t *testing.T) {
	f := newInventoryFixture(t)
	repo := repositories.NewWarehouseTransferRepository(f.db.Query())

	transfer := f.createTransfer(t, uint64(f.warehouse1.ID), uint64(f.warehouse2.ID),
		models.WarehouseTransferItem{ProductID: uint64(f.product1.ID), Quantity: 4},
	)

	originalRef := transfer.ReferenceNumber
	originalCreatedBy := transfer.CreatedBy
	originalStatus := transfer.Status
	newDate := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	newNote := "Updated transfer note"

	// 1. UpdateHeader
	err := repo.UpdateHeader(transfer, repositories.UpdateWarehouseTransferHeaderData{
		TransactionDate: newDate,
		FromWarehouseID: uint64(f.warehouse2.ID),
		ToWarehouseID:   uint64(f.warehouse1.ID),
		Note:            newNote,
	})
	require.NoError(t, err)

	// Reload from DB
	reloaded, err := repo.FindByID(uint64(transfer.ID))
	require.NoError(t, err)
	require.Equal(t, uint64(f.warehouse2.ID), reloaded.FromWarehouseID)
	require.Equal(t, uint64(f.warehouse1.ID), reloaded.ToWarehouseID)
	require.Equal(t, newNote, reloaded.Note)
	require.Equal(t, originalRef, reloaded.ReferenceNumber)
	require.Equal(t, originalCreatedBy, reloaded.CreatedBy)
	require.Equal(t, originalStatus, reloaded.Status)
	require.Nil(t, reloaded.PostedAt)

	// 2. DeleteItemsByTransferID
	err = repo.DeleteItemsByTransferID(uint64(transfer.ID))
	require.NoError(t, err)

	items, err := repo.FindItems(uint64(transfer.ID))
	require.NoError(t, err)
	require.Empty(t, items)

	// 3. DeleteByID
	err = repo.DeleteByID(uint64(transfer.ID))
	require.NoError(t, err)

	deleted, err := repo.FindByID(uint64(transfer.ID))
	require.Error(t, err)
	require.True(t, errors.Is(err, frameworkerrors.OrmRecordNotFound))
	require.Nil(t, deleted)
}
