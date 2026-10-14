package application

import (
	"context"
	"testing"
)

// 回归：create-and-pay 响应的 channel_name 依赖 Payment.ChannelName。
// 该字段带 gorm:"-"（非数据库字段，JOIN 别名也不会回填），必须在建单时显式赋值，
// 否则渠道 chip 在弹窗里永远为空。
func TestCreateOrderPaymentPopulatesChannelName(t *testing.T) {
	svc, db := setupPaymentServiceWalletTest(t)
	channel, order := createFeePolicyOrderFixture(t, db, "支付宝", "DJ-CHANNEL-NAME")
	registerTestGateway(t, svc, channel.ProviderType, channel.ChannelType, emptyProviderRefProvider{})

	result, err := svc.CreatePayment(CreatePaymentInput{OrderID: order.ID, ChannelID: channel.ID, Context: context.Background()})
	if err != nil {
		t.Fatalf("create payment failed: %v", err)
	}
	if result.Payment == nil {
		t.Fatal("payment must not be nil")
	}
	if result.Payment.ChannelName != channel.Name {
		t.Fatalf("payment channel_name want %q got %q", channel.Name, result.Payment.ChannelName)
	}
}
