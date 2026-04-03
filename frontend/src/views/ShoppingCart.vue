<!--
 * @Description: 购物车页面
 -->
<template>
  <div class="cart-page">
    <div class="cart-wrap">
      <!-- 页面标题 -->
      <div class="page-title">
        <h1><i class="el-icon-shopping-cart-2"></i> 我的购物车</h1>
        <p class="promo-tip" v-if="getShoppingCart.length > 0">
          <i class="el-icon-present"></i> 满2000减200，满3000减300
        </p>
      </div>

      <!-- 有商品时 -->
      <template v-if="getShoppingCart.length > 0">
        <!-- 表头 -->
        <div class="cart-table-header">
          <div class="col-check">
            <el-checkbox v-model="isAllCheck">全选</el-checkbox>
          </div>
          <div class="col-product">商品信息</div>
          <div class="col-price">单价</div>
          <div class="col-qty">数量</div>
          <div class="col-subtotal">小计</div>
          <div class="col-action">操作</div>
        </div>

        <!-- 商品列表 -->
        <div class="cart-items">
          <div
            class="cart-item"
            :class="{ checked: item.check }"
            v-for="(item, index) in getShoppingCart"
            :key="item.id"
          >
            <div class="col-check">
              <el-checkbox
                :value="item.check"
                @change="checkChange($event, index)"
              ></el-checkbox>
            </div>
            <div class="col-product">
              <router-link
                :to="{ path: '/goods/details', query: { productID: item.productID } }"
                class="product-info"
              >
                <img :src="$target + item.productImg" class="product-img" />
                <span class="product-name">{{ item.productName }}</span>
              </router-link>
            </div>
            <div class="col-price">
              <span class="price">¥{{ item.price }}</span>
            </div>
            <div class="col-qty">
              <el-input-number
                size="mini"
                :value="item.num"
                @change="handleChange($event, index, item.productID)"
                :min="1"
                :max="item.maxNum"
              ></el-input-number>
            </div>
            <div class="col-subtotal">
              <span class="subtotal">¥{{ (item.price * item.num).toFixed(2) }}</span>
            </div>
            <div class="col-action">
              <el-popconfirm
                title="确定要删除这件商品吗？"
                confirm-button-text="删除"
                cancel-button-text="取消"
                @confirm="deleteItem(item.id, item.productID)"
              >
                <el-button slot="reference" type="text" class="delete-btn">
                  <i class="el-icon-delete"></i> 删除
                </el-button>
              </el-popconfirm>
            </div>
          </div>
        </div>

        <!-- 底部结算栏 -->
        <div class="cart-footer">
          <div class="footer-left">
            <el-checkbox v-model="isAllCheck" class="footer-check">全选</el-checkbox>
            <router-link to="/goods" class="continue-link">
              <i class="el-icon-back"></i> 继续购物
            </router-link>
            <el-button size="small" plain @click="table = true">
              <i class="el-icon-magic-stick"></i> 满减助手
            </el-button>
          </div>
          <div class="footer-right">
            <div class="summary">
              <span class="summary-count">
                已选 <em>{{ getCheckNum }}</em> 件，共 {{ getNum }} 件商品
              </span>
              <span class="summary-total">
                合计：<em class="total-amount">¥{{ getTotalPrice.toFixed(2) }}</em>
              </span>
            </div>
            <router-link :to="getCheckNum > 0 ? '/confirmOrder' : ''">
              <el-button
                type="danger"
                size="medium"
                :disabled="getCheckNum === 0"
                class="checkout-btn"
              >
                去结算 ({{ getCheckNum }})
              </el-button>
            </router-link>
          </div>
        </div>
      </template>

      <!-- 购物车为空 -->
      <div v-else class="cart-empty">
        <div class="empty-content">
          <i class="el-icon-shopping-bag-2 empty-icon"></i>
          <h2>购物车还是空的</h2>
          <p>去挑选心仪的商品吧</p>
          <router-link to="/goods">
            <el-button type="primary" round>去逛逛</el-button>
          </router-link>
        </div>
      </div>
    </div>

    <!-- 满减助手抽屉 -->
    <el-drawer
      title="满减助手"
      :visible.sync="table"
      direction="ltr"
      size="420px"
      :with-header="true"
    >
      <div style="padding: 0 16px">
        <div v-for="(item, index) in getShoppingCart" :key="item.id">
          <FullminusList
            :item="item"
            :index="index"
            @shutDownDrawer="shutDownDrawer"
          ></FullminusList>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import { mapGetters } from 'vuex'
import FullminusList from '../components/FullminusList.vue'

export default {
  data() {
    return { table: false }
  },
  components: { FullminusList },
  methods: {
    ...mapActions(['updateShoppingCart', 'deleteShoppingCart', 'checkAll']),
    handleChange(currentValue, key, productID) {
      this.updateShoppingCart({ key: key, prop: 'check', val: true })
      this.$axios
        .post('/api/user/shoppingCart/updateShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: productID,
          num: currentValue,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.updateShoppingCart({ key: key, prop: 'num', val: currentValue })
            this.notifySucceed(res.data.msg)
          } else {
            this.notifyError(res.data.msg)
          }
        })
        .catch((err) => Promise.reject(err))
    },
    checkChange(val, key) {
      this.updateShoppingCart({ key: key, prop: 'check', val: val })
    },
    deleteItem(id, productID) {
      this.$axios
        .post('/api/user/shoppingCart/deleteShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: productID,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.deleteShoppingCart(id)
            this.notifySucceed(res.data.msg)
          } else {
            this.notifyError(res.data.msg)
          }
        })
        .catch((err) => Promise.reject(err))
    },
    shutDownDrawer() {
      this.table = false
    },
  },
  computed: {
    ...mapGetters(['getShoppingCart', 'getCheckNum', 'getTotalPrice', 'getNum']),
    isAllCheck: {
      get() { return this.$store.getters.getIsAllCheck },
      set(val) { this.checkAll(val) },
    },
  },
}
</script>

<style scoped>
.cart-page {
  background: var(--bg, #f5f5f5);
  min-height: calc(100vh - 260px);
  padding: 24px 0 40px;
}
.cart-wrap {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px;
}

/* 页面标题 */
.page-title {
  display: flex;
  align-items: baseline;
  gap: 16px;
  margin-bottom: 20px;
}
.page-title h1 {
  font-size: 22px;
  font-weight: 600;
  color: var(--text, #333);
}
.page-title h1 i {
  color: var(--primary, #ff6700);
  margin-right: 4px;
}
.promo-tip {
  font-size: 12px;
  color: #f56c6c;
  background: #fef0f0;
  padding: 4px 12px;
  border-radius: 12px;
}
.promo-tip i { margin-right: 2px; }

/* 表头 */
.cart-table-header {
  display: flex;
  align-items: center;
  height: 48px;
  background: #fff;
  border-radius: 8px 8px 0 0;
  padding: 0 20px;
  font-size: 13px;
  color: #999;
  border-bottom: 1px solid #f0f0f0;
}

/* 列宽 */
.col-check { width: 60px; flex-shrink: 0; }
.col-product { flex: 1; min-width: 0; }
.col-price { width: 120px; text-align: center; flex-shrink: 0; }
.col-qty { width: 150px; text-align: center; flex-shrink: 0; }
.col-subtotal { width: 120px; text-align: center; flex-shrink: 0; }
.col-action { width: 100px; text-align: center; flex-shrink: 0; }

/* 商品行 */
.cart-items {
  background: #fff;
  border-radius: 0 0 8px 8px;
}
.cart-item {
  display: flex;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f5f5f5;
  transition: background 0.15s;
}
.cart-item:last-child { border-bottom: none; }
.cart-item:hover { background: #fafafa; }
.cart-item.checked { background: #fff8f0; }

/* 商品信息 */
.product-info {
  display: flex;
  align-items: center;
  gap: 16px;
  color: inherit;
}
.product-img {
  width: 80px;
  height: 80px;
  object-fit: contain;
  border-radius: 6px;
  background: #f9f9f9;
  border: 1px solid #f0f0f0;
  flex-shrink: 0;
}
.product-name {
  font-size: 14px;
  color: var(--text, #333);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.product-info:hover .product-name {
  color: var(--primary, #ff6700);
}

/* 价格 */
.price {
  font-size: 14px;
  color: var(--text-secondary, #666);
}

/* 小计 */
.subtotal {
  font-size: 16px;
  font-weight: 600;
  color: var(--primary, #ff6700);
}

/* 删除 */
.delete-btn {
  color: #999 !important;
  font-size: 13px;
}
.delete-btn:hover {
  color: #f56c6c !important;
}

/* 底部结算栏 */
.cart-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  box-shadow: 0 -2px 8px rgba(0,0,0,0.04);
  position: sticky;
  bottom: 0;
  z-index: 10;
}
.footer-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.footer-check {
  font-size: 13px;
}
.continue-link {
  font-size: 13px;
  color: #999;
  transition: color 0.2s;
}
.continue-link:hover { color: var(--primary, #ff6700); }

.footer-right {
  display: flex;
  align-items: center;
  gap: 24px;
}
.summary {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}
.summary-count {
  font-size: 13px;
  color: #999;
}
.summary-count em {
  font-style: normal;
  color: var(--primary, #ff6700);
  font-weight: 500;
}
.summary-total {
  font-size: 14px;
  color: var(--text, #333);
}
.total-amount {
  font-style: normal;
  font-size: 24px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}
.checkout-btn {
  height: 44px;
  padding: 0 36px;
  font-size: 16px;
  border-radius: 22px;
}

/* 空购物车 */
.cart-empty {
  padding: 80px 0;
}
.empty-content {
  text-align: center;
}
.empty-icon {
  font-size: 80px;
  color: #ddd;
  margin-bottom: 16px;
}
.empty-content h2 {
  font-size: 20px;
  color: #999;
  font-weight: 400;
  margin-bottom: 8px;
}
.empty-content p {
  font-size: 14px;
  color: #bbb;
  margin-bottom: 24px;
}
</style>
