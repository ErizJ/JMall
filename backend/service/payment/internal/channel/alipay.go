package channel

// ============================================================
// 支付宝渠道 - 接入指南
// ============================================================
//
// 从 Mock 升级到支付宝的步骤：
//
// 1. 在支付宝开放平台创建应用，获取：
//    - AppID
//    - 应用私钥（RSA2）
//    - 支付宝公钥（用于验签）
//
// 2. 在 config 中添加支付宝配置：
//    Alipay:
//      AppId: "..."
//      PrivateKey: "..."
//      AlipayPublicKey: "..."
//      IsProduction: false  # 沙箱/生产切换
//
// 3. 实现 AlipayChannel：
//    - CreatePayment: 调用 alipay.trade.page.pay（PC网页支付）
//      或 alipay.trade.wap.pay（手机网页支付）
//      或 alipay.trade.app.pay（APP支付）
//    - QueryPayment: 调用 alipay.trade.query
//    - Refund: 调用 alipay.trade.refund
//    - VerifyNotify: 使用支付宝公钥验证异步通知签名（RSA2）
//
// 4. 推荐使用社区 SDK：
//    go get github.com/smartwalle/alipay/v3
//
// 5. 在 init() 中注册：Register(NewAlipayChannel(cfg))
//
// 6. 回调处理：
//    支付宝异步通知 POST 到 /payment/notify
//    参数为 form-urlencoded 格式
//    验签通过后走统一的回调处理逻辑
//
// ============================================================

// TODO: 实现 AlipayChannel
// type AlipayChannel struct {
//     appId     string
//     client    *alipay.Client
// }
