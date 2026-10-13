package epusdtadapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/epusdt"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

func TestEpusdtAdapter_Type(t *testing.T) {
	a := NewEpusdtAdapter()
	want := constants.PaymentProviderEpusdt + ":"
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestEpusdtAdapter_ValidateConfig_EmptyRejected(t *testing.T) {
	a := NewEpusdtAdapter()
	err := a.ValidateConfig(jsonmap.JSON{}, "")
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, paymentcontract.ErrGatewayConfigInvalid) {
		t.Fatalf("expected wrapped paymentcontract.ErrGatewayConfigInvalid, got %v", err)
	}
}

func TestEpusdtAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewEpusdtAdapter()
	_, err := a.CreatePayment(context.Background(), jsonmap.JSON{}, paymentcontract.GatewayCreateInput{
		OrderNo:  "ORDER_1",
		Currency: "USDT",
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, paymentcontract.ErrGatewayConfigInvalid) {
		t.Fatalf("expected wrapped paymentcontract.ErrGatewayConfigInvalid, got %v", err)
	}
}

func TestEpusdtDisplayChannelType(t *testing.T) {
	tests := []struct {
		name string
		cfg  *epusdt.Config
		want string
	}{
		{name: "nil", cfg: nil, want: ""},
		{
			name: "cashier",
			cfg:  &epusdt.Config{OrderMode: constants.PaymentEpusdtOrderModeCashier},
			want: "",
		},
		{
			name: "tron usdt",
			cfg:  &epusdt.Config{Token: " USDT ", Network: " TRON "},
			want: "usdt.tron",
		},
		{
			name: "tron trx",
			cfg:  &epusdt.Config{Token: " trx ", Network: " tron "},
			want: "trx.tron",
		},
		{
			name: "fallback payment type",
			cfg:  &epusdt.Config{Token: "usdc", Network: "ethereum", PaymentType: "ethereum-usdc"},
			want: "usdc.ethereum",
		},
		{
			name: "unknown without payment type",
			cfg:  &epusdt.Config{Token: "usdc", Network: "ethereum"},
			want: "usdc.ethereum",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := epusdtDisplayChannelType(tc.cfg); got != tc.want {
				t.Fatalf("epusdtDisplayChannelType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEpusdtAdapter_MapEpusdtError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", epusdt.ErrConfigInvalid, paymentcontract.ErrGatewayConfigInvalid},
		{"request", epusdt.ErrRequestFailed, paymentcontract.ErrGatewayRequestFailed},
		{"response", epusdt.ErrResponseInvalid, paymentcontract.ErrGatewayResponseInvalid},
		{"signature", epusdt.ErrSignatureInvalid, paymentcontract.ErrGatewaySignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapEpusdtError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapEpusdtError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}

func newEpusdtMockServer(t *testing.T, response string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments/gmpay/v1/order/create-transaction" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestEpusdtAdapter_CreatePayment_QRModeUsesReceiveAddress(t *testing.T) {
	srv := newEpusdtMockServer(t, `{"status_code":200,"message":"ok","data":{"trade_id":"T-QR-1","order_id":"ORD-1","amount":6.8,"actual_amount":1,"receive_address":"TWmpTestAddr001","token":"usdt","network":"tron"}}`)

	cfg := jsonmap.JSON{
		"gateway_url": srv.URL,
		"pid":         "1000",
		"secret_key":  "sk-test",
		"token":       "usdt",
		"network":     "tron",
		"currency":    "cny",
		"notify_url":  "https://shop.example.com/api/v1/payments/callback",
		"return_url":  "https://shop.example.com/pay",
	}

	a := NewEpusdtAdapter()
	result, err := a.CreatePayment(context.Background(), cfg, paymentcontract.GatewayCreateInput{
		OrderNo: "ORD-1",
		Subject: "VIP Plan",
		Amount:  money.FromDecimal(decimal.RequireFromString("6.80")),
		Extra:   jsonmap.JSON{"interaction_mode": constants.PaymentInteractionQR},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.QRCodeURL != "TWmpTestAddr001" {
		t.Fatalf("QRCodeURL = %q, want receive address", result.QRCodeURL)
	}
	if result.RedirectURL != "" {
		t.Fatalf("RedirectURL = %q, want empty in qr mode", result.RedirectURL)
	}
}

func TestEpusdtAdapter_CreatePayment_QRModeWithoutAddressFallsBackToPaymentURL(t *testing.T) {
	// 收银台占位单没有链上地址时，二维码内容回退为收银台链接并保留跳转地址。
	srv := newEpusdtMockServer(t, `{"status_code":200,"message":"ok","data":{"trade_id":"T-CASH-1","order_id":"ORD-1","amount":6.8,"actual_amount":0,"receive_address":"","token":"","network":""}}`)

	cfg := jsonmap.JSON{
		"gateway_url": srv.URL,
		"pid":         "1000",
		"secret_key":  "sk-test",
		"order_mode":  constants.PaymentEpusdtOrderModeCashier,
		"currency":    "cny",
		"notify_url":  "https://shop.example.com/api/v1/payments/callback",
		"return_url":  "https://shop.example.com/pay",
	}

	a := NewEpusdtAdapter()
	result, err := a.CreatePayment(context.Background(), cfg, paymentcontract.GatewayCreateInput{
		OrderNo: "ORD-1",
		Amount:  money.FromDecimal(decimal.RequireFromString("6.80")),
		Extra:   jsonmap.JSON{"interaction_mode": constants.PaymentInteractionQR},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	want := srv.URL + "/pay/checkout-counter/T-CASH-1"
	if result.QRCodeURL != want {
		t.Fatalf("QRCodeURL = %q, want fallback %q", result.QRCodeURL, want)
	}
	if result.RedirectURL != want {
		t.Fatalf("RedirectURL = %q, want %q", result.RedirectURL, want)
	}
}

func TestEpusdtAdapter_CreatePayment_RedirectKeepsPaymentURL(t *testing.T) {
	srv := newEpusdtMockServer(t, `{"status_code":200,"message":"ok","data":{"trade_id":"T-RD-1","order_id":"ORD-1","amount":6.8,"actual_amount":1,"receive_address":"TWmpTestAddr001","token":"usdt","network":"tron"}}`)

	cfg := jsonmap.JSON{
		"gateway_url": srv.URL,
		"pid":         "1000",
		"secret_key":  "sk-test",
		"token":       "usdt",
		"network":     "tron",
		"currency":    "cny",
		"notify_url":  "https://shop.example.com/api/v1/payments/callback",
		"return_url":  "https://shop.example.com/pay",
	}

	a := NewEpusdtAdapter()
	result, err := a.CreatePayment(context.Background(), cfg, paymentcontract.GatewayCreateInput{
		OrderNo: "ORD-1",
		Amount:  money.FromDecimal(decimal.RequireFromString("6.80")),
		Extra:   jsonmap.JSON{"interaction_mode": constants.PaymentInteractionRedirect},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	want := srv.URL + "/pay/checkout-counter/T-RD-1"
	if result.QRCodeURL != want || result.RedirectURL != want {
		t.Fatalf("QRCodeURL=%q RedirectURL=%q, want both %q", result.QRCodeURL, result.RedirectURL, want)
	}
}
