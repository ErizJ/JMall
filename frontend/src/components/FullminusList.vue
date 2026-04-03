<!--
 * @Description: 满减助手列表组件 - 美团凑单风格
 -->
<template>
  <div class="combo-item" v-if="pairProduct && pairProduct.product_id">
    <!-- 当前购物车商品 -->
    <div class="product-row current">
      <img :src="$target + item.productImg" class="product-thumb" />
      <div class="product-detail">
        <p class="product-name">{{ item.productName }}</p>
        <span class="product-price">¥{{ item.price }}</span>
      </div>
      <span class="in-cart-tag">已在购物车</span>
    </div>

    <!-- 加号连接 -->
    <div class="combo-plus">
      <i class="el-icon-plus"></i>
    </div>

    <!-- 推荐搭配商品 -->
    <div class="product-row pair">
      <img :src="$target + pairProduct.product_picture" class="product-thumb" />
      <div class="product-detail">
        <p class="product-name">{{ pairProduct.product_name }}</p>
        <p class="product-desc">{{ pairProduct.product_title }}</p>
        <div class="price-row">
          <span class="product-price">¥{{ pairProduct.product_selling_price }}</span>
          <span
            class="product-price-original"
            v-if="pairProduct.product_price != pairProduct.product_selling_price"
          >¥{{ pairProduct.product_price }}</span>
        </div>
      </div>
    </div>

    <!-- 底部满减信息 + 操作 -->
    <div class="combo-footer">
      <div class="combo-saving">
        <span class="saving-label">组合价</span>
        <span class="saving-price">¥{{ comboPrice }}</span>
        <span class="saving-original">¥{{ originalPrice }}</span>
        <span class="saving-tag">省¥{{ combination.priceReductionRange }}</span>
      </div>
      <button
        class="combo-btn"
        :class="{ added: added }"
        :disabled="adding"
        @click="selectThisCombination"
      >
        <template v-if="added">
          <i class="el-icon-check"></i> 已加入
        </template>
        <template v-else-if="adding">
          <i class="el-icon-loading"></i>
        </template>
        <template v-else>
          加入购物车
        </template>
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'FullminusList',
  props: ['item', 'index'],
  data() {
    return {
      combinationProductList: [],
      combinationProductId: 0,
      combination: {},
      adding: false,
      added: false,
    }
  },
  computed: {
    ...mapGetters(['getShoppingCart']),
    pairProduct() {
      return this.combinationProductList[0] || {}
    },
    comboPrice() {
      if (!this.combination.priceReductionRange) return 0
      return (
        this.item.price +
        (this.pairProduct.product_price || 0) -
        this.combination.priceReductionRange
      )
    },
    originalPrice() {
      return this.item.price + (this.pairProduct.product_price || 0)
    },
  },
  mounted() {
    this.getPairProduct()
  },
  methods: {
    ...mapActions(['unshiftShoppingCart']),

    selectThisCombination() {
      this.adding = true
      const userId = this.$store.getters.getUser.user_id
      const productId = this.pairProduct.product_id

      this.$axios
        .post('/api/user/shoppingCart/isExistShoppingCart', {
          user_id: userId,
          product_id: productId,
        })
        .then((res) => {
          if (res.data.code === '002') {
            // 不在购物车，添加
            this.$axios
              .post('/api/user/shoppingCart/addShoppingCart', {
                user_id: userId,
                product_id: productId,
              })
              .then((res) => {
                this.unshiftShoppingCart(res.data.shoppingCartData[0])
                this.showAdded()
              })
              .finally(() => { this.adding = false })
          } else {
            // 已在购物车
            this.showAdded()
            this.adding = false
          }
        })
        .catch(() => { this.adding = false })
    },

    showAdded() {
      this.added = true
      setTimeout(() => { this.added = false }, 1500)
    },

    getPairProduct() {
      this.$axios
        .post('/api/management/getProductCombination', {
          product_id: this.item.productID,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.combination = (res.data.category || [])[0] || {}
            this.combinationProductId = this.combination.vice_product_id
            if (!this.combinationProductId) return

            this.$axios
              .post('/api/management/getCombinationProductList', {
                product_id: this.combinationProductId,
              })
              .then((res) => {
                if (res.data.code === '001') {
                  this.combinationProductList = res.data.category || []
                }
              })
          }
        })
    },
  },
}
</script>

<style scoped>
.combo-item {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid #f0f0f0;
}

/* ===== 商品行 ===== */
.product-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
}
.product-thumb {
  width: 52px;
  height: 52px;
  border-radius: 6px;
  object-fit: contain;
  background: #f9f9f9;
  flex-shrink: 0;
}
.product-detail {
  flex: 1;
  min-width: 0;
}
.product-name {
  font-size: 13px;
  color: #333;
  margin: 0 0 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}
.product-desc {
  font-size: 11px;
  color: #bbb;
  margin: 0 0 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.price-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.product-price {
  font-size: 14px;
  font-weight: 700;
  color: #ff6700;
}
.product-price-original {
  font-size: 11px;
  color: #ccc;
  text-decoration: line-through;
}
.in-cart-tag {
  font-size: 10px;
  color: #52c41a;
  background: #f0faf0;
  border: 1px solid #d4edda;
  padding: 1px 6px;
  border-radius: 3px;
  flex-shrink: 0;
  white-space: nowrap;
}

/* ===== 加号 ===== */
.combo-plus {
  text-align: center;
  padding: 2px 0;
  color: #ddd;
  font-size: 14px;
}

/* ===== 底部 ===== */
.combo-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  margin-top: 6px;
  border-top: 1px dashed #f0f0f0;
}
.combo-saving {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: wrap;
}
.saving-label {
  font-size: 11px;
  color: #999;
}
.saving-price {
  font-size: 18px;
  font-weight: 700;
  color: #ff6700;
}
.saving-original {
  font-size: 12px;
  color: #ccc;
  text-decoration: line-through;
}
.saving-tag {
  font-size: 10px;
  color: #ff6700;
  background: #fff5ee;
  border: 1px solid #ffe8d5;
  padding: 0 5px;
  border-radius: 2px;
  line-height: 1.7;
}

/* ===== 按钮 ===== */
.combo-btn {
  height: 28px;
  padding: 0 14px;
  border: none;
  border-radius: 14px;
  background: #ff6700;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  outline: none;
  white-space: nowrap;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 3px;
}
.combo-btn:hover {
  background: #e55d00;
}
.combo-btn:disabled {
  cursor: default;
}
.combo-btn.added {
  background: #52c41a;
}
</style>
