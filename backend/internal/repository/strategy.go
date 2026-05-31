package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"warden/internal/model"
	"warden/pkg/errcode"
)

// StrategyRepository 策略数据访问接口。所有方法强制带 user_id，杜绝越权。
type StrategyRepository interface {
	List(ctx context.Context, userID uint, kw, tag string) ([]model.Strategy, error)
	Create(ctx context.Context, s *model.Strategy) error
	GetByID(ctx context.Context, userID, id uint) (*model.Strategy, error)
	Update(ctx context.Context, s *model.Strategy) error
	Delete(ctx context.Context, userID, id uint) error

	GetIndicator(ctx context.Context, strategyID uint) (*model.StrategyIndicator, error)
	SaveIndicator(ctx context.Context, ind *model.StrategyIndicator) error

	GetSkill(ctx context.Context, strategyID uint) (*model.StrategySkill, error)
	SaveSkill(ctx context.Context, skill *model.StrategySkill) error
}

// ScreenResultRepository 粗筛任务/结果数据访问接口。
type ScreenResultRepository interface {
	Create(ctx context.Context, r *model.StrategyScreenResult) error
	Update(ctx context.Context, r *model.StrategyScreenResult) error
	GetByID(ctx context.Context, userID, strategyID, id uint) (*model.StrategyScreenResult, error)
	Latest(ctx context.Context, userID, strategyID uint) (*model.StrategyScreenResult, error)
}

type strategyRepo struct{ db *gorm.DB }

// NewStrategyRepository 构造基于 GORM 的策略仓储。
func NewStrategyRepository(db *gorm.DB) StrategyRepository { return &strategyRepo{db: db} }

func (r *strategyRepo) List(ctx context.Context, userID uint, kw, tag string) ([]model.Strategy, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if kw != "" {
		q = q.Where("name LIKE ?", "%"+kw+"%")
	}
	if tag != "" {
		q = q.Where("tags LIKE ?", "%"+tag+"%")
	}
	var list []model.Strategy
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *strategyRepo) Create(ctx context.Context, s *model.Strategy) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *strategyRepo) GetByID(ctx context.Context, userID, id uint) (*model.Strategy, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var s model.Strategy
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrStrategyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *strategyRepo) Update(ctx context.Context, s *model.Strategy) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).
		Model(&model.Strategy{}).
		Where("user_id = ? AND id = ?", s.UserID, s.ID).
		Updates(map[string]any{
			"name":        s.Name,
			"description": s.Description,
			"tags":        s.Tags,
			"status":      s.Status,
		}).Error
}

func (r *strategyRepo) Delete(ctx context.Context, userID, id uint) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.Strategy{}).Error
}

func (r *strategyRepo) GetIndicator(ctx context.Context, strategyID uint) (*model.StrategyIndicator, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var ind model.StrategyIndicator
	err := r.db.WithContext(ctx).Where("strategy_id = ?", strategyID).First(&ind).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ind, nil
}

func (r *strategyRepo) SaveIndicator(ctx context.Context, ind *model.StrategyIndicator) error {
	if r.db == nil {
		return errNoDB
	}
	existing, err := r.GetIndicator(ctx, ind.StrategyID)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.db.WithContext(ctx).Create(ind).Error
	}
	return r.db.WithContext(ctx).
		Model(&model.StrategyIndicator{}).
		Where("id = ?", existing.ID).
		Update("conditions", ind.Conditions).Error
}

func (r *strategyRepo) GetSkill(ctx context.Context, strategyID uint) (*model.StrategySkill, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var skill model.StrategySkill
	err := r.db.WithContext(ctx).
		Where("strategy_id = ?", strategyID).
		Order("version DESC").
		First(&skill).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *strategyRepo) SaveSkill(ctx context.Context, skill *model.StrategySkill) error {
	if r.db == nil {
		return errNoDB
	}
	latest, err := r.GetSkill(ctx, skill.StrategyID)
	if err != nil {
		return err
	}
	skill.Version = 1
	if latest != nil {
		skill.Version = latest.Version + 1
	}
	return r.db.WithContext(ctx).Create(skill).Error
}

type screenResultRepo struct{ db *gorm.DB }

// NewScreenResultRepository 构造基于 GORM 的粗筛结果仓储。
func NewScreenResultRepository(db *gorm.DB) ScreenResultRepository { return &screenResultRepo{db: db} }

func (r *screenResultRepo) Create(ctx context.Context, sr *model.StrategyScreenResult) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).Create(sr).Error
}

func (r *screenResultRepo) Update(ctx context.Context, sr *model.StrategyScreenResult) error {
	if r.db == nil {
		return errNoDB
	}
	return r.db.WithContext(ctx).
		Model(&model.StrategyScreenResult{}).
		Where("id = ?", sr.ID).
		Updates(map[string]any{
			"status":         sr.Status,
			"universe_count": sr.UniverseCount,
			"matched_count":  sr.MatchedCount,
			"candidates":     sr.Candidates,
			"trade_date":     sr.TradeDate,
			"error_msg":      sr.ErrorMsg,
		}).Error
}

func (r *screenResultRepo) GetByID(ctx context.Context, userID, strategyID, id uint) (*model.StrategyScreenResult, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var sr model.StrategyScreenResult
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND strategy_id = ? AND id = ?", userID, strategyID, id).
		First(&sr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrScreenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

func (r *screenResultRepo) Latest(ctx context.Context, userID, strategyID uint) (*model.StrategyScreenResult, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var sr model.StrategyScreenResult
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND strategy_id = ?", userID, strategyID).
		Order("id DESC").
		First(&sr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrScreenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sr, nil
}
