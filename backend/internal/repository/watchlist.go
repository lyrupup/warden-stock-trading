// Package repository 是数据访问层，仅做 DB CRUD（全程 WithContext），不含业务逻辑。
package repository

import (
	"context"

	"gorm.io/gorm"

	"warden/internal/model"
	"warden/pkg/errcode"
)

// errNoDB 在未连接数据库时返回，避免对 nil *gorm.DB 解引用导致 panic
// （本地无依赖降级运行场景）。
var errNoDB = errcode.ErrInternal.WithMessage("数据库未连接")

// WatchlistRepository 自选股数据访问接口。所有方法强制带 user_id，杜绝越权。
type WatchlistRepository interface {
	List(ctx context.Context, userID uint) ([]model.WatchlistItem, error)
	Create(ctx context.Context, item *model.WatchlistItem) error
	Delete(ctx context.Context, userID, id uint) error
	ExistsByCode(ctx context.Context, userID uint, code string) (bool, error)
}

type watchlistRepo struct {
	db *gorm.DB
}

// NewWatchlistRepository 构造基于 GORM 的自选股仓储。
func NewWatchlistRepository(db *gorm.DB) WatchlistRepository {
	return &watchlistRepo{db: db}
}

func (r *watchlistRepo) List(ctx context.Context, userID uint) ([]model.WatchlistItem, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var items []model.WatchlistItem
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("sort ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *watchlistRepo) Create(ctx context.Context, item *model.WatchlistItem) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *watchlistRepo) Delete(ctx context.Context, userID, id uint) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.WatchlistItem{}).Error
}

func (r *watchlistRepo) ExistsByCode(ctx context.Context, userID uint, code string) (bool, error) {
	if r.db == nil {
		return false, errNoDB
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.WatchlistItem{}).
		Where("user_id = ? AND stock_code = ?", userID, code).
		Count(&count).Error
	return count > 0, err
}
