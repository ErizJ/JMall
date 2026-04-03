<!--
 * @Description: 商品详情页面组件 - 现代电商风格
 -->
<template>
  <div id="details">
    <!-- 面包屑导航 -->
    <div class="breadcrumb-bar">
      <div class="breadcrumb-inner">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item :to="{ path: '/goods' }">全部商品</el-breadcrumb-item>
          <el-breadcrumb-item>商品详情</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
    </div>

    <!-- 主要内容 -->
    <div class="detail-container">
      <div class="detail-card">
        <!-- 左侧商品图片 -->
        <div class="gallery">
          <div class="gallery-main">
            <el-carousel height="480px" v-if="productPicture.length > 1" indicator-position="none" arrow="hover" @change="onCarouselChange">
              <el-carousel-item v-for="item in productPicture" :key="item.id">
                <img :src="$target + item.product_picture" :alt="item.intro" />
              </el-carousel-item>
            </el-carousel>
            <div v-else-if="productPicture.length === 1" class="single-img">
              <img :src="$target + productPicture[0].product_picture" :alt="productPicture[0].intro" />
            </div>
          </div>
          <!-- 缩略图列表 -->
          <div class="gallery-thumbs" v-if="productPicture.length > 1">
            <div
              v-for="(item, idx) in productPicture"
              :key="'thumb-' + item.id"
              :class="['thumb-item', { active: currentSlide === idx }]"
            >
              <img :src="$target + item.product_picture" :alt="item.intro" />
            </div>
          </div>
        </div>

        <!-- 右侧信息区 -->
        <div class="info-panel">
          <h1 class="product-title">{{ productDetails.product_name }}</h1>
          <p class="product-desc">{{ productDetails.product_intro }}</p>

          <!-- 价格区域 -->
          <div class="price-block">
            <div class="price-row">
              <span class="label">售价</span>
              <span class="current-price">¥{{ productDetails.product_selling_price }}</span>
              <span
                v-if="productDetails.product_price != productDetails.product_selling_price"
                class="original-price"
              >¥{{ productDetails.product_price }}</span>
              <span
                v-if="productDetails.product_price != productDetails.product_selling_price && productDetails.product_price > 0"
                class="discount-tag"
              >省¥{{ (productDetails.product_price - productDetails.product_selling_price).toFixed(2) }}</span>
            </div>
            <div class="price-row" v-if="productDetails.product_isPromotion">
              <span class="label">促销</span>
              <el-tag size="small" type="danger" effect="dark">限时促销</el-tag>
            </div>
          </div>

          <!-- 商品信息 -->
          <div class="meta-block">
            <div class="meta-row">
              <span class="label">配送</span>
              <span class="value"><i class="el-icon-truck"></i> 免运费 · 预计1-3天送达</span>
            </div>
            <div class="meta-row">
              <span class="label">服务</span>
              <span class="value service-tags">
                <span><i class="el-icon-circle-check"></i> JMall自营</span>
                <span><i class="el-icon-circle-check"></i> 7天无理由退货</span>
                <span><i class="el-icon-circle-check"></i> 正品保证</span>
              </span>
            </div>
          </div>

          <!-- 购买数量 -->
          <div class="qty-block">
            <span class="label">数量</span>
            <el-input-number v-model="buyNum" :min="1" :max="productDetails.product_num || 99" size="medium"></el-input-number>
            <span class="stock" v-if="productDetails.product_num">库存 {{ productDetails.product_num }} 件</span>
          </div>

          <!-- 操作按钮 -->
          <div class="action-block">
            <el-button class="btn-buy" @click="buyNow" :disabled="dis">立即购买</el-button>
            <el-button class="btn-cart" :disabled="dis" @click="addShoppingCart">
              <i class="el-icon-shopping-cart-2"></i> 加入购物车
            </el-button>
            <el-button :class="['btn-fav', { collected: isCollected }]" @click="toggleCollect">
              <i :class="isCollected ? 'el-icon-star-on' : 'el-icon-star-off'"></i>
              {{ isCollected ? '已收藏' : '收藏' }}
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 猜你喜欢 -->
    <div class="detail-container" style="margin-top: 20px;">
      <GuessYouLike />
    </div>
  </div>
</template>
<script>
import { mapActions } from 'vuex'
import GuessYouLike from '../components/GuessYouLike.vue'
export default {
  components: { GuessYouLike },
  data() {
    return {
      dis: false,
      buyNum: 1,
      currentSlide: 0,
      isCollected: false,
      productID: '',
      productDetails: '',
      productPicture: '',
    }
  },
  activated() {
    if (this.$route.query.productID != undefined) {
      this.productID = this.$route.query.productID
    }
  },
  watch: {
    productID: function (val) {
      this.getDetails(val)
      this.getDetailsPicture(val)
      this.checkCollectStatus(val)
      this.buyNum = 1
      this.dis = false
      this.isCollected = false
    },
  },
  methods: {
    ...mapActions(['unshiftShoppingCart', 'addShoppingCartNum']),
    onCarouselChange(idx) {
      this.currentSlide = idx
    },
    getDetails(val) {
      this.$axios
        .post('/api/product/getDetails', { productID: val })
        .then((res) => {
          this.productDetails = res.data.Product[0]
          this.reportBehavior(this.productDetails, 1)
        })
        .catch((err) => Promise.reject(err))
    },
    getDetailsPicture(val) {
      this.$axios
        .post('/api/product/getDetailsPicture', { productID: val })
        .then((res) => {
          this.productPicture = res.data.ProductPicture
        })
        .catch((err) => Promise.reject(err))
    },
    // 立即购买
    buyNow() {
      if (!this.$store.getters.getUser) {
        this.$store.dispatch('setShowLogin', true)
        return
      }
      // 先加入购物车，再跳转确认订单
      this.$axios
        .post('/api/user/shoppingCart/addShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: this.productID,
        })
        .then((res) => {
          if (res.data.code === '001' || res.data.code === '002') {
            if (res.data.code === '001') {
              this.unshiftShoppingCart(res.data.shoppingCartData[0])
            }
            // 自动勾选该商品并跳转确认订单
            this.$store.dispatch('checkAll', false)
            const cart = this.$store.getters.getShoppingCart
            for (let i = 0; i < cart.length; i++) {
              if (String(cart[i].productID) === String(this.productID)) {
                this.$store.dispatch('updateShoppingCart', { key: i, prop: 'check', val: true })
                break
              }
            }
            this.$router.push('/confirmOrder')
          } else if (res.data.code === '003') {
            this.dis = true
            this.notifyError(res.data.msg)
          } else {
            this.notifyError(res.data.msg)
          }
        })
        .catch((err) => Promise.reject(err))
    },
    addShoppingCart() {
      if (!this.$store.getters.getUser) {
        this.$store.dispatch('setShowLogin', true)
        return
      }
      this.$axios
        .post('/api/user/shoppingCart/addShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: this.productID,
        })
        .then((res) => {
          switch (res.data.code) {
            case '001':
              this.unshiftShoppingCart(res.data.shoppingCartData[0])
              this.notifySucceed(res.data.msg)
              break
            case '002':
              this.addShoppingCartNum(this.productID)
              this.notifySucceed(res.data.msg)
              break
            case '003':
              this.dis = true
              this.notifyError(res.data.msg)
              break
            default:
              this.notifyError(res.data.msg)
          }
        })
        .catch((err) => Promise.reject(err))
    },
    // 检查收藏状态
    checkCollectStatus(productId) {
      if (!this.$store.getters.getUser) return
      this.$axios
        .post('/api/user/collect/isCollected', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: productId,
        })
        .then((res) => {
          if (res.data.code === '200') {
            this.isCollected = res.data.is_collected
          }
        })
        .catch(() => {})
    },
    // 收藏/取消收藏
    toggleCollect() {
      if (!this.$store.getters.getUser) {
        this.$store.dispatch('setShowLogin', true)
        return
      }
      if (this.isCollected) {
        // 取消收藏
        this.$axios
          .post('/api/user/collect/deleteCollect', {
            user_id: this.$store.getters.getUser.user_id,
            product_id: this.productID,
          })
          .then((res) => {
            if (res.data.code == '001' || res.data.code === '200') {
              this.isCollected = false
              this.notifySucceed('已取消收藏')
            } else {
              this.notifyError(res.data.msg || '操作失败')
            }
          })
          .catch((err) => Promise.reject(err))
      } else {
        // 添加收藏
        this.$axios
          .post('/api/user/collect/addCollect', {
            user_id: this.$store.getters.getUser.user_id,
            product_id: this.productID,
          })
          .then((res) => {
            if (res.data.code == '001' || res.data.code === '200') {
              this.isCollected = true
              this.notifySucceed('收藏成功')
            } else {
              this.notifyError(res.data.msg || '收藏失败')
            }
          })
          .catch((err) => Promise.reject(err))
      }
    },
    reportBehavior(product, behaviorType) {
      if (!this.$store.getters.getUser || !product) return
      this.$axios.post('/api/recommend/reportBehavior', {
        product_id: Number(product.product_id),
        category_id: Number(product.category_id),
        behavior_type: behaviorType,
      }).catch(() => {})
    },
  },
}
</script>
<style scoped>
#details {
  background: var(--bg, #f5f5f5);
  padding-bottom: 40px;
}

/* 面包屑 */
.breadcrumb-bar {
  background: var(--bg-white, #fff);
  border-bottom: 1px solid var(--border, #e8e8e8);
  padding: 14px 0;
}
.breadcrumb-inner {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px;
}

/* 容器 */
.detail-container {
  max-width: var(--content-width, 1226px);
  margin: 20px auto 0;
  padding: 0 20px;
}
.detail-card {
  background: var(--bg-white, #fff);
  border-radius: 8px;
  padding: 30px;
  display: flex;
  gap: 40px;
}

/* 图片画廊 */
.gallery {
  width: 480px;
  flex-shrink: 0;
}
.gallery-main {
  width: 480px;
  height: 480px;
  border-radius: 8px;
  overflow: hidden;
  background: #f9f9f9;
  border: 1px solid var(--border, #f0f0f0);
}
.gallery-main img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.single-img {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.gallery-thumbs {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.thumb-item {
  width: 64px;
  height: 64px;
  border-radius: 6px;
  overflow: hidden;
  border: 2px solid transparent;
  cursor: pointer;
  transition: border-color 0.2s;
}
.thumb-item img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.thumb-item.active,
.thumb-item:hover {
  border-color: var(--primary, #ff6700);
}

/* 信息面板 */
.info-panel {
  flex: 1;
  min-width: 0;
}
.product-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text, #333);
  line-height: 1.4;
  margin-bottom: 8px;
}
.product-desc {
  font-size: 14px;
  color: var(--text-muted, #999);
  margin-bottom: 20px;
  line-height: 1.6;
}

/* 价格区域 */
.price-block {
  background: linear-gradient(135deg, #fff8f0 0%, #fff3e6 100%);
  border-radius: 8px;
  padding: 20px 24px;
  margin-bottom: 20px;
}
.price-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.price-row + .price-row {
  margin-top: 10px;
}
.price-block .label {
  font-size: 13px;
  color: #999;
  width: 36px;
  flex-shrink: 0;
}
.current-price {
  font-size: 32px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}
.original-price {
  font-size: 14px;
  color: #bbb;
  text-decoration: line-through;
}
.discount-tag {
  font-size: 12px;
  color: #fff;
  background: var(--primary, #ff6700);
  padding: 2px 8px;
  border-radius: 3px;
}

/* 商品信息 */
.meta-block {
  padding: 16px 0;
  border-bottom: 1px solid var(--border, #f0f0f0);
}
.meta-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px 0;
  font-size: 14px;
}
.meta-block .label {
  font-size: 13px;
  color: #999;
  width: 36px;
  flex-shrink: 0;
  line-height: 22px;
}
.meta-row .value {
  color: var(--text-secondary, #666);
  line-height: 22px;
}
.meta-row .value i {
  color: var(--primary, #ff6700);
  margin-right: 2px;
}
.service-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
.service-tags span {
  color: var(--text-muted, #999);
  font-size: 13px;
}
.service-tags span i {
  color: #67c23a;
}

/* 数量 */
.qty-block {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 0;
}
.qty-block .label {
  font-size: 13px;
  color: #999;
  width: 36px;
  flex-shrink: 0;
}
.stock {
  font-size: 13px;
  color: #bbb;
}

/* 操作按钮 */
.action-block {
  display: flex;
  gap: 12px;
  padding-top: 8px;
}
.btn-buy {
  height: 48px;
  padding: 0 48px;
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  background: var(--primary, #ff6700);
  border: none;
  border-radius: 6px;
}
.btn-buy:hover {
  background: var(--primary-dark, #e55d00);
}
.btn-buy:disabled {
  background: #ddd;
  color: #999;
}
.btn-cart {
  height: 48px;
  padding: 0 36px;
  font-size: 16px;
  color: var(--primary, #ff6700);
  border: 2px solid var(--primary, #ff6700);
  background: transparent;
  border-radius: 6px;
}
.btn-cart:hover {
  background: rgba(255, 103, 0, 0.06);
}
.btn-cart:disabled {
  border-color: #ddd;
  color: #999;
}
.btn-fav {
  height: 48px;
  padding: 0 24px;
  font-size: 14px;
  color: var(--text-secondary, #666);
  border: 1px solid var(--border, #ddd);
  background: transparent;
  border-radius: 6px;
  transition: all 0.25s;
}
.btn-fav:hover {
  color: var(--primary, #ff6700);
  border-color: var(--primary, #ff6700);
}
.btn-fav.collected {
  color: var(--primary, #ff6700);
  border-color: var(--primary, #ff6700);
  background: rgba(255, 103, 0, 0.06);
}
.btn-fav.collected:hover {
  color: #999;
  border-color: #ddd;
  background: transparent;
}

/* Element UI carousel 覆盖 */
#details .el-carousel .el-carousel__indicator .el-carousel__button {
  background-color: rgba(163, 163, 163, 0.8);
}
</style>
