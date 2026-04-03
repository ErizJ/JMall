<!--
 * @Description: 我的订单页面
 -->
<template>
  <div class="order-page">
    <div class="order-wrap">
      <div class="page-title">
        <h1><i class="el-icon-document"></i> 我的订单</h1>
      </div>

      <!-- 有订单 -->
      <template v-if="orders.length > 0">
        <div class="order-card" v-for="order in orders" :key="order.order_id">
          <!-- 订单头 -->
          <div class="card-head">
            <div class="head-left">
              <span class="order-no">订单号：{{ order.order_id }}</span>
              <el-tag size="mini" :type="statusType(order.status)">{{ statusText(order.status) }}</el-tag>
            </div>
            <span class="order-time"><i class="el-icon-time"></i> {{ order.order_time }}</span>
          </div>

          <!-- 商品列表 -->
          <div class="card-body">
            <div class="item-row" v-for="item in order.items" :key="item.id">
              <router-link :to="{ path: '/goods/details', query: { productID: item.product_id } }" class="item-link">
                <img :src="$target + item.product_img" class="item-img" />
                <span class="item-name">{{ item.product_name }}</span>
              </router-link>
              <span class="item-price">¥{{ item.product_price }}</span>
              <span class="item-qty">x{{ item.product_num }}</span>
              <span class="item-subtotal">¥{{ (item.product_price * item.product_num).toFixed(2) }}</span>
            </div>
          </div>

          <!-- 订单尾 -->
          <div class="card-foot">
            <span class="foot-count">共 {{ order.item_count }} 件</span>
            <div class="foot-right">
              <span class="foot-total">合计：<em>¥{{ order.total_amount.toFixed(2) }}</em></span>
              <router-link
                v-if="order.status === 0"
                :to="{ path: '/payment', query: { orderId: order.order_id, totalPrice: order.total_amount } }"
                class="btn-pay"
              >去支付</router-link>
            </div>
          </div>
        </div>
      </template>

      <!-- 空订单 -->
      <div v-else class="order-empty">
        <i class="el-icon-document empty-icon"></i>
        <h2>还没有订单</h2>
        <p>去挑选心仪的商品吧</p>
        <router-link to="/goods"><el-button type="primary" round>去逛逛</el-button></router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return { orders: [] }
  },
  activated() {
    this.$axios
      .post('/api/user/order/getOrder', {
        user_id: this.$store.getters.getUser.user_id,
      })
      .then((res) => {
        if (res.data.code === '200' || res.data.code === '001') {
          this.orders = res.data.orders || []
        }
      })
      .catch(() => {})
  },
  methods: {
    statusText(s) { return { 0: '待支付', 1: '已支付', 2: '已取消', 3: '已退款' }[s] || '未知' },
    statusType(s) { return { 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info' },
  },
}
</script>

<style scoped>
.order-page { background: var(--bg, #f5f5f5); min-height: calc(100vh - 260px); padding: 24px 0 40px; }
.order-wrap { max-width: var(--content-width, 1226px); margin: 0 auto; padding: 0 20px; }
.page-title { margin-bottom: 20px; }
.page-title h1 { font-size: 22px; font-weight: 600; color: #333; }
.page-title h1 i { color: var(--primary, #ff6700); margin-right: 4px; }

.order-card { background: #fff; border-radius: 8px; border: 1px solid #f0f0f0; margin-bottom: 16px; overflow: hidden; }
.order-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.05); }

.card-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 20px; background: #fafafa; border-bottom: 1px solid #f0f0f0;
}
.head-left { display: flex; align-items: center; gap: 10px; }
.order-no { font-size: 13px; color: #333; font-weight: 500; font-family: monospace; }
.order-time { font-size: 12px; color: #999; }

.card-body { padding: 0 20px; }
.item-row { display: flex; align-items: center; padding: 14px 0; border-bottom: 1px solid #f8f8f8; }
.item-row:last-child { border-bottom: none; }
.item-link { display: flex; align-items: center; gap: 12px; flex: 1; min-width: 0; color: inherit; }
.item-link:hover .item-name { color: var(--primary, #ff6700); }
.item-img { width: 56px; height: 56px; object-fit: contain; border-radius: 6px; background: #f9f9f9; border: 1px solid #f0f0f0; flex-shrink: 0; }
.item-name { font-size: 13px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.item-price { width: 80px; text-align: center; font-size: 13px; color: #666; flex-shrink: 0; }
.item-qty { width: 50px; text-align: center; font-size: 13px; color: #999; flex-shrink: 0; }
.item-subtotal { width: 100px; text-align: right; font-size: 14px; font-weight: 600; color: var(--primary, #ff6700); flex-shrink: 0; }

.card-foot {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 20px; background: #fafafa; border-top: 1px solid #f0f0f0; font-size: 13px; color: #999;
}
.foot-right { display: flex; align-items: center; gap: 16px; }
.foot-total em { font-style: normal; font-size: 20px; font-weight: 700; color: var(--primary, #ff6700); }
.btn-pay {
  display: inline-block; padding: 0 20px; height: 32px; line-height: 32px;
  background: #f56c6c; color: #fff; border-radius: 16px; font-size: 13px;
}
.btn-pay:hover { background: #f78989; }

.order-empty { text-align: center; padding: 80px 0; }
.empty-icon { font-size: 64px; color: #ddd; }
.order-empty h2 { font-size: 18px; color: #999; font-weight: 400; margin: 12px 0 6px; }
.order-empty p { font-size: 14px; color: #bbb; margin-bottom: 20px; }
</style>
