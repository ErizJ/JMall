package channel

// ============================================================
// 微信支付渠道 - 接入指南
// ============================================================
//
// 从 Mock 升级到微信支付的步骤：
//
// 1. 申请微信商户号，获取以下凭证：
//    - AppID（公众号/小程序/APP的应用ID）
//    - MchID（商户号）
//    - APIKey（API密钥，用于签名）
//    - 证书文件（apiclient_cert.pem / apiclient_key.pem）
//
// 2. 在 config 中添加微信支付配置：
//    Wechat:
//      AppId: "wx..."
//      MchId: "..."
//      ApiKey: "..."
//      CertPath: "/path/to/cert.pem"
//      KeyPath: "/path/to/key.pem"
//
// 3. 实现 WechatChannel：
//    - CreatePayment: 调用微信统一下单接口
//      POST https://api.mch.weixin.qq.com/v3/pay/transactions/native (扫码支付)
//      或 /v3/pay/transactions/jsapi (JSAPI支付)
//    - QueryPayment: 调用微信查询订单接口
//      GET https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/{out_trade_no}
//    - Refund: 调用微信退款接口
//      POST https://api.mch.weixin.qq.com/v3/refund/domestic/refunds
//    - VerifyNotify: 验证微信回调签名（HMAC-SHA256）
//
// 4. 推荐使用官方 SDK：
//    go get github.com/wechatpay-apiv3/wechatpay-go
//
// 5. 在 init() 中注册：Register(NewWechatChannel(cfg))
//
// 6. 回调处理：
//    微信回调 POST 到 /payment/notify，body 为 JSON
//    需要用 APIv3 密钥解密通知内容
//    验签通过后调用现有的 PaymentNotifyHandler 逻辑
//
// ============================================================

// TODO: 实现 WechatChannel
// type WechatChannel struct {
//     appId  string
//     mchId  string
//     apiKey string
//     client *wechatpay.Client
// }
