package service

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"warden/internal/dto/request"
	"warden/internal/dto/response"
	"warden/internal/integration/market"
	"warden/internal/model"
	"warden/internal/repository"
	"warden/internal/strategy/catalog"
	"warden/internal/strategy/factor"
	"warden/internal/strategy/rule"
	"warden/pkg/errcode"
)

const (
	defaultScreenLimit = 200
	// klineLookback 拉取的 K 线根数，需足够计算长周期均线 + 连续振幅。
	klineLookback = 120
	// defaultScreenConcurrency 粗筛内并发拉取 K 线的默认协程数，与 gotdx 连接池默认值对齐。
	defaultScreenConcurrency = 8
)

// StrategyService M2 选股策略 + 量化粗筛业务接口。
type StrategyService interface {
	List(ctx context.Context, userID uint, kw, tag string) ([]response.Strategy, error)
	Create(ctx context.Context, userID uint, req *request.CreateStrategyReq) (*response.Strategy, error)
	Get(ctx context.Context, userID, id uint) (*response.Strategy, error)
	Update(ctx context.Context, userID, id uint, req *request.UpdateStrategyReq) error
	Delete(ctx context.Context, userID, id uint) error
	Copy(ctx context.Context, userID, id uint) (*response.Strategy, error)

	UpdateIndicators(ctx context.Context, userID, id uint, group *rule.RuleGroup) error
	GetSkill(ctx context.Context, userID, id uint) (string, int, error)
	SaveSkill(ctx context.Context, userID, id uint, content string) error

	Catalog() []catalog.Item
	Templates() []catalog.Template

	RunScreen(ctx context.Context, userID, strategyID uint, req *request.ScreenReq) (uint, error)
	GetScreenResult(ctx context.Context, userID, strategyID, taskID uint) (*response.ScreenResult, error)
	LatestScreenResult(ctx context.Context, userID, strategyID uint) (*response.ScreenResult, error)
	PreviewScreen(ctx context.Context, userID uint, req *request.PreviewScreenReq) (*response.ScreenResult, error)
}

type strategyService struct {
	repo              repository.StrategyRepository
	screenRepo        repository.ScreenResultRepository
	watchRepo         repository.WatchlistRepository
	provider          market.IMarketProvider
	screenConcurrency int
}

// StrategyOption 策略服务可选注入项，保持 NewStrategyService 向后兼容。
type StrategyOption func(*strategyService)

// WithScreenConcurrency 配置粗筛并发度，<=0 时使用 defaultScreenConcurrency。
func WithScreenConcurrency(n int) StrategyOption {
	return func(s *strategyService) { s.screenConcurrency = n }
}

// NewStrategyService 注入依赖构造策略业务。
func NewStrategyService(
	repo repository.StrategyRepository,
	screenRepo repository.ScreenResultRepository,
	watchRepo repository.WatchlistRepository,
	provider market.IMarketProvider,
	opts ...StrategyOption,
) StrategyService {
	s := &strategyService{repo: repo, screenRepo: screenRepo, watchRepo: watchRepo, provider: provider}
	for _, opt := range opts {
		opt(s)
	}
	if s.screenConcurrency <= 0 {
		s.screenConcurrency = defaultScreenConcurrency
	}
	return s
}

func (s *strategyService) List(ctx context.Context, userID uint, kw, tag string) ([]response.Strategy, error) {
	list, err := s.repo.List(ctx, userID, kw, tag)
	if err != nil {
		return nil, err
	}
	out := make([]response.Strategy, 0, len(list))
	for _, m := range list {
		out = append(out, response.Strategy{
			ID: m.ID, Name: m.Name, Description: m.Description,
			Tags: m.Tags, Status: m.Status, CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}

func (s *strategyService) Create(ctx context.Context, userID uint, req *request.CreateStrategyReq) (*response.Strategy, error) {
	if req.Indicators != nil {
		if err := validateGroup(*req.Indicators); err != nil {
			return nil, err
		}
	}
	m := &model.Strategy{UserID: userID, Name: req.Name, Description: req.Description, Tags: req.Tags, Status: model.StrategyStatusEnabled}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	if req.Indicators != nil {
		if err := s.saveIndicator(ctx, m.ID, req.Indicators); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, userID, m.ID)
}

func (s *strategyService) Get(ctx context.Context, userID, id uint) (*response.Strategy, error) {
	m, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	out := &response.Strategy{
		ID: m.ID, Name: m.Name, Description: m.Description,
		Tags: m.Tags, Status: m.Status, CreatedAt: m.CreatedAt,
	}
	if group, gErr := s.loadIndicator(ctx, id); gErr == nil && group != nil {
		out.Indicators = group
	}
	if skill, sErr := s.repo.GetSkill(ctx, id); sErr == nil && skill != nil {
		out.Skill = skill.Content
		out.SkillVersion = skill.Version
	}
	return out, nil
}

func (s *strategyService) Update(ctx context.Context, userID, id uint, req *request.UpdateStrategyReq) error {
	m, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	m.Name, m.Description, m.Tags = req.Name, req.Description, req.Tags
	if err := s.repo.Update(ctx, m); err != nil {
		return err
	}
	if req.Indicators != nil {
		return s.UpdateIndicators(ctx, userID, id, req.Indicators)
	}
	return nil
}

func (s *strategyService) Delete(ctx context.Context, userID, id uint) error {
	if _, err := s.repo.GetByID(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, userID, id)
}

func (s *strategyService) Copy(ctx context.Context, userID, id uint) (*response.Strategy, error) {
	src, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	req := &request.CreateStrategyReq{
		Name: src.Name + " 副本", Description: src.Description, Tags: src.Tags, Indicators: src.Indicators,
	}
	return s.Create(ctx, userID, req)
}

func (s *strategyService) UpdateIndicators(ctx context.Context, userID, id uint, group *rule.RuleGroup) error {
	if _, err := s.repo.GetByID(ctx, userID, id); err != nil {
		return err
	}
	if group == nil {
		return errcode.ErrIndicatorInvalid.WithMessage("指标定义不能为空")
	}
	if err := validateGroup(*group); err != nil {
		return err
	}
	return s.saveIndicator(ctx, id, group)
}

func (s *strategyService) GetSkill(ctx context.Context, userID, id uint) (string, int, error) {
	if _, err := s.repo.GetByID(ctx, userID, id); err != nil {
		return "", 0, err
	}
	skill, err := s.repo.GetSkill(ctx, id)
	if err != nil {
		return "", 0, err
	}
	if skill == nil {
		return "", 0, nil
	}
	return skill.Content, skill.Version, nil
}

func (s *strategyService) SaveSkill(ctx context.Context, userID, id uint, content string) error {
	if _, err := s.repo.GetByID(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.SaveSkill(ctx, &model.StrategySkill{StrategyID: id, Content: content})
}

func (s *strategyService) Catalog() []catalog.Item { return catalog.Items() }

func (s *strategyService) Templates() []catalog.Template { return catalog.Templates() }

// RunScreen 基于策略已存指标定义对股票池粗筛，落库并返回任务 ID。
// V1 采用请求内同步执行（适合自选/自定义代码池），结果状态直接为 done/failed。
func (s *strategyService) RunScreen(ctx context.Context, userID, strategyID uint, req *request.ScreenReq) (uint, error) {
	if _, err := s.repo.GetByID(ctx, userID, strategyID); err != nil {
		return 0, err
	}
	group, err := s.loadIndicator(ctx, strategyID)
	if err != nil {
		return 0, err
	}
	if group == nil {
		return 0, errcode.ErrIndicatorInvalid.WithMessage("策略尚未定义量化指标")
	}

	paramsJSON, _ := json.Marshal(req)
	rec := &model.StrategyScreenResult{
		UserID: userID, StrategyID: strategyID,
		Status: model.ScreenStatusRunning, Params: model.JSON(paramsJSON),
	}
	if err := s.screenRepo.Create(ctx, rec); err != nil {
		return 0, err
	}

	result, err := s.screen(ctx, *group, req, userID)
	if err != nil {
		rec.Status = model.ScreenStatusFailed
		rec.ErrorMsg = err.Error()
		_ = s.screenRepo.Update(ctx, rec)
		return rec.ID, err
	}

	candJSON, _ := json.Marshal(result.Candidates)
	rec.Status = model.ScreenStatusDone
	rec.UniverseCount = result.UniverseCount
	rec.MatchedCount = result.MatchedCount
	rec.Candidates = model.JSON(candJSON)
	if req.TradeDate != "" {
		if td, pErr := time.Parse("2006-01-02", req.TradeDate); pErr == nil {
			rec.TradeDate = &td
		}
	}
	if err := s.screenRepo.Update(ctx, rec); err != nil {
		return rec.ID, err
	}
	return rec.ID, nil
}

func (s *strategyService) GetScreenResult(ctx context.Context, userID, strategyID, taskID uint) (*response.ScreenResult, error) {
	rec, err := s.screenRepo.GetByID(ctx, userID, strategyID, taskID)
	if err != nil {
		return nil, err
	}
	return screenModelToResponse(rec), nil
}

func (s *strategyService) LatestScreenResult(ctx context.Context, userID, strategyID uint) (*response.ScreenResult, error) {
	rec, err := s.screenRepo.Latest(ctx, userID, strategyID)
	if err != nil {
		return nil, err
	}
	return screenModelToResponse(rec), nil
}

func (s *strategyService) PreviewScreen(ctx context.Context, userID uint, req *request.PreviewScreenReq) (*response.ScreenResult, error) {
	if req.Indicators == nil {
		return nil, errcode.ErrIndicatorInvalid.WithMessage("预览需提供指标定义")
	}
	if err := validateGroup(*req.Indicators); err != nil {
		return nil, err
	}
	result, err := s.screen(ctx, *req.Indicators, &req.ScreenReq, userID)
	if err != nil {
		return nil, err
	}
	result.Status = model.ScreenStatusDone
	return result, nil
}

// screen 是粗筛核心：解析股票池 → 并发拉 K 线 → 规则匹配 → 聚合候选 → 评分排序截断。
//
// 并发模型：信号量 channel 限流 + WaitGroup 同步，单只标的拉取/评估失败不中断整轮；
// 仅指标配置类错误（未知因子/op 非法）作为致命错误中止并向上抛出。
// 外层 ctx 取消会即时退出工作循环，下游 provider.Kline 会感知 ctx.Done 主动放弃。
func (s *strategyService) screen(ctx context.Context, group rule.RuleGroup, req *request.ScreenReq, userID uint) (*response.ScreenResult, error) {
	codes, names, err := s.resolveUniverse(ctx, userID, req.Universe)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, errcode.ErrUniverseInvalid.WithMessage("股票池为空")
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultScreenLimit
	}

	concurrency := s.screenConcurrency
	if concurrency <= 0 {
		concurrency = defaultScreenConcurrency
	}
	if concurrency > len(codes) {
		concurrency = len(codes)
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		cands      = make([]response.ScreenCandidate, 0, 16)
		fatalErr   error
		sem        = make(chan struct{}, concurrency)
	)

	for _, code := range codes {
		select {
		case <-ctx.Done():
			// 外层超时/取消：停止派发新任务（已派发的会在自身分支感知 ctx 退出）。
			goto done
		case sem <- struct{}{}:
		}

		mu.Lock()
		stop := fatalErr != nil
		mu.Unlock()
		if stop {
			<-sem
			break
		}

		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := ctx.Err(); err != nil {
				return
			}
			klines, kErr := s.provider.Kline(ctx, code, "day", "qfq")
			if kErr != nil || len(klines) == 0 {
				return // 单只失败跳过，不中断整轮
			}
			res, eErr := rule.Eval(group, factor.Series{Bars: klines})
			if eErr != nil {
				// 指标配置错误（未知因子/op 非法），整轮一致，作为致命错误抛出。
				mu.Lock()
				if fatalErr == nil {
					fatalErr = errcode.ErrIndicatorInvalid.Wrap(eErr)
				}
				mu.Unlock()
				return
			}
			if !res.Matched {
				return
			}
			factors := make(map[string]decimal.Decimal, len(res.Snapshot))
			for k, v := range res.Snapshot {
				factors[k] = v.Decimal().Round(4)
			}
			cand := response.ScreenCandidate{
				StockCode: code,
				StockName: names[code],
				Score:     decimal.NewFromFloat(res.Score()).Round(4),
				Factors:   factors,
				Matched:   res.HitRules,
			}
			mu.Lock()
			cands = append(cands, cand)
			mu.Unlock()
		}(code)
	}

done:
	wg.Wait()

	if fatalErr != nil {
		return nil, fatalErr
	}
	if err := ctx.Err(); err != nil {
		return nil, errcode.ErrTimeout.Wrap(err)
	}

	sort.SliceStable(cands, func(i, j int) bool { return cands[i].Score.GreaterThan(cands[j].Score) })
	matched := len(cands)
	if len(cands) > limit {
		cands = cands[:limit]
	}
	s.fillNames(ctx, cands, names)

	return &response.ScreenResult{
		StrategyID:    0,
		Status:        model.ScreenStatusDone,
		TradeDate:     req.TradeDate,
		UniverseCount: len(codes),
		MatchedCount:  matched,
		Candidates:    cands,
	}, nil
}

// resolveUniverse 解析股票池为代码列表 + 名称映射。
func (s *strategyService) resolveUniverse(ctx context.Context, userID uint, u request.Universe) ([]string, map[string]string, error) {
	names := make(map[string]string)
	switch u.Type {
	case "codes":
		return dedup(u.Codes), names, nil
	case "watchlist", "":
		items, err := s.watchRepo.List(ctx, userID)
		if err != nil {
			return nil, nil, err
		}
		codes := make([]string, 0, len(items))
		for _, it := range items {
			codes = append(codes, it.StockCode)
			names[it.StockCode] = it.StockName
		}
		return codes, names, nil
	case "all", "board":
		return nil, nil, errcode.ErrUniverseInvalid.WithMessage("全市场/板块粗筛需个股指标快照（后续迭代），请使用自选股或自定义代码池")
	default:
		return nil, nil, errcode.ErrUniverseInvalid.WithMessage("未知股票池类型: " + u.Type)
	}
}

// fillNames best-effort 通过行情源补全候选股名称（失败不阻塞）。
func (s *strategyService) fillNames(ctx context.Context, cands []response.ScreenCandidate, names map[string]string) {
	missing := make([]string, 0)
	for _, c := range cands {
		if c.StockName == "" {
			missing = append(missing, c.StockCode)
		}
	}
	if len(missing) == 0 {
		return
	}
	quotes, err := s.provider.Quotes(ctx, missing)
	if err != nil {
		return
	}
	for _, q := range quotes {
		names[q.StockCode] = q.StockName
	}
	for i := range cands {
		if cands[i].StockName == "" {
			cands[i].StockName = names[cands[i].StockCode]
		}
	}
}

func (s *strategyService) saveIndicator(ctx context.Context, strategyID uint, group *rule.RuleGroup) error {
	b, err := json.Marshal(group)
	if err != nil {
		return errcode.ErrIndicatorInvalid.Wrap(err)
	}
	return s.repo.SaveIndicator(ctx, &model.StrategyIndicator{StrategyID: strategyID, Conditions: model.JSON(b)})
}

func (s *strategyService) loadIndicator(ctx context.Context, strategyID uint) (*rule.RuleGroup, error) {
	ind, err := s.repo.GetIndicator(ctx, strategyID)
	if err != nil {
		return nil, err
	}
	if ind == nil || len(ind.Conditions) == 0 {
		return nil, nil
	}
	var group rule.RuleGroup
	if err := json.Unmarshal(ind.Conditions, &group); err != nil {
		return nil, errcode.ErrIndicatorInvalid.Wrap(err)
	}
	return &group, nil
}

// validateGroup 校验规则组结构（递归），用空 K 线触发因子参数校验。
func validateGroup(g rule.RuleGroup) error {
	if len(g.Rules) == 0 && len(g.Groups) == 0 {
		return errcode.ErrIndicatorInvalid.WithMessage("规则组不能为空")
	}
	// 用空序列求值：参数非法/未知因子会立即报错，数据不足则忽略（合法）。
	if _, err := rule.Eval(g, factor.Series{}); err != nil {
		if isInsufficient(err) {
			return nil
		}
		return errcode.ErrIndicatorInvalid.Wrap(err)
	}
	return nil
}

func isInsufficient(err error) bool {
	return err != nil && (err == factor.ErrInsufficientData)
}

func screenModelToResponse(m *model.StrategyScreenResult) *response.ScreenResult {
	out := &response.ScreenResult{
		ID: m.ID, StrategyID: m.StrategyID, Status: m.Status,
		UniverseCount: m.UniverseCount, MatchedCount: m.MatchedCount,
		ErrorMsg: m.ErrorMsg, CreatedAt: m.CreatedAt,
		Candidates: []response.ScreenCandidate{},
	}
	if m.TradeDate != nil {
		out.TradeDate = m.TradeDate.Format("2006-01-02")
	}
	if len(m.Candidates) > 0 {
		_ = json.Unmarshal(m.Candidates, &out.Candidates)
	}
	return out
}

func dedup(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
