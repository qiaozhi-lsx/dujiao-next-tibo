package application

import (
	"context"
	"strings"

	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"

	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"

	"github.com/dujiao-next/internal/constants"
)

// SubmitTxHashInput 买家提交链上交易哈希补单输入。
type SubmitTxHashInput struct {
	PaymentID uint
	TxHash    string
	Context   context.Context
}

// SubmitTxHash 把买家提交的链上交易哈希转发给支持人工核验的网关（epusdt /pay/submit-tx-hash）。
// 网关核验哈希对应交易真实结算了订单后，会走正常回调完成订单；
// 这里只负责转发与错误映射，不直接修改支付状态。
func (s *PaymentService) SubmitTxHash(input SubmitTxHashInput) (*paymentdomain.Payment, error) {
	if input.PaymentID == 0 {
		return nil, ErrPaymentInvalid
	}
	txHash := strings.TrimSpace(input.TxHash)
	if txHash == "" || len(txHash) > 128 {
		return nil, ErrPaymentInvalid
	}
	payment, err := s.paymentRepo.GetByID(input.PaymentID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	if payment.Status == constants.PaymentStatusSuccess {
		return payment, nil
	}

	channel, err := s.channelRepo.GetByID(payment.ChannelID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if channel == nil {
		return nil, ErrPaymentChannelNotFound
	}
	if strings.TrimSpace(payment.ProviderRef) == "" {
		return nil, ErrPaymentInvalid
	}

	// 统一通过 Registry 路由；未实现 GatewayTxHashSubmitter 的 provider
	// type assertion 失败，返回 ErrPaymentProviderNotSupported。
	if s.paymentProviderRegistry == nil {
		return nil, ErrPaymentProviderNotSupported
	}
	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return nil, ErrPaymentProviderNotSupported
	}
	submitter, ok := p.(paymentcontract.GatewayTxHashSubmitter)
	if !ok {
		return nil, ErrPaymentProviderNotSupported
	}
	if err := submitter.ValidateConfig(channel.ConfigJSON, channel.InteractionMode); err != nil {
		return nil, mapProviderErrorToService(err)
	}

	ctx, cancel := detachOutboundRequestContext(input.Context)
	defer cancel()

	if err := submitter.SubmitTxHash(ctx, channel.ConfigJSON, payment.ProviderRef, txHash); err != nil {
		return nil, mapProviderErrorToService(err)
	}
	return payment, nil
}
