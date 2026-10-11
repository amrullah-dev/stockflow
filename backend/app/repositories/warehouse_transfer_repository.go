package repositories

import (
	"fmt"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// WarehouseTransferRepository handles data access for warehouse transfer
// transactions and their items. Transaction boundaries are owned by the
// service layer and are represented by the query supplied when constructing
// the repository.
type WarehouseTransferRepository struct {
	query contractsorm.Query
}

func NewWarehouseTransferRepository(query contractsorm.Query) *WarehouseTransferRepository {
	return &WarehouseTransferRepository{query: query}
}

// FindByID returns a warehouse transfer together with its items.
func (r *WarehouseTransferRepository) FindByID(id uint64) (*models.WarehouseTransfer, error) {
	transfer := &models.WarehouseTransfer{}

	if err := r.query.
		With("Items").
		Where("id = ?", id).
		FindOrFail(transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// FindAll returns transfer headers with their source and destination
// warehouses. Items and creator details are intentionally omitted from list
// responses.
func (r *WarehouseTransferRepository) FindAll() ([]models.WarehouseTransfer, error) {
	var transfers []models.WarehouseTransfer

	if err := r.query.
		With("FromWarehouse").
		With("ToWarehouse").
		OrderBy("id", "asc").
		Find(&transfers); err != nil {
		return nil, err
	}

	return transfers, nil
}

// FindByIDForRead returns a transfer with the relationships needed by the
// detail read API. It does not lock the row or start a transaction.
func (r *WarehouseTransferRepository) FindByIDForRead(id uint64) (*models.WarehouseTransfer, error) {
	transfer := &models.WarehouseTransfer{}

	if err := r.query.
		With("FromWarehouse").
		With("ToWarehouse").
		With("CreatedByUser").
		With("Items.Product").
		Where("id = ?", id).
		FindOrFail(transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// FindByIDForUpdate returns and locks a warehouse transfer together with its
// items. The supplied query must already be inside a database transaction.
func (r *WarehouseTransferRepository) FindByIDForUpdate(id uint64) (*models.WarehouseTransfer, error) {
	var transfer models.WarehouseTransfer
	if err := r.query.Where("id = ?", id).First(&transfer); err != nil {
		fmt.Printf("[DEBUG REPO] Transfer Find Error: %v for ID: %d\n", err, id)
		return nil, err
	}
	var items []models.WarehouseTransferItem
	if err := r.query.Where("transfer_id = ?", id).Find(&items); err != nil {
		return nil, err
	}
	transfer.Items = items
	return &transfer, nil
}

func (r *WarehouseTransferRepository) FindByReferenceNumber(referenceNumber string) (*models.WarehouseTransfer, error) {
	transfer := &models.WarehouseTransfer{}

	if err := r.query.
		Where("reference_number = ?", referenceNumber).
		FindOrFail(transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

func (r *WarehouseTransferRepository) Create(transfer *models.WarehouseTransfer) error {
	return r.query.Create(transfer)
}

func (r *WarehouseTransferRepository) CreateItem(item *models.WarehouseTransferItem) error {
	return r.query.Create(item)
}

func (r *WarehouseTransferRepository) FindItems(transferID uint64) ([]models.WarehouseTransferItem, error) {
	var items []models.WarehouseTransferItem

	if err := r.query.
		Model(&models.WarehouseTransferItem{}).
		Where("transfer_id = ?", transferID).
		Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *WarehouseTransferRepository) UpdateStatus(transfer *models.WarehouseTransfer, status string) error {
	_, err := r.query.Model(transfer).Update("status", status)
	return err
}

func (r *WarehouseTransferRepository) UpdatePostedAt(transfer *models.WarehouseTransfer, postedAt *time.Time) error {
	_, err := r.query.Model(transfer).Update("posted_at", postedAt)
	return err
}

func (r *WarehouseTransferRepository) FindByStatus(status string) ([]models.WarehouseTransfer, error) {
	var transfers []models.WarehouseTransfer

	if err := r.query.
		Where("status = ?", status).
		Find(&transfers); err != nil {
		return nil, err
	}

	return transfers, nil
}

func (r *WarehouseTransferRepository) FindByDateRange(start, end time.Time) ([]models.WarehouseTransfer, error) {
	var transfers []models.WarehouseTransfer

	if err := r.query.
		Where("transaction_date BETWEEN ? AND ?", start, end).
		Find(&transfers); err != nil {
		return nil, err
	}

	return transfers, nil
}

// UpdateWarehouseTransferHeaderData contains explicitly mutable header fields for a Warehouse Transfer transaction.
type UpdateWarehouseTransferHeaderData struct {
	TransactionDate time.Time
	FromWarehouseID uint64
	ToWarehouseID   uint64
	Note            string
}

// UpdateHeader updates only the explicitly mutable fields of a warehouse transfer header.
// Fields like id, reference_number, created_by, status, and posted_at are never modified.
func (r *WarehouseTransferRepository) UpdateHeader(transfer *models.WarehouseTransfer, data UpdateWarehouseTransferHeaderData) error {
	_, err := r.query.Model(transfer).Update(map[string]any{
		"transaction_date":  data.TransactionDate,
		"from_warehouse_id": data.FromWarehouseID,
		"to_warehouse_id":   data.ToWarehouseID,
		"note":              data.Note,
	})
	if err != nil {
		return err
	}

	transfer.TransactionDate = data.TransactionDate
	transfer.FromWarehouseID = data.FromWarehouseID
	transfer.ToWarehouseID = data.ToWarehouseID
	transfer.Note = data.Note
	return nil
}

// DeleteItemsByTransferID deletes all line items associated with a warehouse transfer transaction.
func (r *WarehouseTransferRepository) DeleteItemsByTransferID(transferID uint64) error {
	_, err := r.query.Where("transfer_id = ?", transferID).Delete(&models.WarehouseTransferItem{})
	return err
}

// DeleteByID deletes a warehouse transfer header by its ID.
func (r *WarehouseTransferRepository) DeleteByID(id uint64) error {
	_, err := r.query.Where("id = ?", id).Delete(&models.WarehouseTransfer{})
	return err
}
