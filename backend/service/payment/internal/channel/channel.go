// Package channel 定义支付渠道抽象接口。
//
// 设计理由（Strategy 模式）：
//   - 所有支付渠道实现统一接口，业务逻辑层不感知具体渠道差异
//   - 新增渠道只需实现 PayChannel 接口并注册到 Registry
//   - Mock 渠道用于开发测试，生产环境切换只需更改配置
//   - 从 Mock 升级到真实支付，只需实现对应的 wechat/alipay channel
package channel

import (
	"context"
	"fmt"
	"sync"
)

// PayRequest 统一支付请求
type PayRequest struct {
	PaymentNo string // 支付流水号
	OrderId   int64  // 业务订单号
	Amount    int64  // 金额（分）
	Subject   string // 商品描述
	NotifyUrl string // 回调地址
}

// PayResponse 统一支付响应
type PayResponse struct {
	ChannelTradeNo string // 渠道预交易号（部分渠道在创建时返回）
	PayUrl         string // 支付跳转URL / 二维码URL
	Extra          string // 扩展信息（JSON）
}

// RefundRequest 统一退款请求
type RefundRequest struct {
	PaymentNo      string
	RefundNo       string
	ChannelTradeNo string // 原支付渠道交易号
	TotalAmount    int64  // 原支付金额（分）
	RefundAmount   int64  // 退款金额（分）
	Reason         string
}

// RefundResponse 统一退款响应
type RefundResponse struct {
	ChannelRefundNo string // 渠道退款单号
}

// PayChannel 支付渠道接口
// 每个渠道（mock/wechat/alipay）实现此接口
type PayChannel interface {
	// Name 返回渠道标识
	Name() string
	// CreatePayment 发起支付（调用第三方预下单接口）
	CreatePayment(ctx context.Context, req *PayRequest) (*PayResponse, error)
	// QueryPayment 主动查询支付结果（用于补偿场景）
	QueryPayment(ctx context.Context, paymentNo string) (success bool, channelTradeNo string, err error)
	// Refund 发起退款
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
	// VerifyNotify 验证回调签名（生产环境必须）
	VerifyNotify(ctx context.Context, params map[string]string) (bool, error)
}

// Registry 渠道注册中心（单例）
var (
	registry = make(map[string]PayChannel)
	mu       sync.RWMutex
)

// Register 注册支付渠道
func Register(ch PayChannel) {
	mu.Lock()
	defer mu.Unlock()
	registry[ch.Name()] = ch
}

// Get 获取支付渠道
func Get(name string) (PayChannel, error) {
	mu.RLock()
	defer mu.RUnlock()
	ch, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported payment channel: %s", name)
	}
	return ch, nil
}
