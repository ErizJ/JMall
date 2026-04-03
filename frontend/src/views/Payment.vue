<!--
 * @Description: 支付页面组件 - 现代电商风格
 -->
<template>
  <div class="payment">
    <!-- 步骤条 -->
    <div class="steps-bar">
      <div class="steps-inner">
        <el-steps :active="stepActive" finish-status="success" align-center>
          <el-step title="提交订单" icon="el-icon-s-order"></el-step>
          <el-step title="选择支付" icon="el-icon-wallet"></el-step>
          <el-step title="支付完成" icon="el-icon-circle-check"></el-step>
        </el-steps>
      </div>
    </div>

    <!-- 主要内容 -->
    <div class="pay-container">
      <!-- 支付成功状态 -->
      <div class="result-card" v-if="payStatus === 2">
        <div class="result-icon success">
          <i class="el-icon-check"></i>
        </div>
        <h2>支付成功</h2>
        <p class="result-desc">您的订单已支付成功，感谢您的购买</p>
        <div class="result-info">
          <span>订单编号：{{ orderId }}</span>
          <span>支付金额：<em>¥{{ totalPrice }}</em></span>
        </div>
        <div class="result-actions">
          <router-link to="/order"><el-button type="primary" round>查看订单</el-button></router-link>
          <router-link to="/goods"><el-button round>继续购物</el-button></router-link>
        </div>
      </div>

      <!-- 支付失败状态 -->
      <div class="result-card" v-else-if="payStatus === 3 || payStatus === 4">
        <div class="result-icon fail">
          <i class="el-icon-close"></i>
        </div>
        <h2>{{ payStatus === 3 ? '支付失败' : '支付已关闭' }}</h2>
        <p class="result-desc">{{ payStatus === 3 ? '支付失败，请重新下单' : '该支付已关闭' }}</p>
        <div class="result-actions">
          <router-link to="/order"><el-button type="primary" round>查看订单</el-button></router-link>
        </div>
      </div>

      <!-- 正常支付流程 -->
      <template v-else>
        <!-- 订单摘要 -->
        <div class="summary-card">
          <div class="summary-left">
            <div class="summary-row">
              <span class="summary-label">订单编号</span>
              <span class="summary-value mono">{{ orderId }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">商品件数</span>
              <span class="summary-value">{{ orderItems.length }} 件</span>
            </div>
          </div>
          <div class="summary-right">
            <span class="pay-label">应付金额</span>
            <span class="pay-amount">¥{{ totalPrice }}</span>
          </div>
        </div>

        <!-- 商品清单（可折叠） -->
        <div class="goods-card" v-if="orderItems.length > 0">
          <div class="card-title" @click="showGoods = !showGoods">
            <span><i class="el-icon-goods"></i> 商品清单</span>
            <i :class="showGoods ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"></i>
          </div>
          <transition name="el-zoom-in-top">
            <div class="goods-list" v-show="showGoods">
              <div class="goods-item" v-for="item in orderItems" :key="item.product_id">
                <img :src="$target + item.productImg" class="goods-img" />
                <div class="goods-info">
                  <p class="goods-name">{{ item.productName }}</p>
                  <p class="goods-spec">¥{{ item.price }} × {{ item.num }}</p>
                </div>
                <span class="goods-total">¥{{ (item.price * item.num).toFixed(2) }}</span>
              </div>
            </div>
          </transition>
        </div>

        <!-- 选择支付方式 -->
        <div class="channel-card">
          <div class="card-title"><i class="el-icon-bank-card"></i> 选择支付方式</div>
          <div class="channel-list">
            <div
              :class="['channel-option', selectedChannel === 'mock' ? 'active' : '']"
              @click="selectedChannel = 'mock'"
            >
              <div class="channel-radio"><div class="radio-dot"></div></div>
              <i class="el-icon-monitor channel-icon"></i>
              <div class="channel-text">
                <span class="channel-name">模拟支付</span>
                <span class="channel-desc">开发测试用，点击即完成支付</span>
              </div>
            </div>
            <div
              :class="['channel-option disabled']"
            >
              <div class="channel-radio"></div>
              <i class="el-icon-chat-dot-round channel-icon" style="color:#07c160"></i>
              <div class="channel-text">
                <span class="channel-name">微信支付</span>
                <span class="channel-desc">暂未开放</span>
              </div>
            </div>
            <div
              :class="['channel-option disabled']"
            >
              <div class="channel-radio"></div>
              <i class="el-icon-money channel-icon" style="color:#1677ff"></i>
              <div class="channel-text">
                <span class="channel-name">支付宝</span>
                <span class="channel-desc">暂未开放</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 等待支付状态 -->
        <div class="pending-card" v-if="paymentNo && (payStatus === 0 || payStatus === 1)">
          <i class="el-icon-time pending-icon"></i>
          <div>
            <p class="pending-text">支付单已创建，请确认支付</p>
            <p class="pending-desc">请点击下方"确认支付"按钮完成模拟支付</p>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-bar">
          <router-link to="/order" class="link-back"><i class="el-icon-arrow-left"></i> 返回订单</router-link>
          <div class="action-right">
            <span class="action-total">应付：<em>¥{{ totalPrice }}</em></span>
            <el-button
              v-if="!paymentNo"
              type="primary"
              size="medium"
              :loading="paying"
              :disabled="paying"
              @click="createPayment"
              class="pay-btn"
            >{{ paying ? '创建中...' : '立即支付' }}</el-button>
            <el-button
              v-else-if="payStatus === 0 || payStatus === 1"
              type="primary"
              size="medium"
              :loading="confirming"
              :disabled="confirming"
              @click="confirmMockPay"
              class="pay-btn"
            >{{ confirming ? '确认中...' : '确认支付（模拟）' }}</el-button>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      orderId: null,
      totalPrice: 0,
      orderItems: [],
      showGoods: true,
      selectedChannel: 'mock',
      paymentNo: '',
      payStatus: -1,
      paying: false,
      confirming: false,
      pollTimer: null,
    }
  },
  computed: {
    stepActive() {
      if (this.payStatus === 2) return 3
      if (this.paymentNo) return 2
      return 1
    },
  },
  created() {
    this.orderId = this.$route.query.orderId
    this.totalPrice = this.$route.query.totalPrice || 0
    if (this.$route.query.items) {
      try { this.orderItems = JSON.parse(this.$route.query.items) } catch (e) { this.orderItems = [] }
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
    confirmMockPay() {
      if (this.confirming) return
      this.confirming = true
      this.$axios
        .post('/api/payment/mock/pay', { payment_no: this.paymentNo })
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
    startPolling() {
      this.stopPolling()
      this.pollTimer = setInterval(() => { this.queryStatus() }, 3000)
    },
    stopPolling() {
      if (this.pollTimer) { clearInterval(this.pollTimer); this.pollTimer = null }
    },
    queryStatus() {
      if (!this.paymentNo) return
      this.$axios
        .post('/api/payment/status', { payment_no: this.paymentNo })
        .then((res) => {
          if (res.data.code === '200') {
            this.payStatus = res.data.status
            if (this.payStatus >= 2) this.stopPolling()
          }
        })
    },
  },
}
</script>
<style scoped>
.payment {
  background: var(--bg, #f5f5f5);
  min-height: calc(100vh - 260px);
  padding-bottom: 40px;
}

/* 步骤条 */
.steps-bar {
  background: var(--bg-white, #fff);
  border-bottom: 1px solid var(--border, #e8e8e8);
  padding: 24px 0;
}
.steps-inner {
  max-width: 600px;
  margin: 0 auto;
}

/* 容器 */
.pay-container {
  max-width: var(--content-width, 1226px);
  margin: 24px auto 0;
  padding: 0 20px;
}

/* 订单摘要 */
.summary-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  padding: 24px 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.summary-row {
  display: flex;
  gap: 12px;
  padding: 4px 0;
  font-size: 14px;
}
.summary-label { color: var(--text-muted, #999); }
.summary-value { color: var(--text, #333); }
.mono { font-family: monospace; }
.summary-right { text-align: right; }
.pay-label { font-size: 14px; color: var(--text-muted, #999); display: block; margin-bottom: 4px; }
.pay-amount { font-size: 32px; font-weight: 700; color: var(--primary, #ff6700); }

/* 商品清单 */
.goods-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}
.card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 30px;
  font-size: 15px;
  font-weight: 500;
  color: var(--text, #333);
  cursor: pointer;
  border-bottom: 1px solid var(--border, #f0f0f0);
}
.card-title i:first-child { color: var(--primary, #ff6700); margin-right: 6px; }
.goods-list { padding: 0 30px; }
.goods-item {
  display: flex;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #f8f8f8;
}
.goods-item:last-child { border-bottom: none; }
.goods-img {
  width: 56px;
  height: 56px;
  border-radius: 6px;
  object-fit: contain;
  background: #f9f9f9;
  border: 1px solid #f0f0f0;
  flex-shrink: 0;
  margin-right: 14px;
}
.goods-info { flex: 1; min-width: 0; }
.goods-name {
  font-size: 14px;
  color: var(--text, #333);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.goods-spec { font-size: 13px; color: var(--text-muted, #999); margin-top: 4px; }
.goods-total { font-size: 15px; font-weight: 600; color: var(--primary, #ff6700); flex-shrink: 0; margin-left: 20px; }

/* 支付方式 */
.channel-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}
.channel-card .card-title {
  cursor: default;
}
.channel-list { padding: 8px 30px 20px; }
.channel-option {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 20px;
  border: 2px solid var(--border, #e8e8e8);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 10px;
}
.channel-option:last-child { margin-bottom: 0; }
.channel-option:hover:not(.disabled) { border-color: var(--primary-light, #ff8533); }
.channel-option.active {
  border-color: var(--primary, #ff6700);
  background: rgba(255, 103, 0, 0.03);
}
.channel-option.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.channel-radio {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.channel-option.active .channel-radio {
  border-color: var(--primary, #ff6700);
}
.radio-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--primary, #ff6700);
}
.channel-icon { font-size: 28px; color: var(--primary, #ff6700); flex-shrink: 0; }
.channel-text { flex: 1; }
.channel-name { font-size: 15px; color: var(--text, #333); font-weight: 500; display: block; }
.channel-desc { font-size: 12px; color: var(--text-muted, #999); margin-top: 2px; display: block; }

/* 等待支付 */
.pending-card {
  background: #fff8f0;
  border: 1px solid #ffe4cc;
  border-radius: 8px;
  padding: 20px 30px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.pending-icon { font-size: 32px; color: #e6a23c; }
.pending-text { font-size: 15px; color: var(--text, #333); font-weight: 500; }
.pending-desc { font-size: 13px; color: var(--text-muted, #999); margin-top: 4px; }

/* 操作栏 */
.action-bar {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  padding: 20px 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.link-back {
  font-size: 14px;
  color: var(--text-muted, #999);
  transition: color 0.2s;
}
.link-back:hover { color: var(--primary, #ff6700); }
.action-right {
  display: flex;
  align-items: center;
  gap: 24px;
}
.action-total {
  font-size: 14px;
  color: var(--text, #333);
}
.action-total em {
  font-style: normal;
  font-size: 24px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}
.pay-btn {
  height: 44px;
  padding: 0 40px;
  font-size: 16px;
  border-radius: 22px;
  background: var(--primary, #ff6700);
  border-color: var(--primary, #ff6700);
}
.pay-btn:hover {
  background: var(--primary-dark, #e55d00);
  border-color: var(--primary-dark, #e55d00);
}

/* 结果页 */
.result-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  padding: 60px 40px;
  text-align: center;
}
.result-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}
.result-icon i { font-size: 40px; color: #fff; }
.result-icon.success { background: #67c23a; }
.result-icon.fail { background: #f56c6c; }
.result-card h2 { font-size: 24px; color: var(--text, #333); font-weight: 600; margin-bottom: 8px; }
.result-desc { font-size: 14px; color: var(--text-muted, #999); margin-bottom: 24px; }
.result-info {
  display: flex;
  justify-content: center;
  gap: 32px;
  font-size: 14px;
  color: var(--text-secondary, #666);
  margin-bottom: 32px;
}
.result-info em { font-style: normal; color: var(--primary, #ff6700); font-weight: 600; }
.result-actions { display: flex; justify-content: center; gap: 16px; }

/* Element steps 主题色覆盖 */
.payment >>> .el-step__head.is-finish { color: var(--primary, #ff6700); border-color: var(--primary, #ff6700); }
.payment >>> .el-step__title.is-finish { color: var(--primary, #ff6700); }
.payment >>> .el-step__head.is-process { color: var(--primary, #ff6700); border-color: var(--primary, #ff6700); }
.payment >>> .el-step__title.is-process { color: var(--primary, #ff6700); font-weight: 600; }
</style>
