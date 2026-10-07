package seeders

import (
	"context"
	"fmt"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services"
)

const (
	developmentRoleName  = "Admin"
	developmentRoleDesc  = "Administrator StockFlow"
	developmentUserName  = "Administrator"
	developmentUserEmail = "admin@stockflow.local"
	developmentPassword  = "Admin123!"
	developmentMarker    = "[DEV-SEED]"
)

type developmentCategory struct {
	Name        string
	Description string
	IsActive    bool
}

type developmentSupplier struct {
	Code     string
	Name     string
	Phone    string
	Email    string
	Address  string
	IsActive bool
}

type developmentProduct struct {
	Code         string
	Name         string
	Description  string
	Unit         string
	MinimumStock int64
	CategoryName string
	IsActive     bool
}

type developmentWarehouse struct {
	Code        string
	Name        string
	Location    string
	Description string
	IsActive    bool
}

type developmentStockIn struct {
	Note          string
	SupplierCode  string
	WarehouseCode string
	Date          string
	Posted        bool
	Items         []services.CreateStockInItemInput
}

type developmentStockOut struct {
	Note          string
	WarehouseCode string
	Date          string
	Destination   string
	Posted        bool
	Items         []services.CreateStockOutItemInput
}

type developmentTransfer struct {
	Note              string
	FromWarehouseCode string
	ToWarehouseCode   string
	Date              string
	Posted            bool
	Items             []services.CreateWarehouseTransferItemInput
}

var developmentCategories = []developmentCategory{
	{Name: "Electronics", Description: "Produk elektronik", IsActive: true},
	{Name: "Office Supplies", Description: "Perlengkapan kantor", IsActive: true},
	{Name: "Furniture", Description: "Perabot kantor", IsActive: true},
	{Name: "Networking", Description: "Perangkat jaringan", IsActive: true},
}

var developmentSuppliers = []developmentSupplier{
	{Code: "SUP001", Name: "PT Sumber Jaya", Phone: "081234567001", Email: "supplier1@stockflow.local", Address: "Jl. Contoh No. 1", IsActive: true},
	{Code: "SUP002", Name: "CV Maju Bersama", Phone: "081234567002", Email: "supplier2@stockflow.local", Address: "Jl. Contoh No. 2", IsActive: true},
	{Code: "SUP003", Name: "PT Nusantara Supply", Phone: "081234567003", Email: "supplier3@stockflow.local", Address: "Jl. Contoh No. 3", IsActive: true},
	{Code: "SUP004", Name: "CV Sentosa Abadi", Phone: "081234567004", Email: "supplier4@stockflow.local", Address: "Jl. Contoh No. 4", IsActive: true},
}

var developmentProducts = []developmentProduct{
	{Code: "PRD001", Name: "Keyboard Mechanical", Description: "Keyboard mechanical untuk kebutuhan komputer", Unit: "pcs", MinimumStock: 10, CategoryName: "Electronics", IsActive: true},
	{Code: "PRD002", Name: "Mouse Wireless", Description: "Mouse wireless untuk kebutuhan komputer", Unit: "pcs", MinimumStock: 10, CategoryName: "Electronics", IsActive: true},
	{Code: "PRD003", Name: "Monitor 24 Inch", Description: "Monitor 24 inch untuk kebutuhan komputer", Unit: "unit", MinimumStock: 5, CategoryName: "Electronics", IsActive: true},
	{Code: "PRD004", Name: "Laptop Stand", Description: "Laptop stand ergonomis", Unit: "pcs", MinimumStock: 5, CategoryName: "Furniture", IsActive: true},
	{Code: "PRD005", Name: "Kertas A4", Description: "Kertas A4 untuk kebutuhan administrasi", Unit: "rim", MinimumStock: 20, CategoryName: "Office Supplies", IsActive: true},
	{Code: "PRD006", Name: "Pulpen Gel", Description: "Pulpen gel untuk kebutuhan kantor", Unit: "box", MinimumStock: 10, CategoryName: "Office Supplies", IsActive: true},
	{Code: "PRD007", Name: "Buku Tulis", Description: "Buku tulis untuk kebutuhan administrasi", Unit: "pcs", MinimumStock: 15, CategoryName: "Office Supplies", IsActive: true},
	{Code: "PRD008", Name: "Router WiFi", Description: "Router WiFi untuk jaringan kantor", Unit: "unit", MinimumStock: 5, CategoryName: "Networking", IsActive: true},
	{Code: "PRD009", Name: "Kabel LAN", Description: "Kabel LAN Cat6", Unit: "pcs", MinimumStock: 30, CategoryName: "Networking", IsActive: true},
	{Code: "PRD010", Name: "Switch 8 Port", Description: "Switch jaringan 8 port", Unit: "unit", MinimumStock: 3, CategoryName: "Networking", IsActive: true},
}

var developmentWarehouses = []developmentWarehouse{
	{Code: "WH001", Name: "Gudang Utama", Location: "Bandung", Description: "Gudang utama StockFlow", IsActive: true},
	{Code: "WH002", Name: "Gudang Cabang", Location: "Cimahi", Description: "Gudang cabang StockFlow", IsActive: true},
}

var developmentStockIns = []developmentStockIn{
	{Note: developmentMarker + " STOCK-IN-001", SupplierCode: "SUP001", WarehouseCode: "WH001", Date: "2026-10-01", Posted: true, Items: []services.CreateStockInItemInput{{ProductID: 1, Quantity: 40, UnitPrice: 100000}, {ProductID: 2, Quantity: 60, UnitPrice: 50000}}},
	{Note: developmentMarker + " STOCK-IN-002", SupplierCode: "SUP002", WarehouseCode: "WH001", Date: "2026-10-02", Posted: true, Items: []services.CreateStockInItemInput{{ProductID: 3, Quantity: 25, UnitPrice: 2500000}, {ProductID: 5, Quantity: 50, UnitPrice: 70000}}},
	{Note: developmentMarker + " STOCK-IN-003", SupplierCode: "SUP003", WarehouseCode: "WH001", Date: "2026-10-03", Posted: true, Items: []services.CreateStockInItemInput{{ProductID: 8, Quantity: 15, UnitPrice: 800000}, {ProductID: 9, Quantity: 120, UnitPrice: 15000}}},
	{Note: developmentMarker + " STOCK-IN-004", SupplierCode: "SUP001", WarehouseCode: "WH002", Date: "2026-10-04", Posted: true, Items: []services.CreateStockInItemInput{{ProductID: 1, Quantity: 20, UnitPrice: 100000}, {ProductID: 2, Quantity: 20, UnitPrice: 50000}, {ProductID: 7, Quantity: 50, UnitPrice: 30000}}},
	{Note: developmentMarker + " STOCK-IN-005", SupplierCode: "SUP004", WarehouseCode: "WH002", Date: "2026-10-05", Posted: true, Items: []services.CreateStockInItemInput{{ProductID: 5, Quantity: 40, UnitPrice: 70000}, {ProductID: 8, Quantity: 10, UnitPrice: 800000}, {ProductID: 9, Quantity: 70, UnitPrice: 15000}}},
	{Note: developmentMarker + " STOCK-IN-DRAFT-001", SupplierCode: "SUP002", WarehouseCode: "WH001", Date: "2026-10-06", Posted: false, Items: []services.CreateStockInItemInput{{ProductID: 4, Quantity: 5, UnitPrice: 450000}}},
	{Note: developmentMarker + " STOCK-IN-DRAFT-002", SupplierCode: "SUP003", WarehouseCode: "WH002", Date: "2026-10-06", Posted: false, Items: []services.CreateStockInItemInput{{ProductID: 6, Quantity: 12, UnitPrice: 25000}}},
	{Note: developmentMarker + " STOCK-IN-DRAFT-003", SupplierCode: "SUP004", WarehouseCode: "WH001", Date: "2026-10-07", Posted: false, Items: []services.CreateStockInItemInput{{ProductID: 10, Quantity: 3, UnitPrice: 650000}}},
}

var developmentStockOuts = []developmentStockOut{
	{Note: developmentMarker + " STOCK-OUT-001", WarehouseCode: "WH001", Date: "2026-10-08", Destination: "Tim IT", Posted: true, Items: []services.CreateStockOutItemInput{{ProductID: 1, Quantity: 10}}},
	{Note: developmentMarker + " STOCK-OUT-002", WarehouseCode: "WH001", Date: "2026-10-08", Destination: "Tim Operasional", Posted: true, Items: []services.CreateStockOutItemInput{{ProductID: 2, Quantity: 20}}},
	{Note: developmentMarker + " STOCK-OUT-003", WarehouseCode: "WH001", Date: "2026-10-08", Destination: "Ruang Meeting", Posted: true, Items: []services.CreateStockOutItemInput{{ProductID: 3, Quantity: 10}}},
	{Note: developmentMarker + " STOCK-OUT-004", WarehouseCode: "WH002", Date: "2026-10-08", Destination: "Cabang Cimahi", Posted: true, Items: []services.CreateStockOutItemInput{{ProductID: 1, Quantity: 15}}},
	{Note: developmentMarker + " STOCK-OUT-005", WarehouseCode: "WH002", Date: "2026-10-08", Destination: "Administrasi", Posted: true, Items: []services.CreateStockOutItemInput{{ProductID: 5, Quantity: 15}}},
	{Note: developmentMarker + " STOCK-OUT-DRAFT-001", WarehouseCode: "WH001", Date: "2026-10-08", Destination: "Perencanaan", Posted: false, Items: []services.CreateStockOutItemInput{{ProductID: 4, Quantity: 2}}},
	{Note: developmentMarker + " STOCK-OUT-DRAFT-002", WarehouseCode: "WH002", Date: "2026-10-08", Destination: "Perencanaan", Posted: false, Items: []services.CreateStockOutItemInput{{ProductID: 6, Quantity: 4}}},
	{Note: developmentMarker + " STOCK-OUT-DRAFT-003", WarehouseCode: "WH001", Date: "2026-10-08", Destination: "Perencanaan", Posted: false, Items: []services.CreateStockOutItemInput{{ProductID: 10, Quantity: 1}}},
}

var developmentTransfers = []developmentTransfer{
	{Note: developmentMarker + " TRANSFER-001", FromWarehouseCode: "WH001", ToWarehouseCode: "WH002", Date: "2026-10-08", Posted: true, Items: []services.CreateWarehouseTransferItemInput{{ProductID: 1, Quantity: 5}}},
	{Note: developmentMarker + " TRANSFER-002", FromWarehouseCode: "WH001", ToWarehouseCode: "WH002", Date: "2026-10-08", Posted: true, Items: []services.CreateWarehouseTransferItemInput{{ProductID: 9, Quantity: 20}}},
	{Note: developmentMarker + " TRANSFER-003", FromWarehouseCode: "WH002", ToWarehouseCode: "WH001", Date: "2026-10-08", Posted: true, Items: []services.CreateWarehouseTransferItemInput{{ProductID: 8, Quantity: 5}}},
	{Note: developmentMarker + " TRANSFER-DRAFT-001", FromWarehouseCode: "WH001", ToWarehouseCode: "WH002", Date: "2026-10-08", Posted: false, Items: []services.CreateWarehouseTransferItemInput{{ProductID: 4, Quantity: 2}}},
	{Note: developmentMarker + " TRANSFER-DRAFT-002", FromWarehouseCode: "WH002", ToWarehouseCode: "WH001", Date: "2026-10-08", Posted: false, Items: []services.CreateWarehouseTransferItemInput{{ProductID: 6, Quantity: 3}}},
}

type DevelopmentSeeder struct{}

func (s *DevelopmentSeeder) Signature() string { return "development" }

func (s *DevelopmentSeeder) Run() error {
	role, user, categories, suppliers, products, warehouses, err := ensureDevelopmentMasters()
	if err != nil {
		return err
	}

	if err := ensureDevelopmentStocks(products, warehouses); err != nil {
		return err
	}

	ctx := context.Background()
	stockInService := services.NewStockInService(facades.Orm())
	stockOutService := services.NewStockOutService(facades.Orm())
	transferService := services.NewWarehouseTransferService(facades.Orm())

	if err := ensureDevelopmentStockIns(ctx, stockInService, uint64(user.ID), suppliers, warehouses, products); err != nil {
		return err
	}
	if err := ensureDevelopmentStockOuts(ctx, stockOutService, uint64(user.ID), warehouses, products); err != nil {
		return err
	}
	if err := ensureDevelopmentTransfers(ctx, transferService, uint64(user.ID), warehouses, products); err != nil {
		return err
	}

	_ = role
	_ = categories
	return nil
}

func ensureDevelopmentMasters() (*models.Role, *models.User, map[string]*models.Category, map[string]*models.Supplier, map[string]*models.Product, map[string]*models.Warehouse, error) {
	var role *models.Role
	var user *models.User
	var categories map[string]*models.Category
	var suppliers map[string]*models.Supplier
	var products map[string]*models.Product
	var warehouses map[string]*models.Warehouse
	err := facades.Orm().Transaction(func(tx contractsorm.Query) error {
		var err error
		role, err = ensureDevelopmentRole(tx)
		if err != nil {
			return err
		}
		user, err = ensureDevelopmentUser(tx, role)
		if err != nil {
			return err
		}
		categories, err = ensureDevelopmentCategories(tx)
		if err != nil {
			return err
		}
		suppliers, err = ensureDevelopmentSuppliers(tx)
		if err != nil {
			return err
		}
		products, err = ensureDevelopmentProducts(tx, categories)
		if err != nil {
			return err
		}
		warehouses, err = ensureDevelopmentWarehouses(tx)
		return err
	})
	return role, user, categories, suppliers, products, warehouses, err
}

func ensureDevelopmentRole(tx contractsorm.Query) (*models.Role, error) {
	roles := make([]models.Role, 0, 1)
	if err := tx.Where("name = ?", developmentRoleName).Find(&roles); err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		role := &models.Role{Name: developmentRoleName, Description: developmentRoleDesc}
		return role, tx.Create(role)
	}
	role := &roles[0]
	_, err := tx.Model(role).Update(map[string]any{"description": developmentRoleDesc})
	return role, err
}

func ensureDevelopmentUser(tx contractsorm.Query, role *models.Role) (*models.User, error) {
	users := make([]models.User, 0, 1)
	if err := tx.WithTrashed().Where("email = ?", developmentUserEmail).Find(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		hashedPassword, err := facades.Hash().Make(developmentPassword)
		if err != nil {
			return nil, err
		}
		user := &models.User{RoleID: uint64(role.ID), Name: developmentUserName, Email: developmentUserEmail, Password: hashedPassword, IsActive: true}
		return user, tx.Create(user)
	}
	user := &users[0]
	if user.DeletedAt.Valid {
		if _, err := tx.Restore(user); err != nil {
			return nil, err
		}
	}
	_, err := tx.Model(user).Update(map[string]any{"role_id": uint64(role.ID), "name": developmentUserName, "is_active": true})
	return user, err
}

func ensureDevelopmentCategories(tx contractsorm.Query) (map[string]*models.Category, error) {
	result := make(map[string]*models.Category, len(developmentCategories))
	for _, item := range developmentCategories {
		rows := make([]models.Category, 0, 1)
		if err := tx.WithTrashed().Where("name = ?", item.Name).Find(&rows); err != nil {
			return nil, err
		}
		var category *models.Category
		if len(rows) == 0 {
			category = &models.Category{Name: item.Name, Description: item.Description, IsActive: item.IsActive}
			if err := tx.Create(category); err != nil {
				return nil, err
			}
		} else {
			category = &rows[0]
			if category.DeletedAt.Valid {
				if _, err := tx.Restore(category); err != nil {
					return nil, err
				}
			}
			if _, err := tx.Model(category).Update(map[string]any{"description": item.Description, "is_active": item.IsActive}); err != nil {
				return nil, err
			}
		}
		result[item.Name] = category
	}
	return result, nil
}

func ensureDevelopmentSuppliers(tx contractsorm.Query) (map[string]*models.Supplier, error) {
	result := make(map[string]*models.Supplier, len(developmentSuppliers))
	for _, item := range developmentSuppliers {
		rows := make([]models.Supplier, 0, 1)
		if err := tx.WithTrashed().Where("code = ?", item.Code).Find(&rows); err != nil {
			return nil, err
		}
		var supplier *models.Supplier
		if len(rows) == 0 {
			supplier = &models.Supplier{Code: item.Code, Name: item.Name, Phone: item.Phone, Email: item.Email, Address: item.Address, IsActive: item.IsActive}
			if err := tx.Create(supplier); err != nil {
				return nil, err
			}
		} else {
			supplier = &rows[0]
			if supplier.DeletedAt.Valid {
				if _, err := tx.Restore(supplier); err != nil {
					return nil, err
				}
			}
			if _, err := tx.Model(supplier).Update(map[string]any{"name": item.Name, "phone": item.Phone, "email": item.Email, "address": item.Address, "is_active": item.IsActive}); err != nil {
				return nil, err
			}
		}
		result[item.Code] = supplier
	}
	return result, nil
}

func ensureDevelopmentProducts(tx contractsorm.Query, categories map[string]*models.Category) (map[string]*models.Product, error) {
	result := make(map[string]*models.Product, len(developmentProducts))
	for _, item := range developmentProducts {
		category := categories[item.CategoryName]
		rows := make([]models.Product, 0, 1)
		if err := tx.WithTrashed().Where("code = ?", item.Code).Find(&rows); err != nil {
			return nil, err
		}
		var product *models.Product
		if len(rows) == 0 {
			product = &models.Product{CategoryID: uint64(category.ID), Code: item.Code, Name: item.Name, Description: item.Description, Unit: item.Unit, MinimumStock: item.MinimumStock, IsActive: item.IsActive}
			if err := tx.Create(product); err != nil {
				return nil, err
			}
		} else {
			product = &rows[0]
			if product.DeletedAt.Valid {
				if _, err := tx.Restore(product); err != nil {
					return nil, err
				}
			}
			if _, err := tx.Model(product).Update(map[string]any{"category_id": uint64(category.ID), "name": item.Name, "description": item.Description, "unit": item.Unit, "minimum_stock": item.MinimumStock, "is_active": item.IsActive}); err != nil {
				return nil, err
			}
		}
		result[item.Code] = product
	}
	return result, nil
}

func ensureDevelopmentWarehouses(tx contractsorm.Query) (map[string]*models.Warehouse, error) {
	result := make(map[string]*models.Warehouse, len(developmentWarehouses))
	for _, item := range developmentWarehouses {
		rows := make([]models.Warehouse, 0, 1)
		if err := tx.WithTrashed().Where("code = ?", item.Code).Find(&rows); err != nil {
			return nil, err
		}
		var warehouse *models.Warehouse
		if len(rows) == 0 {
			warehouse = &models.Warehouse{Code: item.Code, Name: item.Name, Location: item.Location, Description: item.Description, IsActive: item.IsActive}
			if err := tx.Create(warehouse); err != nil {
				return nil, err
			}
		} else {
			warehouse = &rows[0]
			if warehouse.DeletedAt.Valid {
				if _, err := tx.Restore(warehouse); err != nil {
					return nil, err
				}
			}
			if _, err := tx.Model(warehouse).Update(map[string]any{"name": item.Name, "location": item.Location, "description": item.Description, "is_active": item.IsActive}); err != nil {
				return nil, err
			}
		}
		result[item.Code] = warehouse
	}
	return result, nil
}

func ensureDevelopmentStocks(products map[string]*models.Product, warehouses map[string]*models.Warehouse) error {
	return facades.Orm().Transaction(func(tx contractsorm.Query) error {
		for _, warehouse := range warehouses {
			for _, product := range products {
				stocks := make([]models.WarehouseStock, 0, 1)
				if err := tx.Where("warehouse_id = ? AND product_id = ?", warehouse.ID, product.ID).Find(&stocks); err != nil {
					return err
				}
				if len(stocks) == 0 {
					if err := tx.Create(&models.WarehouseStock{WarehouseID: uint64(warehouse.ID), ProductID: uint64(product.ID), Quantity: 0}); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func ensureDevelopmentStockIns(ctx context.Context, service *services.StockInService, actorID uint64, suppliers map[string]*models.Supplier, warehouses map[string]*models.Warehouse, products map[string]*models.Product) error {
	for _, item := range developmentStockIns {
		existing, err := findSeedStockIn(item.Note, actorID)
		if err != nil {
			return err
		}
		if existing == nil {
			input := services.CreateStockInDraftInput{SupplierID: uint64(suppliers[item.SupplierCode].ID), WarehouseID: uint64(warehouses[item.WarehouseCode].ID), TransactionDate: item.Date, Note: item.Note, Items: resolveStockInItems(item.Items, products)}
			created, err := service.CreateDraft(ctx, input, actorID)
			if err != nil {
				return err
			}
			if item.Posted {
				if err := service.Post(ctx, created.ID, actorID); err != nil {
					return err
				}
			}
			continue
		}
		if item.Posted && existing.Status == services.StockStatusDraft {
			if err := service.Post(ctx, uint64(existing.ID), actorID); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureDevelopmentStockOuts(ctx context.Context, service *services.StockOutService, actorID uint64, warehouses map[string]*models.Warehouse, products map[string]*models.Product) error {
	for _, item := range developmentStockOuts {
		existing, err := findSeedStockOut(item.Note, actorID)
		if err != nil {
			return err
		}
		if existing == nil {
			input := services.CreateStockOutDraftInput{WarehouseID: uint64(warehouses[item.WarehouseCode].ID), TransactionDate: item.Date, Destination: item.Destination, Note: item.Note, Items: resolveStockOutItems(item.Items, products)}
			created, err := service.CreateDraft(ctx, input, actorID)
			if err != nil {
				return err
			}
			if item.Posted {
				if err := service.Post(ctx, created.ID, actorID); err != nil {
					return err
				}
			}
			continue
		}
		if item.Posted && existing.Status == services.StockStatusDraft {
			if err := service.Post(ctx, uint64(existing.ID), actorID); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureDevelopmentTransfers(ctx context.Context, service *services.WarehouseTransferService, actorID uint64, warehouses map[string]*models.Warehouse, products map[string]*models.Product) error {
	for _, item := range developmentTransfers {
		existing, err := findSeedTransfer(item.Note, actorID)
		if err != nil {
			return err
		}
		if existing == nil {
			input := services.CreateWarehouseTransferDraftInput{FromWarehouseID: uint64(warehouses[item.FromWarehouseCode].ID), ToWarehouseID: uint64(warehouses[item.ToWarehouseCode].ID), TransactionDate: item.Date, Note: item.Note, Items: resolveTransferItems(item.Items, products)}
			created, err := service.CreateDraft(ctx, input, actorID)
			if err != nil {
				return err
			}
			if item.Posted {
				if err := service.Post(ctx, created.ID, actorID); err != nil {
					return err
				}
			}
			continue
		}
		if item.Posted && existing.Status == services.StockStatusDraft {
			if err := service.Post(ctx, uint64(existing.ID), actorID); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveStockInItems(items []services.CreateStockInItemInput, products map[string]*models.Product) []services.CreateStockInItemInput {
	result := make([]services.CreateStockInItemInput, len(items))
	copy(result, items)
	for i := range result {
		result[i].ProductID = productIDByNumericSeed(result[i].ProductID, products)
	}
	return result
}

func resolveStockOutItems(items []services.CreateStockOutItemInput, products map[string]*models.Product) []services.CreateStockOutItemInput {
	result := make([]services.CreateStockOutItemInput, len(items))
	copy(result, items)
	for i := range result {
		result[i].ProductID = productIDByNumericSeed(result[i].ProductID, products)
	}
	return result
}

func resolveTransferItems(items []services.CreateWarehouseTransferItemInput, products map[string]*models.Product) []services.CreateWarehouseTransferItemInput {
	result := make([]services.CreateWarehouseTransferItemInput, len(items))
	copy(result, items)
	for i := range result {
		result[i].ProductID = productIDByNumericSeed(result[i].ProductID, products)
	}
	return result
}

func productIDByNumericSeed(seed uint64, products map[string]*models.Product) uint64 {
	code := "PRD" + formatThreeDigits(seed)
	return uint64(products[code].ID)
}

func formatThreeDigits(value uint64) string {
	return fmt.Sprintf("%03d", value)
}

func findSeedStockIn(note string, actorID uint64) (*models.StockIn, error) {
	rows := make([]models.StockIn, 0, 1)
	if err := facades.Orm().Query().Where("note = ? AND created_by = ?", note, actorID).Find(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func findSeedStockOut(note string, actorID uint64) (*models.StockOut, error) {
	rows := make([]models.StockOut, 0, 1)
	if err := facades.Orm().Query().Where("note = ? AND created_by = ?", note, actorID).Find(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func findSeedTransfer(note string, actorID uint64) (*models.WarehouseTransfer, error) {
	rows := make([]models.WarehouseTransfer, 0, 1)
	if err := facades.Orm().Query().Where("note = ? AND created_by = ?", note, actorID).Find(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
