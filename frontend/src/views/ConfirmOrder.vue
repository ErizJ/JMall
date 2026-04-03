<!--
 * @Description: 确认订单页面组件 - 现代电商风格
 -->
<template>
  <div class="confirmOrder">
    <!-- 步骤条 -->
    <div class="steps-bar">
      <div class="steps-inner">
        <el-steps :active="1" finish-status="success" align-center>
          <el-step title="购物车" icon="el-icon-shopping-cart-2"></el-step>
          <el-step title="确认订单" icon="el-icon-s-order"></el-step>
          <el-step title="支付" icon="el-icon-wallet"></el-step>
        </el-steps>
      </div>
    </div>

    <!-- 主要内容 -->
    <div class="order-container">
      <!-- 收货地址 -->
      <div class="section-card">
        <div class="card-title"><i class="el-icon-location-outline"></i> 收货地址</div>
        <div class="address-list">
          <div
            :class="['address-item', { active: item.id === confirmAddress }]"
            v-for="item in address"
            :key="item.id"
            @click="confirmAddress = item.id"
          >
            <div class="addr-check">
              <i v-if="item.id === confirmAddress" class="el-icon-success"></i>
            </div>
            <div class="addr-info">
              <div class="addr-top">
                <span class="addr-name">{{ item.name }}</span>
                <span class="addr-phone">{{ item.phone }}</span>
              </div>
              <p class="addr-detail">{{ item.address }}</p>
            </div>
          </div>
          <div class="address-item add-new">
            <i class="el-icon-plus"></i>
            <span>添加新地址</span>
          </div>
        </div>
      </div>

      <!-- 商品清单 -->
      <div class="section-card">
        <div class="card-title"><i class="el-icon-goods"></i> 商品清单</div>
        <div class="product-list">
          <div class="product-item" v-for="item in getCheckGoods" :key="item.id">
            <img :src="$target + item.productImg" class="prod-img" />
            <div class="prod-info">
              <p class="prod-name">{{ item.productName }}</p>
            </div>
            <span class="prod-price">¥{{ item.price }}</span>
            <span class="prod-qty">× {{ item.num }}</span>
            <span class="prod-subtotal">¥{{ (item.price * item.num).toFixed(2) }}</span>
          </div>
        </div>
      </div>

      <!-- 配送与发票 -->
      <div class="section-card">
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">配送方式</span>
            <span class="info-value"><i class="el-icon-truck"></i> 包邮</span>
          </div>
          <div class="info-item">
            <span class="info-label">发票信息</span>
            <span class="info-value">电子发票 · 个人 · 商品明细</span>
          </div>
        </div>
      </div>

      <!-- 结算汇总 -->
      <div class="section-card settle-card">
        <div class="settle-rows">
          <div class="settle-row">
            <span>商品件数</span>
            <span>{{ getCheckNum }} 件</span>
          </div>
          <div class="settle-row">
            <span>商品总价</span>
            <span>¥{{ getTotalPrice.toFixed(2) }}</span>
          </div>
          <div class="settle-row" v-if="sale > 0">
            <span>满减优惠</span>
            <span class="discount">-¥{{ sale }}</span>
          </div>
          <div class="settle-row">
            <span>运费</span>
            <span>免运费</span>
          </div>
          <div class="settle-row total">
            <span>应付总额</span>
            <span class="total-price">¥{{ (getTotalPrice - sale).toFixed(2) }}</span>
          </div>
        </div>
      </div>

      <!-- 底部操作栏 -->
      <div class="action-bar">
        <router-link to="/shoppingCart" class="link-back">
          <i class="el-icon-arrow-left"></i> 返回购物车
        </router-link>
        <div class="action-right">
          <div class="action-summary">
            <span class="action-count">共 {{ getCheckNum }} 件商品</span>
            <span class="action-total">应付：<em>¥{{ (getTotalPrice - sale).toFixed(2) }}</em></span>
          </div>
          <el-button type="primary" size="medium" class="submit-btn" @click="addOrder">提交订单</el-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import { mapGetters, mapActions } from 'vuex'
export default {
  data() {
    return {
      sale: 0,
      confirmAddress: 1,
      address: [
        { id: 1, name: '张三', phone: '189****2638', address: '广东 广州市 海珠区 某某大学' },
        { id: 2, name: '李四', phone: '159****3182', address: '广东 深圳市 南山区 某某科技园' },
      ],
    }
  },
  created() {
    if (this.getCheckNum < 1) {
      this.notifyError('请勾选商品后再结算')
      this.$router.push({ path: '/shoppingCart' })
    }
    if (this.getTotalPrice >= 3000) {
      this.sale = 300
    } else if (this.getTotalPrice >= 2000) {
      this.sale = 200
    }
  },
  computed: {
    ...mapGetters(['getCheckNum', 'getTotalPrice', 'getCheckGoods']),
  },
  methods: {
    ...mapActions(['deleteShoppingCart']),
    addOrder() {
      const items = this.getCheckGoods.map((g) => ({
        product_id: g.productID,
        product_num: g.num,
        product_price: g.price,
      }))
      this.$axios
        .post('/api/user/order/addOrder', {
          user_id: this.$store.getters.getUser.user_id,
          items: items,
        })
        .then((res) => {
          const products = this.getCheckGoods
          if (res.data.code === '200') {
            for (let i = 0; i < products.length; i++) {
              this.deleteShoppingCart(products[i].id)
            }
            this.notifySucceed('订单创建成功，请完成支付')
            const orderItems = products.map((p) => ({
              productImg: p.productImg,
              productName: p.productName,
              price: p.price,
              num: p.num,
            }))
            this.$router.push({
              path: '/payment',
              query: {
                orderId: res.data.order_id,
                totalPrice: this.getTotalPrice - this.sale,
                items: JSON.stringify(orderItems),
              },
            })
          } else {
            this.notifyError(res.data.msg || '下单失败')
          }
        })
        .catch((err) => Promise.reject(err))
    },
  },
}
</script>
<style scoped>
.confirmOrder {
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
  max-width: 500px;
  margin: 0 auto;
}

/* 容器 */
.order-container {
  max-width: var(--content-width, 1226px);
  margin: 24px auto 0;
  padding: 0 20px;
}

/* 通用卡片 */
.section-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}
.card-title {
  padding: 16px 30px;
  font-size: 15px;
  font-weight: 500;
  color: var(--text, #333);
  border-bottom: 1px solid var(--border, #f0f0f0);
}
.card-title i { color: var(--primary, #ff6700); margin-right: 6px; }

/* 收货地址 */
.address-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 20px 30px;
}
.address-item {
  width: 240px;
  border: 2px solid var(--border, #e8e8e8);
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  gap: 10px;
}
.address-item:hover { border-color: var(--primary-light, #ff8533); }
.address-item.active { border-color: var(--primary, #ff6700); background: rgba(255, 103, 0, 0.02); }
.addr-check {
  width: 20px;
  flex-shrink: 0;
  padding-top: 2px;
}
.addr-check i { color: var(--primary, #ff6700); font-size: 18px; }
.addr-top { display: flex; gap: 12px; align-items: center; margin-bottom: 6px; }
.addr-name { font-size: 15px; font-weight: 500; color: var(--text, #333); }
.addr-phone { font-size: 13px; color: var(--text-muted, #999); }
.addr-detail { font-size: 13px; color: var(--text-secondary, #666); line-height: 1.5; }
.add-new {
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 6px;
  color: var(--text-muted, #999);
  border-style: dashed;
}
.add-new i { font-size: 24px; }
.add-new span { font-size: 13px; }

/* 商品清单 */
.product-list { padding: 0 30px; }
.product-item {
  display: flex;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f8f8f8;
}
.product-item:last-child { border-bottom: none; }
.prod-img {
  width: 64px;
  height: 64px;
  border-radius: 6px;
  object-fit: contain;
  background: #f9f9f9;
  border: 1px solid #f0f0f0;
  flex-shrink: 0;
  margin-right: 16px;
}
.prod-info { flex: 1; min-width: 0; }
.prod-name {
  font-size: 14px;
  color: var(--text, #333);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.prod-price { width: 90px; text-align: center; font-size: 14px; color: var(--text-secondary, #666); flex-shrink: 0; }
.prod-qty { width: 60px; text-align: center; font-size: 14px; color: var(--text-muted, #999); flex-shrink: 0; }
.prod-subtotal { width: 100px; text-align: right; font-size: 15px; font-weight: 600; color: var(--primary, #ff6700); flex-shrink: 0; }

/* 配送与发票 */
.info-grid { padding: 20px 30px; }
.info-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  font-size: 14px;
}
.info-label { color: var(--text-muted, #999); width: 80px; flex-shrink: 0; }
.info-value { color: var(--text-secondary, #666); }
.info-value i { color: var(--primary, #ff6700); margin-right: 4px; }

/* 结算汇总 */
.settle-card { padding: 20px 30px; }
.settle-rows { max-width: 300px; margin-left: auto; }
.settle-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  color: var(--text-secondary, #666);
}
.settle-row .discount { color: #f56c6c; }
.settle-row.total {
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--border, #f0f0f0);
}
.total-price {
  font-size: 24px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}

/* 底部操作栏 */
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
.action-summary { text-align: right; }
.action-count { font-size: 13px; color: var(--text-muted, #999); display: block; }
.action-total { font-size: 14px; color: var(--text, #333); }
.action-total em {
  font-style: normal;
  font-size: 24px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}
.submit-btn {
  height: 44px;
  padding: 0 40px;
  font-size: 16px;
  border-radius: 22px;
  background: var(--primary, #ff6700);
  border-color: var(--primary, #ff6700);
}
.submit-btn:hover {
  background: var(--primary-dark, #e55d00);
  border-color: var(--primary-dark, #e55d00);
}

/* Element steps 主题色覆盖 */
.confirmOrder >>> .el-step__head.is-finish { color: var(--primary, #ff6700); border-color: var(--primary, #ff6700); }
.confirmOrder >>> .el-step__title.is-finish { color: var(--primary, #ff6700); }
.confirmOrder >>> .el-step__head.is-process { color: var(--primary, #ff6700); border-color: var(--primary, #ff6700); }
.confirmOrder >>> .el-step__title.is-process { color: var(--primary, #ff6700); font-weight: 600; }
</style>
