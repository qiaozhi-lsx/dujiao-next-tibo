package orderhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/modules/catalog/product/manualform"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	procurementcontract "github.com/dujiao-next/internal/modules/procurement/contract"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/gin-gonic/gin"
)

// FulfillService 支付后补填开通信息与进度查询端口（由采购模块契约提供）。
type FulfillService interface {
	SubmitFulfillInfo(ctx context.Context, localOrderID uint, formData jsonmap.JSON) error
	GetFulfillProgress(ctx context.Context, localOrderID uint) (*procurementcontract.FulfillProgress, error)
}

// FulfillInfoRequest 提交开通信息请求。
type FulfillInfoRequest struct {
	ManualFormData jsonmap.JSON `json:"manual_form_data" binding:"required"`
}

// FulfillHandler 处理「支付后填写开通信息 / 查询开通进度」HTTP（登录用户与游客）。
type FulfillHandler struct {
	users  UserOrderQuery
	guests GuestOrderQuery
	fulfil FulfillService
}

func NewFulfillHandler(users UserOrderQuery, guests GuestOrderQuery, fulfil FulfillService) *FulfillHandler {
	if users == nil || guests == nil || fulfil == nil {
		panic("order fulfill handler: required dependency is nil")
	}
	return &FulfillHandler{users: users, guests: guests, fulfil: fulfil}
}

// SubmitUserFulfillInfo 登录用户提交开通信息。
func (h *FulfillHandler) SubmitUserFulfillInfo(c *gin.Context) {
	order, ok := h.loadUserOrder(c)
	if !ok {
		return
	}
	h.submit(c, order)
}

// SubmitGuestFulfillInfo 游客提交开通信息。
func (h *FulfillHandler) SubmitGuestFulfillInfo(c *gin.Context) {
	order, ok := h.loadGuestOrder(c)
	if !ok {
		return
	}
	h.submit(c, order)
}

// GetUserFulfillProgress 登录用户查询开通进度。
func (h *FulfillHandler) GetUserFulfillProgress(c *gin.Context) {
	order, ok := h.loadUserOrder(c)
	if !ok {
		return
	}
	h.progress(c, order)
}

// GetGuestFulfillProgress 游客查询开通进度。
func (h *FulfillHandler) GetGuestFulfillProgress(c *gin.Context) {
	order, ok := h.loadGuestOrder(c)
	if !ok {
		return
	}
	h.progress(c, order)
}

func (h *FulfillHandler) loadUserOrder(c *gin.Context) (*orderdomain.Order, bool) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return nil, false
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return nil, false
	}
	order, err := h.users.GetOrderByUserOrderNoForTenant(tenantFromRequest(c), orderNo, uid)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			ginutil.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
			return nil, false
		}
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return nil, false
	}
	return order, order != nil
}

func (h *FulfillHandler) loadGuestOrder(c *gin.Context) (*orderdomain.Order, bool) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return nil, false
	}
	if password == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_password_required", nil)
		return nil, false
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return nil, false
	}
	order, err := h.guests.GetOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
	if err != nil {
		if errors.Is(err, ErrGuestOrderNotFound) {
			ginutil.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
			return nil, false
		}
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return nil, false
	}
	return order, order != nil
}

func (h *FulfillHandler) submit(c *gin.Context, order *orderdomain.Order) {
	var req FulfillInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}
	err := h.fulfil.SubmitFulfillInfo(c.Request.Context(), order.ID, req.ManualFormData)
	if err == nil {
		c.Header("Cache-Control", "no-store")
		response.Success(c, gin.H{"submitted": true})
		return
	}

	var rejected *procurementcontract.SessionCheckRejectedError
	switch {
	case errors.As(err, &rejected):
		// 预检不通过：真实 HTTP 422，data 带 reason_code / message / account 供前端内联展示。
		msg := i18n.T(i18n.ResolveLocale(c), "error.fulfill_session_rejected")
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusUnprocessableEntity, response.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Msg:        msg,
			Data: gin.H{
				"reason_code": rejected.ReasonCode,
				"message":     rejected.Message,
				"account":     rejected.Account,
			},
		})
	case errors.Is(err, procurementcontract.ErrFulfillInfoAlreadySubmitted):
		respondFulfillConflict(c, "error.fulfill_info_already_submitted")
	case errors.Is(err, procurementcontract.ErrFulfillInfoNotAllowed):
		respondFulfillConflict(c, "error.fulfill_info_not_allowed")
	case errors.Is(err, manualform.ErrRequiredMissing):
		ginutil.RespondError(c, response.CodeBadRequest, "error.manual_form_required_missing", nil)
	case errors.Is(err, manualform.ErrFieldInvalid):
		ginutil.RespondError(c, response.CodeBadRequest, "error.manual_form_field_invalid", nil)
	case errors.Is(err, manualform.ErrTypeInvalid):
		ginutil.RespondError(c, response.CodeBadRequest, "error.manual_form_type_invalid", nil)
	case errors.Is(err, manualform.ErrOptionInvalid):
		ginutil.RespondError(c, response.CodeBadRequest, "error.manual_form_option_invalid", nil)
	case errors.Is(err, manualform.ErrSchemaInvalid):
		ginutil.RespondError(c, response.CodeBadRequest, "error.manual_form_schema_invalid", nil)
	default:
		ginutil.RespondError(c, response.CodeInternal, "error.order_update_failed", err)
	}
}

// respondFulfillConflict 状态冲突：真实 HTTP 409。
func respondFulfillConflict(c *gin.Context, key string) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusConflict, response.Response{
		StatusCode: http.StatusConflict,
		Msg:        i18n.T(i18n.ResolveLocale(c), key),
		Data:       nil,
	})
}

func (h *FulfillHandler) progress(c *gin.Context, order *orderdomain.Order) {
	result, err := h.fulfil.GetFulfillProgress(c.Request.Context(), order.ID)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	if result.MessageKey != "" && result.Message == "" {
		result.Message = i18n.T(i18n.ResolveLocale(c), result.MessageKey)
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, result)
}
