<!--
 * @Description: 支付页面组件
 -->
<template>
  <div class="payment">
    <!-- 头部 -->
    <div class="payment-header">
      <div class="header-content">
        <p>
          <i class="el-icon-wallet"></i>
        </p>
        <p>订单支付</p>
      </div>
    </div>
    <!-- 头部END -->

    <!-- 主要内容 -->
    <div class="content">
      <!-- 订单信息 -->
      <div class="section-order">
        <p class="title">订单信息</p>
        <div class="order-info">
          <div class="info-row">
            <span class="label">订单编号：</span>
            <span class="value">{{ orderId }}</span>
          </div>
          <div class="info-row">
            <span class="label">商品件数：</span>
            <span class="value">{{ orderItems.length }} 件</span>
          </div>
          <div class="info-row total">
            <span class="label">应付金额：</span>
            <span class="value price">¥{{ totalPrice }}</span>
          </div>
        </div>
      </div>
      <!-- 订单信息END -->

      <!-- 商品清单 -->
      <div class="section-goods" v-if="orderItems.length > 0">
        <p class="title">商品清单</p>
        <div class="goods-list">
          <ul>
            <li v-for="item in orderItems" :key="item.product_id">
              <img :src="$target + item.productImg" />
              <span class="pro-name">{{ item.productName }}</span>
              <span class="pro-price">{{ item.price }}元 x {{ item.num }}</span>
              <span class="pro-total">{{ item.price * item.num }}元</span>
            </li>
          </ul>
        </div>
      </div>
      <!-- 商品清单END -->

      <!-- 选择支付方式 -->
      <div class="section-channel">
        <p class="title">支付方式</p>
        <div class="channel-list">
          <div
            :class="['channel-item', selectedChannel === 'mock' ? 'active' : '']"
            @click="selectedChannel = 'mock'"
          >
            <i class="el-icon-monitor"></i>
            <span>模拟支付</span>
            <p class="desc">开发测试用，点击即完成支付</p>
          </div>
          <div
            :class="['channel-item', selectedChannel === 'wechat' ? 'active' : '']"
            @click="selectedChannel = 'wechat'"
          >
            <i class="el-icon-chat-dot-round"></i>
            <span>微信支付</span>
            <p class="desc">暂未开放</p>
          </div>
          <div
            :class="['channel-item', selectedChannel === 'alipay' ? 'active' : '']"
            @click="selectedChannel = 'alipay'"
          >
            <i class="el-icon-money"></i>
            <span>支付宝</span>
            <p class="desc">暂未开放</p>
          </div>
        </div>
      </div>
      <!-- 选择支付方式END -->

      <!-- 支付状态 -->
      <div class="section-status" v-if="paymentNo">
        <div :class="['status-box', statusClass]">
          <i :class="statusIcon"></i>
          <p class="status-text">{{ statusText }}</p>
          <p class="status-desc" v-if="statusDesc">{{ statusDesc }}</p>
        </div>
      </div>
      <!-- 支付状态END -->

      <!-- 操作按钮 -->
      <div class="section-bar">
        <div class="btn">
          <router-link to="/order" class="btn-base btn-return">查看订单</router-link>
          <a
            v-if="!paymentNo"
            href="javascript:void(0);"
            :class="['btn-base', paying ? 'btn-disabled' : 'btn-primary']"
            @click="createPayment"
          >
            <i v-if="paying" class="el-icon-loading"></i>
            {{ paying ? '支付中...' : '立即支付' }}
          </a>
          <a
            v-else-if="payStatus === 0 || payStatus === 1"
            href="javascript:void(0);"
            :class="['btn-base', confirming ? 'btn-disabled' : 'btn-primary']"
            @click="confirmMockPay"
          >
            <i v-if="confirming" class="el-icon-loading"></i>
            {{ confirming ? '确认中...' : '确认支付（模拟）' }}
          </a>
          <router-link
            v-else-if="payStatus === 2"
            to="/order"
            class="btn-base btn-primary"
          >支付成功，查看订单</router-link>
        </div>
      </div>
      <!-- 操作按钮END -->
    </div>
    <!-- 主要内容END -->
  </div>
</template>

<script>
export default {
  data() {
    return {
      orderId: null,
      totalPrice: 0,
      orderItems: [],
      selectedChannel: 'mock',
      paymentNo: '',
      payStatus: -1, // -1未创建 0待支付 1支付中 2成功 3失败 4关闭 5退款
      paying: false,
      confirming: false,
      pollTimer: null,
    }
  },
  computed: {
    statusClass() {
      if (this.payStatus === 2) return 'success'
      if (this.payStatus === 3 || this.payStatus === 4) return 'fail'
      return 'pending'
    },
    statusIcon() {
      if (this.payStatus === 2) return 'el-icon-success'
      if (this.payStatus === 3 || this.payStatus === 4) return 'el-icon-error'
      return 'el-icon-time'
    },
    statusText() {
      const map = {
        0: '等待支付',
        1: '支付处理中',
        2: '支付成功',
        3: '支付失败',
        4: '支付已关闭',
        5: '已退款',
      }
      return map[this.payStatus] || '未知状态'
    },
    statusDesc() {
      if (this.payStatus === 0 && this.selectedChannel === 'mock') {
        return '请点击下方"确认支付"按钮完成模拟支付'
      }
      if (this.payStatus === 2) {
        return '您的订单已支付成功，感谢您的购买'
      }
      if (this.payStatus === 3) {
        return '支付失败，请重新下单'
      }
      return ''
    },
  },
  created() {
    this.orderId = this.$route.query.orderId
    this.totalPrice = this.$route.query.totalPrice || 0
    // 从路由参数恢复商品列表
    if (this.$route.query.items) {
      try {
        this.orderItems = JSON.parse(this.$route.query.items)
      } catch (e) {
        this.orderItems = []
      }
    }
    if (!this.orderId) {
      this.notifyError('订单信息缺失，请重新下单')
      this.$router.push({ path: '/shoppingCart' })
    }
  },
  beforeDestroy() {
    this.stopPolling()
  },
  methods: {
    // 创建支付单
    createPayment() {
      if (this.paying) return
      this.paying = true
      this.$axios
        .post('/api/payment/create', {
          order_id: Number(this.orderId),
          channel: this.selectedChannel,
        })
        .then((res) => {
          this.paying = false
          if (res.data.code === '200') {
            this.paymentNo = res.data.payment_no
            this.payStatus = 0
            this.notifySucceed('支付单创建成功')
            this.startPolling()
          } else {
            this.notifyError(res.data.msg || '创建支付单失败')
          }
        })
        .catch(() => {
          this.paying = false
          this.notifyError('网络异常，请稍后重试')
        })
    },
    // 模拟支付确认
    confirmMockPay() {
      if (this.confirming) return
      this.confirming = true
      this.$axios
        .post('/api/payment/mock/pay', {
          payment_no: this.paymentNo,
        })
        .then((res) => {
          this.confirming = false
          if (res.data.code === '200') {
            this.payStatus = 2
            this.stopPolling()
            this.notifySucceed('支付成功')
          } else {
            this.notifyError(res.data.msg || '支付确认失败')
          }
        })
        .catch(() => {
          this.confirming = false
          this.notifyError('网络异常，请稍后重试')
        })
    },
    // 轮询支付状态
    startPolling() {
      this.stopPolling()
      this.pollTimer = setInterval(() => {
        this.queryStatus()
      }, 3000)
    },
    stopPolling() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
        this.pollTimer = null
      }
    },
    queryStatus() {
      if (!this.paymentNo) return
      this.$axios
        .post('/api/payment/status', {
          payment_no: this.paymentNo,
        })
        .then((res) => {
          if (res.data.code === '200') {
            this.payStatus = res.data.status
            // 终态停止轮询
            if (this.payStatus >= 2) {
              this.stopPolling()
            }
          }
        })
    },
  },
}
</script>

<style scoped>
.payment {
  background-color: #f5f5f5;
  padding-bottom: 40px;
  min-height: 600px;
}

/* 头部 */
.payment .payment-header {
  background-color: #fff;
  border-bottom: 2px solid #409EFF;
  margin-bottom: 20px;
}
.payment .payment-header .header-content {
  width: 1225px;
  margin: 0 auto;
  height: 80px;
}
.payment .payment-header .header-content p {
  float: left;
  font-size: 28px;
  line-height: 80px;
  color: #424242;
  margin-right: 20px;
}
.payment .payment-header .header-content p i {
  font-size: 45px;
  color: #409EFF;
  line-height: 80px;
}

/* 主要内容 */
.payment .content {
  width: 1225px;
  margin: 0 auto;
  background-color: #fff;
  padding: 40px 48px;
}

/* 通用标题 */
.payment .content .title {
  color: #333;
  font-size: 18px;
  line-height: 20px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e0e0e0;
}

/* 订单信息 */
.payment .section-order {
  margin-bottom: 30px;
}
.payment .order-info .info-row {
  line-height: 36px;
  font-size: 14px;
  color: #616161;
}
.payment .order-info .info-row .label {
  display: inline-block;
  width: 100px;
  color: #757575;
}
.payment .order-info .info-row.total {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #e0e0e0;
}
.payment .order-info .info-row .price {
  font-size: 28px;
  color: #409EFF;
  font-weight: 600;
}

/* 商品清单 */
.payment .section-goods {
  margin-bottom: 30px;
}
.payment .section-goods .goods-list {
  padding: 5px 0;
}
.payment .section-goods .goods-list li {
  padding: 10px 0;
  color: #424242;
  overflow: hidden;
  border-bottom: 1px solid #f5f5f5;
}
.payment .section-goods .goods-list li img {
  float: left;
  width: 40px;
  height: 40px;
  margin-right: 12px;
  border-radius: 4px;
}
.payment .section-goods .goods-list li .pro-name {
  float: left;
  width: 550px;
  line-height: 40px;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.payment .section-goods .goods-list li .pro-price {
  float: left;
  width: 200px;
  text-align: center;
  line-height: 40px;
  color: #757575;
  font-size: 14px;
}
.payment .section-goods .goods-list li .pro-total {
  float: right;
  width: 150px;
  text-align: right;
  color: #409EFF;
  line-height: 40px;
  font-size: 14px;
}

/* 支付方式 */
.payment .section-channel {
  margin-bottom: 30px;
}
.payment .channel-list {
  display: flex;
  gap: 20px;
}
.payment .channel-item {
  flex: 1;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 24px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}
.payment .channel-item:hover {
  border-color: #b0b0b0;
}
.payment .channel-item.active {
  border-color: #409EFF;
  background-color: #f0f7ff;
}
.payment .channel-item i {
  font-size: 36px;
  color: #409EFF;
  display: block;
  margin-bottom: 10px;
}
.payment .channel-item span {
  font-size: 16px;
  color: #333;
  font-weight: 500;
}
.payment .channel-item .desc {
  font-size: 12px;
  color: #999;
  margin-top: 6px;
}

/* 支付状态 */
.payment .section-status {
  margin-bottom: 30px;
}
.payment .status-box {
  text-align: center;
  padding: 30px;
  border-radius: 8px;
}
.payment .status-box.pending {
  background-color: #fdf6ec;
}
.payment .status-box.success {
  background-color: #f0f9eb;
}
.payment .status-box.fail {
  background-color: #fef0f0;
}
.payment .status-box i {
  font-size: 48px;
  margin-bottom: 12px;
}
.payment .status-box.pending i {
  color: #e6a23c;
}
.payment .status-box.success i {
  color: #67c23a;
}
.payment .status-box.fail i {
  color: #f56c6c;
}
.payment .status-box .status-text {
  font-size: 20px;
  color: #333;
  font-weight: 500;
}
.payment .status-box .status-desc {
  font-size: 14px;
  color: #757575;
  margin-top: 8px;
}

/* 操作按钮 */
.payment .section-bar {
  padding-top: 20px;
  border-top: 2px solid #f5f5f5;
  overflow: hidden;
}
.payment .section-bar .btn {
  float: right;
}
.payment .section-bar .btn .btn-base {
  float: left;
  margin-left: 20px;
  width: 180px;
  height: 42px;
  border: 1px solid #b0b0b0;
  font-size: 14px;
  line-height: 42px;
  text-align: center;
  border-radius: 4px;
  display: inline-block;
}
.payment .section-bar .btn .btn-return {
  color: #757575;
  border-color: #d0d0d0;
}
.payment .section-bar .btn .btn-return:hover {
  color: #409EFF;
  border-color: #409EFF;
}
.payment .section-bar .btn .btn-primary {
  background: #409EFF;
  border-color: #409EFF;
  color: #fff;
}
.payment .section-bar .btn .btn-primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
}
.payment .section-bar .btn .btn-disabled {
  background: #e0e0e0;
  border-color: #e0e0e0;
  color: #b0b0b0;
  cursor: not-allowed;
}
</style>
