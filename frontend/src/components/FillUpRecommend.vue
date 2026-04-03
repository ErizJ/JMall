<!--
 * @Description: 智能凑单推荐组件 - 美团凑单风格
 -->
<template>
  <div class="fillup-wrap" v-if="visible">
    <!-- 满减进度条 -->
    <div class="promo-strip">
      <template v-if="gap > 0">
        <div class="strip-left">
          <span class="strip-tag">满减</span>
          <span class="strip-text">
            还差<em>¥{{ gap.toFixed(2) }}</em>享满{{ nearestRule.threshold }}减{{ nearestRule.reduction }}
          </span>
        </div>
        <div class="strip-bar">
          <div class="strip-bar-fill" :style="{ width: progressPercent + '%' }"></div>
        </div>
      </template>
      <template v-else-if="nearestRule.threshold > 0">
        <div class="strip-left">
          <span class="strip-tag done">已满</span>
          <span class="strip-text done-text">已享满{{ nearestRule.threshold }}减{{ nearestRule.reduction }}优惠</span>
        </div>
      </template>
      <span class="strip-action" @click="toggleList" v-if="recommendations.length > 0">
        {{ expanded ? '收起' : '去凑单' }}
        <i :class="expanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"></i>
      </span>
    </div>

    <!-- 凑单商品列表 -->
    <transition name="slide">
      <div class="fillup-list" v-show="expanded && recommendations.length > 0">
        <div class="list-header">
          <span class="list-title">推荐凑单</span>
          <span class="list-refresh" @click="refresh">
            <i class="el-icon-refresh" :class="{ spinning: loading }"></i> 换一批
          </span>
        </div>
        <div class="list-body">
          <div
            class="fillup-item"
            v-for="item in recommendations"
            :key="item.product_id"
          >
            <router-link
              :to="{ path: '/goods/details', query: { productID: item.product_id } }"
              class="item-main"
            >
              <img :src="$target + item.product_picture" class="item-img" />
              <div class="item-info">
                <p class="item-name">{{ item.product_name }}</p>
                <div class="item-meta">
                  <span class="item-reason">{{ item.recommend_reason }}</span>
                </div>
              </div>
            </router-link>
            <div class="item-right">
              <span class="item-price">¥{{ item.product_selling_price }}</span>
              <button
                class="cart-btn"
                :class="{ added: justAdded === item.product_id }"
                :disabled="addingId === item.product_id"
                @click.stop="addToCart(item.product_id)"
              >
                <i :class="justAdded === item.product_id ? 'el-icon-check' : addingId === item.product_id ? 'el-icon-loading' : 'el-icon-plus'"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'FillUpRecommend',
  data() {
    return {
      loading: false,
      addingId: null,
      justAdded: null,
      expanded: true,
      cartTotal: 0,
      gap: 0,
      nearestRule: { threshold: 0, reduction: 0 },
      recommendations: [],
      visible: false,
    }
  },
  computed: {
    ...mapGetters(['getShoppingCart', 'getTotalPrice']),
    progressPercent() {
      if (this.nearestRule.threshold <= 0) return 100
      return Math.min(100, Math.round((this.cartTotal / this.nearestRule.threshold) * 100))
    },
  },
  watch: {
    getTotalPrice() {
      this.fetchRecommendations()
    },
  },
  mounted() {
    this.fetchRecommendations()
  },
  methods: {
    ...mapActions(['unshiftShoppingCart', 'addShoppingCartNum']),

    toggleList() {
      this.expanded = !this.expanded
    },

    async fetchRecommendations() {
      if (!this.$store.getters.getUser) return
      this.loading = true
      try {
        const res = await this.$axios.post('/api/recommend/fillup', {
          user_id: this.$store.getters.getUser.user_id,
        })
        if (res.data.code === '200') {
          this.cartTotal = res.data.cart_total || 0
          this.gap = res.data.gap || 0
          this.nearestRule = res.data.nearest_rule || { threshold: 0, reduction: 0 }
          this.recommendations = res.data.recommendations || []
          this.visible = this.recommendations.length > 0 || this.gap > 0
        }
      } catch (e) {
        console.error('获取凑单推荐失败', e)
      } finally {
        this.loading = false
      }
    },

    refresh() {
      this.fetchRecommendations()
    },

    async addToCart(productId) {
      this.addingId = productId
      try {
        const checkRes = await this.$axios.post('/api/user/shoppingCart/isExistShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: productId,
        })
        if (checkRes.data.code === '002') {
          const addRes = await this.$axios.post('/api/user/shoppingCart/addShoppingCart', {
            user_id: this.$store.getters.getUser.user_id,
            product_id: productId,
            num: 1,
          })
          if (addRes.data.code === '200') {
            this.showAdded(productId)
            this.$emit('cartUpdated')
          }
        } else {
          this.addShoppingCartNum(productId)
          this.showAdded(productId)
          this.$emit('cartUpdated')
        }
        setTimeout(() => {
          this.recommendations = this.recommendations.filter(r => r.product_id !== productId)
          this.fetchRecommendations()
        }, 600)
      } catch (e) {
        console.error('加购失败', e)
      } finally {
        this.addingId = null
      }
    },

    showAdded(productId) {
      this.justAdded = productId
      setTimeout(() => { this.justAdded = null }, 800)
    },
  },
}
</script>

<style scoped>
.fillup-wrap {
  margin-top: 16px;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

/* ===== 顶部满减进度条 ===== */
.promo-strip {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  background: linear-gradient(90deg, #fff5ee, #fff0e5);
  gap: 12px;
}
.strip-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
.strip-tag {
  display: inline-block;
  background: #ff6700;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 3px;
  line-height: 1.5;
}
.strip-tag.done {
  background: #52c41a;
}
.strip-text {
  font-size: 12px;
  color: #666;
  white-space: nowrap;
}
.strip-text em {
  font-style: normal;
  color: #ff6700;
  font-weight: 700;
  margin: 0 1px;
}
.done-text {
  color: #52c41a;
  font-weight: 500;
}
.strip-bar {
  flex: 1;
  height: 4px;
  background: #ffe0c2;
  border-radius: 2px;
  overflow: hidden;
  min-width: 60px;
}
.strip-bar-fill {
  height: 100%;
  background: #ff6700;
  border-radius: 2px;
  transition: width 0.4s ease;
}
.strip-action {
  font-size: 12px;
  color: #ff6700;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  user-select: none;
}
.strip-action i {
  font-size: 10px;
  margin-left: 2px;
}

/* ===== 凑单列表 ===== */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
  max-height: 400px;
  overflow: hidden;
}
.slide-enter,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}

.fillup-list {
  border-top: 1px solid #f5f0eb;
}
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px 6px;
}
.list-title {
  font-size: 13px;
  font-weight: 600;
  color: #333;
}
.list-refresh {
  font-size: 11px;
  color: #999;
  cursor: pointer;
  user-select: none;
}
.list-refresh:hover {
  color: #ff6700;
}
.spinning {
  animation: spin 0.5s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

.list-body {
  max-height: 320px;
  overflow-y: auto;
  padding: 0 16px 8px;
}
.list-body::-webkit-scrollbar {
  width: 3px;
}
.list-body::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 2px;
}

/* ===== 单个商品行 ===== */
.fillup-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f7f7f7;
}
.fillup-item:last-child {
  border-bottom: none;
}
.item-main {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
  color: inherit;
  text-decoration: none;
}
.item-img {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: contain;
  background: #f9f9f9;
  flex-shrink: 0;
}
.item-info {
  flex: 1;
  min-width: 0;
}
.item-name {
  font-size: 13px;
  color: #333;
  margin: 0 0 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.3;
}
.item-meta {
  display: flex;
  align-items: center;
}
.item-reason {
  font-size: 10px;
  color: #ff6700;
  background: #fff5ee;
  padding: 0 5px;
  border-radius: 2px;
  line-height: 1.6;
  border: 1px solid #ffe8d5;
}

.item-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
  margin-left: 8px;
}
.item-price {
  font-size: 14px;
  font-weight: 700;
  color: #ff6700;
  white-space: nowrap;
}

/* ===== 圆形加购按钮 ===== */
.cart-btn {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  background: #ff6700;
  color: #fff;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
  outline: none;
  padding: 0;
  flex-shrink: 0;
}
.cart-btn:hover {
  background: #e55d00;
  transform: scale(1.1);
}
.cart-btn:disabled {
  cursor: default;
  transform: none;
}
.cart-btn.added {
  background: #52c41a;
}
.cart-btn i {
  font-size: 12px;
  line-height: 1;
}
</style>
