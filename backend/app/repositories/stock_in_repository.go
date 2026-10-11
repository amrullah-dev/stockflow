package repositories

import (
	"fmt"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// StockInRepository handles data access for stock-in transactions and their
// items. Transaction boundaries are owned by the service layer and are
// represented by the query supplied when constructing the repository.
type StockInRepository struct {
	query contractsorm.Query
}

func NewStockInRepository(query contractsorm.Query) *StockInRepository {
	return &StockInRepository{query: query}
}

// FindByID returns a stock-in transaction together with its items.
func (r *StockInRepository) FindByID(id uint64) (*models.StockIn, error) {
	stockIn := &models.StockIn{}

	if err := r.query.
		With("Items").
		Where("id = ?", id).
		FindOrFail(stockIn); err != nil {
		return nil, err
	}

	return stockIn, nil
}

// FindAll returns stock-in headers with their supplier and warehouse.
// Items are intentionally not eager-loaded for list responses.
func (r *StockInRepository) FindAll() ([]models.StockIn, error) {
	var stockIns []models.StockIn

	if err := r.query.
		With("Supplier").
		With("Warehouse").
		OrderBy("id", "asc").
		Find(&stockIns); err != nil {
		return nil, err
	}

	return stockIns, nil
}

// FindByIDForRead returns a stock-in with the relationships needed by the
// detail read API. It does not lock the row and does not start a transaction.
func (r *StockInRepository) FindByIDForRead(id uint64) (*models.StockIn, error) {
	stockIn := &models.StockIn{}

	if err := r.query.
		With("Supplier").
		With("Warehouse").
		With("CreatedByUser").
		With("Items.Product").
		Where("id = ?", id).
		FindOrFail(stockIn); err != nil {
		return nil, err
	}

	return stockIn, nil
}

// FindByIDForUpdate returns and locks a stock-in transaction together with its
// items. The supplied query must already be inside a database transaction.
func (r *StockInRepository) FindByIDForUpdate(id uint64) (*models.StockIn, error) {
	var stockIn models.StockIn
	if err := r.query.Where("id = ?", id).First(&stockIn); err != nil {
		fmt.Printf("[DEBUG REPO] StockIN Find Error: %v for ID: %d\n", err, id)
		return nil, err
	}
	var items []models.StockInItem
	if err := r.query.Where("stock_in_id = ?", id).Find(&items); err != nil {
		return nil, err
	}
	stockIn.Items = items
	return &stockIn, nil
}

func (r *StockInRepository) FindByReferenceNumber(referenceNumber string) (*models.StockIn, error) {
	stockIn := &models.StockIn{}

	if err := r.query.
		Where("reference_number = ?", referenceNumber).
		FindOrFail(stockIn); err != nil {
		return nil, err
	}

	return stockIn, nil
}

func (r *StockInRepository) Create(stockIn *models.StockIn) error {
	return r.query.Create(stockIn)
}

func (r *StockInRepository) CreateItem(item *models.StockInItem) error {
	return r.query.Create(item)
}

func (r *StockInRepository) FindItems(stockInID uint64) ([]models.StockInItem, error) {
	var items []models.StockInItem

	if err := r.query.
		Model(&models.StockInItem{}).
		Where("stock_in_id = ?", stockInID).
		Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *StockInRepository) UpdateStatus(stockIn *models.StockIn, status string) error {
	_, err := r.query.Model(stockIn).Update("status", status)
	return err
}

func (r *StockInRepository) UpdatePostedAt(stockIn *models.StockIn, postedAt *time.Time) error {
	_, err := r.query.Model(stockIn).Update("posted_at", postedAt)
	return err
}

func (r *StockInRepository) FindByStatus(status string) ([]models.StockIn, error) {
	var stockIns []models.StockIn

	if err := r.query.
		Where("status = ?", status).
		Find(&stockIns); err != nil {
		return nil, err
	}

	return stockIns, nil
}

func (r *StockInRepository) FindByDateRange(start, end time.Time) ([]models.StockIn, error) {
	var stockIns []models.StockIn

	if err := r.query.
		Where("transaction_date BETWEEN ? AND ?", start, end).
		Find(&stockIns); err != nil {
		return nil, err
	}

	return stockIns, nil
}

// UpdateStockInHeaderData contains explicitly mutable header fields for a Stock IN transaction.
type UpdateStockInHeaderData struct {
	TransactionDate time.Time
	SupplierID      uint64
	WarehouseID     uint64
	Note            string
}

// UpdateHeader updates only the explicitly mutable fields of a stock-in header.
// Fields like id, reference_number, created_by, status, and posted_at are never modified.
func (r *StockInRepository) UpdateHeader(stockIn *models.StockIn, data UpdateStockInHeaderData) error {
	_, err := r.query.Model(stockIn).Update(map[string]any{
		"transaction_date": data.TransactionDate,
		"supplier_id":      data.SupplierID,
		"warehouse_id":     data.WarehouseID,
		"note":             data.Note,
	})
	if err != nil {
		return err
	}

	stockIn.TransactionDate = data.TransactionDate
	stockIn.SupplierID = data.SupplierID
	stockIn.WarehouseID = data.WarehouseID
	stockIn.Note = data.Note
	return nil
}

// DeleteItemsByStockInID deletes all line items associated with a stock-in transaction.
func (r *StockInRepository) DeleteItemsByStockInID(stockInID uint64) error {
	_, err := r.query.Where("stock_in_id = ?", stockInID).Delete(&models.StockInItem{})
	return err
}

// DeleteByID deletes a stock-in header by its ID.
func (r *StockInRepository) DeleteByID(id uint64) error {
	_, err := r.query.Where("id = ?", id).Delete(&models.StockIn{})
	return err
}
