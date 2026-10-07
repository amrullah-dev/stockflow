package repositories

import (
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// StockOutRepository handles data access for stock-out transactions and their
// items. Transaction boundaries are owned by the service layer and are
// represented by the query supplied when constructing the repository.
type StockOutRepository struct {
	query contractsorm.Query
}

func NewStockOutRepository(query contractsorm.Query) *StockOutRepository {
	return &StockOutRepository{query: query}
}

// FindByID returns a stock-out transaction together with its items.
func (r *StockOutRepository) FindByID(id uint64) (*models.StockOut, error) {
	stockOut := &models.StockOut{}

	if err := r.query.
		With("Items").
		Where("id = ?", id).
		FindOrFail(stockOut); err != nil {
		return nil, err
	}

	return stockOut, nil
}

// FindAll returns stock-out headers with their warehouses. Items are not
// eager-loaded for list responses.
func (r *StockOutRepository) FindAll() ([]models.StockOut, error) {
	var stockOuts []models.StockOut

	if err := r.query.
		With("Warehouse").
		OrderBy("id", "asc").
		Find(&stockOuts); err != nil {
		return nil, err
	}

	return stockOuts, nil
}

// FindByIDForRead returns a stock-out with the relationships needed by the
// detail read API. It does not lock the row or start a transaction.
func (r *StockOutRepository) FindByIDForRead(id uint64) (*models.StockOut, error) {
	stockOut := &models.StockOut{}

	if err := r.query.
		With("Warehouse").
		With("CreatedByUser").
		With("Items.Product").
		Where("id = ?", id).
		FindOrFail(stockOut); err != nil {
		return nil, err
	}

	return stockOut, nil
}

// FindByIDForUpdate returns and locks a stock-out transaction together with
// its items. The supplied query must already be inside a database transaction.
func (r *StockOutRepository) FindByIDForUpdate(id uint64) (*models.StockOut, error) {
	stockOut := &models.StockOut{}

	if err := r.query.
		With("Items").
		Where("id = ?", id).
		LockForUpdate().
		FindOrFail(stockOut); err != nil {
		return nil, err
	}

	return stockOut, nil
}

func (r *StockOutRepository) FindByReferenceNumber(referenceNumber string) (*models.StockOut, error) {
	stockOut := &models.StockOut{}

	if err := r.query.
		Where("reference_number = ?", referenceNumber).
		FindOrFail(stockOut); err != nil {
		return nil, err
	}

	return stockOut, nil
}

func (r *StockOutRepository) Create(stockOut *models.StockOut) error {
	return r.query.Create(stockOut)
}

func (r *StockOutRepository) CreateItem(item *models.StockOutItem) error {
	return r.query.Create(item)
}

func (r *StockOutRepository) FindItems(stockOutID uint64) ([]models.StockOutItem, error) {
	var items []models.StockOutItem

	if err := r.query.
		Where("stock_out_id = ?", stockOutID).
		Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *StockOutRepository) UpdateStatus(stockOut *models.StockOut, status string) error {
	_, err := r.query.Model(stockOut).Update("status", status)
	return err
}

func (r *StockOutRepository) UpdatePostedAt(stockOut *models.StockOut, postedAt *time.Time) error {
	_, err := r.query.Model(stockOut).Update("posted_at", postedAt)
	return err
}

func (r *StockOutRepository) FindByStatus(status string) ([]models.StockOut, error) {
	var stockOuts []models.StockOut

	if err := r.query.
		Where("status = ?", status).
		Find(&stockOuts); err != nil {
		return nil, err
	}

	return stockOuts, nil
}

func (r *StockOutRepository) FindByDateRange(start, end time.Time) ([]models.StockOut, error) {
	var stockOuts []models.StockOut

	if err := r.query.
		Where("transaction_date BETWEEN ? AND ?", start, end).
		Find(&stockOuts); err != nil {
		return nil, err
	}

	return stockOuts, nil
}

// UpdateStockOutHeaderData contains explicitly mutable header fields for a Stock OUT transaction.
type UpdateStockOutHeaderData struct {
	TransactionDate time.Time
	WarehouseID     uint64
	Destination     string
	Note            string
}

// UpdateHeader updates only the explicitly mutable fields of a stock-out header.
// Fields like id, reference_number, created_by, status, and posted_at are never modified.
func (r *StockOutRepository) UpdateHeader(stockOut *models.StockOut, data UpdateStockOutHeaderData) error {
	_, err := r.query.Model(stockOut).Update(map[string]any{
		"transaction_date": data.TransactionDate,
		"warehouse_id":     data.WarehouseID,
		"destination":      data.Destination,
		"note":             data.Note,
	})
	if err != nil {
		return err
	}

	stockOut.TransactionDate = data.TransactionDate
	stockOut.WarehouseID = data.WarehouseID
	stockOut.Destination = data.Destination
	stockOut.Note = data.Note
	return nil
}

// DeleteItemsByStockOutID deletes all line items associated with a stock-out transaction.
func (r *StockOutRepository) DeleteItemsByStockOutID(stockOutID uint64) error {
	_, err := r.query.Where("stock_out_id = ?", stockOutID).Delete(&models.StockOutItem{})
	return err
}

// DeleteByID deletes a stock-out header by its ID.
func (r *StockOutRepository) DeleteByID(id uint64) error {
	_, err := r.query.Where("id = ?", id).Delete(&models.StockOut{})
	return err
}
