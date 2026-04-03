<!--
 * @Description: 我的收藏页面
 -->
<template>
  <div class="collect-page">
    <div class="collect-wrap">
      <!-- 页面标题 -->
      <div class="page-header">
        <div class="header-left">
          <h1><i class="el-icon-star-on"></i> 我的收藏</h1>
          <span class="count" v-if="collectList.length > 0">{{ collectList.length }} 件商品</span>
        </div>
        <router-link to="/goods" class="go-shop" v-if="collectList.length > 0">
          <i class="el-icon-plus"></i> 继续逛逛
        </router-link>
      </div>

      <!-- 有收藏 -->
      <div class="collect-grid" v-if="collectList.length > 0">
        <div class="collect-card" v-for="item in collectList" :key="item.product_id">
          <!-- 删除按钮 -->
          <el-popconfirm
            title="确定取消收藏吗？"
            confirm-button-text="取消收藏"
            cancel-button-text="再想想"
            @confirm="removeCollect(item.product_id)"
          >
            <span slot="reference" class="remove-btn">
              <i class="el-icon-close"></i>
            </span>
          </el-popconfirm>

          <router-link
            :to="{ path: '/goods/details', query: { productID: item.product_id } }"
            class="card-link"
          >
            <!-- 图片 -->
            <div class="card-img">
              <img :src="$target + item.product_picture" alt="" />
            </div>

            <!-- 信息 -->
            <div class="card-body">
              <h3 class="card-name">{{ item.product_name }}</h3>
              <p class="card-desc">{{ item.product_title }}</p>
              <div class="card-bottom">
                <div class="card-price">
                  <span class="now">¥{{ item.product_selling_price }}</span>
                  <span class="old" v-if="item.product_price !== item.product_selling_price">
                    ¥{{ item.product_price }}
                  </span>
                </div>
                <el-button type="primary" size="mini" round class="btn-cart" @click.prevent="addToCart(item)">
                  <i class="el-icon-shopping-cart-2"></i> 加入购物车
                </el-button>
              </div>
            </div>
          </router-link>
        </div>
      </div>

      <!-- 空收藏 -->
      <div v-else class="collect-empty">
        <i class="el-icon-star-off empty-icon"></i>
        <h2>收藏夹还是空的</h2>
        <p>去发现喜欢的商品，点击 ❤ 收藏吧</p>
        <router-link to="/goods">
          <el-button type="primary" round>去逛逛</el-button>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
import { mapActions } from 'vuex'

export default {
  data() {
    return { collectList: [] }
  },
  activated() {
    this.$axios
      .post('/api/user/collect/getCollect', {
        user_id: this.$store.getters.getUser.user_id,
      })
      .then((res) => {
        if (res.data.code === '001') {
          this.collectList = res.data.collectList || []
        }
      })
      .catch(() => {})
  },
  methods: {
    ...mapActions(['unshiftShoppingCart']),
    removeCollect(productId) {
      this.$axios
        .post('/api/user/collect/deleteCollect', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: productId,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.collectList = this.collectList.filter(c => c.product_id !== productId)
            this.notifySucceed('已取消收藏')
          }
        })
    },
    addToCart(item) {
      this.$axios
        .post('/api/user/shoppingCart/addShoppingCart', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: item.product_id,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.unshiftShoppingCart(res.data.shoppingCartData[0])
            this.notifySucceed('已加入购物车')
          } else if (res.data.code === '002') {
            this.notifySucceed('该商品已在购物车')
          } else {
            this.notifyError(res.data.msg)
          }
        })
    },
  },
}
</script>

<style scoped>
.collect-page {
  background: var(--bg, #f5f5f5);
  min-height: calc(100vh - 260px);
  padding: 24px 0 40px;
}
.collect-wrap {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px;
}

/* 页面头 */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.header-left h1 {
  font-size: 22px;
  font-weight: 600;
  color: #333;
}
.header-left h1 i {
  color: var(--primary, #ff6700);
}
.count {
  font-size: 13px;
  color: #999;
  background: #f5f5f5;
  padding: 2px 10px;
  border-radius: 10px;
}
.go-shop {
  font-size: 13px;
  color: #999;
  transition: color 0.2s;
}
.go-shop:hover {
  color: var(--primary, #ff6700);
}

/* 收藏网格 */
.collect-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(228px, 1fr));
  gap: 16px;
}

.collect-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  border: 1px solid transparent;
  transition: all 0.25s;
}
.collect-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  transform: translateY(-3px);
  border-color: #eee;
}

/* 删除按钮 */
.remove-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 3;
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.35);
  color: #fff;
  font-size: 12px;
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
}
.collect-card:hover .remove-btn {
  opacity: 1;
}
.remove-btn:hover {
  background: #f56c6c;
}

.card-link {
  display: block;
  color: inherit;
}

/* 图片 */
.card-img {
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  padding: 16px;
  overflow: hidden;
}
.card-img img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transition: transform 0.3s;
}
.collect-card:hover .card-img img {
  transform: scale(1.05);
}

/* 信息 */
.card-body {
  padding: 12px 14px 14px;
}
.card-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}
.card-desc {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 10px;
}
.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-price {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.now {
  font-size: 18px;
  font-weight: 600;
  color: var(--primary, #ff6700);
}
.old {
  font-size: 12px;
  color: #bbb;
  text-decoration: line-through;
}
.btn-cart {
  font-size: 12px;
  padding: 5px 10px;
}

/* 空状态 */
.collect-empty {
  text-align: center;
  padding: 80px 0;
}
.empty-icon {
  font-size: 72px;
  color: #ddd;
}
.collect-empty h2 {
  font-size: 20px;
  color: #999;
  font-weight: 400;
  margin: 14px 0 6px;
}
.collect-empty p {
  font-size: 14px;
  color: #bbb;
  margin-bottom: 24px;
}
</style>
