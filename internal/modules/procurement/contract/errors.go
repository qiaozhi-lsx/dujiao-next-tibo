package contract

import "errors"

var (
	ErrNotFound           = errors.New("procurement order not found")
	ErrExists             = errors.New("procurement order already exists")
	ErrStatusInvalid      = errors.New("procurement order status invalid")
	ErrOrderNotFound      = errors.New("order not found")
	ErrConnectionNotFound = errors.New("site connection not found")
	// ErrCapabilityUnsupported 上游未提供所请求的可选能力。
	ErrCapabilityUnsupported = errors.New("upstream capability unsupported")
	// ErrFulfillInfoNotAllowed 订单当前不接受开通信息（未支付、无待填表单等）。
	ErrFulfillInfoNotAllowed = errors.New("fulfill info not allowed")
	// ErrFulfillInfoAlreadySubmitted 开通信息已提交，订单正在开通。
	ErrFulfillInfoAlreadySubmitted = errors.New("fulfill info already submitted")
)
