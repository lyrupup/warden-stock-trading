package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"warden/internal/dto/request"
	"warden/internal/middleware"
	"warden/internal/service"
	"warden/internal/strategy/rule"
	"warden/pkg/errcode"
	resp "warden/pkg/response"
)

// StrategyHandler 处理 M2 选股策略 + 量化粗筛端点（对齐 openapi.yaml /strategies/*）。
type StrategyHandler struct {
	svc service.StrategyService
}

// NewStrategyHandler 构造策略 Handler。
func NewStrategyHandler(svc service.StrategyService) *StrategyHandler {
	return &StrategyHandler{svc: svc}
}

// List GET /strategies 策略列表。
func (h *StrategyHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.List(c.Request.Context(), userID, c.Query("kw"), c.Query("tag"))
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, list)
}

// Create POST /strategies 新建策略。
func (h *StrategyHandler) Create(c *gin.Context) {
	var req request.CreateStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	userID := middleware.GetUserID(c)
	s, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, s)
}

// Get GET /strategies/:id 策略详情（含指标、skill）。
func (h *StrategyHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	s, err := h.svc.Get(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, s)
}

// Update PUT /strategies/:id 更新策略。
func (h *StrategyHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	var req request.UpdateStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	if err := h.svc.Update(c.Request.Context(), middleware.GetUserID(c), id, &req); err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, nil)
}

// Delete DELETE /strategies/:id 删除策略。
func (h *StrategyHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), middleware.GetUserID(c), id); err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, nil)
}

// Copy POST /strategies/:id/copy 复制策略。
func (h *StrategyHandler) Copy(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	s, err := h.svc.Copy(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, s)
}

// UpdateIndicators PUT /strategies/:id/indicators 更新指标定义。
func (h *StrategyHandler) UpdateIndicators(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	var group rule.RuleGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	if err := h.svc.UpdateIndicators(c.Request.Context(), middleware.GetUserID(c), id, &group); err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, nil)
}

// GetSkill GET /strategies/:id/skill 获取 skill.md。
func (h *StrategyHandler) GetSkill(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	content, version, err := h.svc.GetSkill(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, gin.H{"content": content, "version": version})
}

// SaveSkill PUT /strategies/:id/skill 保存 skill.md（生成新版本）。
func (h *StrategyHandler) SaveSkill(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	var req request.SaveSkillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	if err := h.svc.SaveSkill(c.Request.Context(), middleware.GetUserID(c), id, req.Content); err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, nil)
}

// Catalog GET /strategies/indicators/catalog 量化因子目录。
func (h *StrategyHandler) Catalog(c *gin.Context) {
	resp.Success(c, h.svc.Catalog())
}

// Templates GET /strategies/templates 内置策略模板。
func (h *StrategyHandler) Templates(c *gin.Context) {
	resp.Success(c, h.svc.Templates())
}

// RunScreen POST /strategies/:id/screen 发起量化粗筛。
func (h *StrategyHandler) RunScreen(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	var req request.ScreenReq
	// 允许空 body（默认自选股池）。
	_ = c.ShouldBindJSON(&req)
	taskID, err := h.svc.RunScreen(c.Request.Context(), middleware.GetUserID(c), id, &req)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, gin.H{"taskId": strconv.FormatUint(uint64(taskID), 10)})
}

// ScreenResult GET /strategies/:id/screen/:taskId 查询粗筛结果。
func (h *StrategyHandler) ScreenResult(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	res, err := h.svc.GetScreenResult(c.Request.Context(), middleware.GetUserID(c), id, taskID)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, res)
}

// ScreenLatest GET /strategies/:id/screen/latest 最近一次粗筛结果。
func (h *StrategyHandler) ScreenLatest(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	res, err := h.svc.LatestScreenResult(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, res)
}

// PreviewScreen POST /strategies/screen/preview 同步快速粗筛预览。
func (h *StrategyHandler) PreviewScreen(c *gin.Context) {
	var req request.PreviewScreenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	res, err := h.svc.PreviewScreen(c.Request.Context(), middleware.GetUserID(c), &req)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, res)
}
