package dictionary

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

// Repository handles all DB operations for DictType and DictItem.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// ─── DictType ────────────────────────────────────────────────────────────────

func (r *Repository) ListTypes(ctx context.Context, offset, limit int, keyword, status string) ([]models.DictType, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.DictType{}).Where("tenant_id = ?", "default")
	if keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		q = q.Where("(LOWER(code) LIKE ? OR LOWER(name) LIKE ?)", like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.DictType
	if err := q.Order("sort_order ASC, created_at ASC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repository) GetTypeByID(ctx context.Context, id uuid.UUID) (*models.DictType, error) {
	var dt models.DictType
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, "default").First(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *Repository) GetTypeByCode(ctx context.Context, code string) (*models.DictType, error) {
	var dt models.DictType
	if err := r.db.WithContext(ctx).Where("code = ? AND tenant_id = ?", code, "default").First(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *Repository) CreateType(ctx context.Context, dt *models.DictType) error {
	dt.TenantID = "default"
	return r.db.WithContext(ctx).Create(dt).Error
}

func (r *Repository) UpdateType(ctx context.Context, dt *models.DictType) error {
	return r.db.WithContext(ctx).Save(dt).Error
}

func (r *Repository) DeleteType(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, "default").Delete(&models.DictType{}).Error
}

func (r *Repository) CountItems(ctx context.Context, typeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.DictItem{}).
		Where("dict_type_id = ? AND tenant_id = ?", typeID, "default").
		Count(&count).Error
	return count, err
}

// CountItemsBatch returns item counts for multiple type IDs.
func (r *Repository) CountItemsBatch(ctx context.Context, typeIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(typeIDs) == 0 {
		return map[uuid.UUID]int64{}, nil
	}
	type row struct {
		DictTypeID uuid.UUID
		Cnt        int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&models.DictItem{}).
		Select("dict_type_id, COUNT(*) as cnt").
		Where("dict_type_id IN ? AND tenant_id = ?", typeIDs, "default").
		Group("dict_type_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]int64, len(rows))
	for _, r := range rows {
		m[r.DictTypeID] = r.Cnt
	}
	return m, nil
}

// ─── DictItem ────────────────────────────────────────────────────────────────

func (r *Repository) ListItems(ctx context.Context, typeID uuid.UUID) ([]models.DictItem, error) {
	var list []models.DictItem
	err := r.db.WithContext(ctx).
		Where("dict_type_id = ? AND tenant_id = ?", typeID, "default").
		Order("sort_order ASC, created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *Repository) GetItemByID(ctx context.Context, typeID, itemID uuid.UUID) (*models.DictItem, error) {
	var item models.DictItem
	if err := r.db.WithContext(ctx).
		Where("id = ? AND dict_type_id = ? AND tenant_id = ?", itemID, typeID, "default").
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) CreateItem(ctx context.Context, item *models.DictItem) error {
	item.TenantID = "default"
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) UpdateItem(ctx context.Context, item *models.DictItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) DeleteItem(ctx context.Context, typeID, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND dict_type_id = ? AND tenant_id = ?", itemID, typeID, "default").
		Delete(&models.DictItem{}).Error
}

// BatchReplaceItems replaces all items under a dict type in a transaction.
func (r *Repository) BatchReplaceItems(ctx context.Context, typeID uuid.UUID, items []models.DictItem) ([]models.DictItem, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Hard-delete existing items (bypass soft delete for clean replace)
		if err := tx.Unscoped().Where("dict_type_id = ? AND tenant_id = ?", typeID, "default").Delete(&models.DictItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].TenantID = "default"
			items[i].DictTypeID = typeID
			if items[i].ID == uuid.Nil {
				items[i].ID = uuid.New()
			}
		}
		return tx.Create(&items).Error
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetItemsByTypeCodes returns items grouped by dict type code.
// Only returns items belonging to normal-status types.
func (r *Repository) GetItemsByTypeCodes(ctx context.Context, codes []string) (map[string][]models.DictItem, error) {
	if len(codes) == 0 {
		return map[string][]models.DictItem{}, nil
	}
	// Find matching types
	var types []models.DictType
	err := r.db.WithContext(ctx).
		Where("code IN ? AND tenant_id = ? AND status = ?", codes, "default", "normal").
		Find(&types).Error
	if err != nil {
		return nil, err
	}
	if len(types) == 0 {
		return map[string][]models.DictItem{}, nil
	}

	typeIDToCode := make(map[uuid.UUID]string, len(types))
	typeIDs := make([]uuid.UUID, 0, len(types))
	for _, t := range types {
		typeIDToCode[t.ID] = t.Code
		typeIDs = append(typeIDs, t.ID)
	}

	var items []models.DictItem
	err = r.db.WithContext(ctx).
		Where("dict_type_id IN ? AND tenant_id = ? AND status = ?", typeIDs, "default", "normal").
		Order("sort_order ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]models.DictItem, len(types))
	for _, item := range items {
		code := typeIDToCode[item.DictTypeID]
		result[code] = append(result[code], item)
	}
	// Ensure all requested codes that have a type are present (even if empty)
	for _, t := range types {
		if _, ok := result[t.Code]; !ok {
			result[t.Code] = []models.DictItem{}
		}
	}
	return result, nil
}

// DeleteItemsByTypeID hard-deletes all items under a dict type.
func (r *Repository) DeleteItemsByTypeID(ctx context.Context, typeID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("dict_type_id = ? AND tenant_id = ?", typeID, "default").
		Delete(&models.DictItem{}).Error
}
