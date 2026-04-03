package channel

import (
	"context"
	"fmt"
	"time"
)

// MockChannel Mock支付渠道
// 用于开发和测试环境，模拟第三方支付行为。
// 调用 CreatePayment 后返回一个 mock pay URL，
// 前端（或测试脚本）调用 /payment/mock/pay 接口模拟用户完成支付。
type MockChannel struct{}

func NewMockChannel() *MockChannel {
	return &MockChannel{}
}

func (m *MockChannel) Name() string {
	return "mock"
}

func (m *MockChannel) CreatePayment(_ context.Context, req *PayRequest) (*PayResponse, error) {
	return &PayResponse{
		ChannelTradeNo: fmt.Sprintf("MOCK_%s_%d", req.PaymentNo, time.Now().UnixMilli()),
		PayUrl:         fmt.Sprintf("/payment/mock/pay?payment_no=%s", req.PaymentNo),
		Extra:          `{"channel":"mock"}`,
	}, nil
}

func (m *MockChannel) QueryPayment(_ context.Context, _ string) (bool, string, error) {
	// Mock 渠道不支持主动查询，总是返回未支付
	return false, "", nil
}

func (m *MockChannel) Refund(_ context.Context, req *RefundRequest) (*RefundResponse, error) {
	return &RefundResponse{
		ChannelRefundNo: fmt.Sprintf("MOCK_REFUND_%s_%d", req.RefundNo, time.Now().UnixMilli()),
	}, nil
}

func (m *MockChannel) VerifyNotify(_ context.Context, _ map[string]string) (bool, error) {
	// Mock 渠道跳过签名验证
	return true, nil
}

func init() {
	Register(NewMockChannel())
}
