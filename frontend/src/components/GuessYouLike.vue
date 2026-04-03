<!--
 * @Description: 猜你喜欢推荐组件 — 瀑布流无限滚动
 -->
<template>
  <div class="guess-wrap" ref="guessWrap">
    <div class="guess-header">
      <div class="header-left">
        <i class="el-icon-magic-stick"></i>
        <span>猜你喜欢</span>
      </div>
      <span class="header-refresh" @click="refresh">
        <i class="el-icon-refresh" :class="{ spinning: loading }"></i> 换一批
      </span>
    </div>

    <div class="guess-grid">
      <router-link
        v-for="item in recommendations"
        :key="item.product_id"
        :to="{ path: '/goods/details', query: { productID: item.product_id } }"
        class="guess-card"
        @click.native="reportClick(item)"
      >
        <div class="card-img-wrap">
          <img :src="$target + item.product_picture" :alt="item.product_name" loading="lazy" />
          <span class="card-tag" v-if="item.recommend_reason">{{ item.recommend_reason }}</span>
        </div>
        <div class="card-info">
          <p class="card-name">{{ item.product_name }}</p>
          <p class="card-title">{{ item.product_title }}</p>
          <div class="card-bottom">
            <span class="card-price">¥{{ item.product_selling_price }}</span>
            <span class="card-sales" v-if="item.product_sales > 0">{{ item.product_sales }}人已购</span>
          </div>
        </div>
      </router-link>
    </div>

    <div class="guess-loading" v-if="loading">
      <i class="el-icon-loading"></i> 加载中...
    </div>
    <div class="guess-nomore" v-if="!hasMore && recommendations.length > 0">
      — 已经到底了 —
    </div>
    <div class="guess-empty" v-if="!loading && recommendations.length === 0">
      暂无推荐，去逛逛商品吧~
    </div>
  </div>
</template>

<script>
export default {
  name: 'GuessYouLike',
  data() {
    return {
      recommendations: [],
      page: 1,
      hasMore: true,
      loading: false,
    }
  },
  mounted() {
    this.fetchRecommendations()
    window.addEventListener('scroll', this.handleScroll)
  },
  beforeDestroy() {
    window.removeEventListener('scroll', this.handleScroll)
  },
  methods: {
    async fetchRecommendations() {
      if (this.loading || !this.hasMore) return
      this.loading = true
      try {
        const token = localStorage.getItem('token') || ''
        if (!token) {
          // 未登录用户直接请求热门
          this.loading = false
          return
        }
        const res = await this.$axios.post('/api/recommend/guessYouLike', {
          page: this.page,
          page_size: 20,
        })
        if (res.data.code === '200') {
          const items = res.data.recommendations || []
          if (this.page === 1) {
            this.recommendations = items
          } else {
            this.recommendations = this.recommendations.concat(items)
          }
          this.hasMore = res.data.has_more
          this.page++
        }
      } catch (e) {
        // 静默处理
      } finally {
        this.loading = false
      }
    },
    refresh() {
      this.page = 1
      this.hasMore = true
      this.recommendations = []
      this.fetchRecommendations()
    },
    handleScroll() {
      const scrollTop = document.documentElement.scrollTop || document.body.scrollTop
      const clientHeight = document.documentElement.clientHeight
      const scrollHeight = document.documentElement.scrollHeight
      // 距离底部 200px 时触发加载
      if (scrollTop + clientHeight >= scrollHeight - 200) {
        this.fetchRecommendations()
      }
    },
    reportClick(item) {
      // 上报点击行为（异步，不阻塞跳转）
      this.$axios.post('/api/recommend/reportBehavior', {
        product_id: item.product_id,
        category_id: item.category_id,
        behavior_type: 2, // 点击
      }).catch(() => {})
    },
  },
}
</script>

<style scoped>
.guess-wrap {
  margin-top: 32px;
}

.guess-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 20px;
  font-weight: 600;
  color: #333;
}
.header-left i {
  color: #ff6700;
  font-size: 22px;
}
.header-refresh {
  font-size: 13px;
  color: #999;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.15s;
}
.header-refresh:hover {
  color: #ff6700;
}
.spinning {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.guess-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.guess-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s;
  cursor: pointer;
  display: block;
}
.guess-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.card-img-wrap {
  position: relative;
  width: 100%;
  padding-top: 100%;
  background: #f9f9f9;
}
.card-img-wrap img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  padding: 12px;
}
.card-tag {
  position: absolute;
  top: 8px;
  left: 0;
  background: linear-gradient(135deg, #ff6700, #ff9a44);
  color: #fff;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 0 10px 10px 0;
}

.card-info {
  padding: 10px 12px 14px;
}
.card-name {
  font-size: 13px;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-title {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-bottom {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-top: 8px;
}
.card-price {
  font-size: 16px;
  font-weight: 600;
  color: #ff6700;
}
.card-sales {
  font-size: 11px;
  color: #bbb;
}

.guess-loading,
.guess-nomore,
.guess-empty {
  text-align: center;
  padding: 24px 0;
  color: #999;
  font-size: 13px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .guess-grid { grid-template-columns: repeat(4, 1fr); }
}
@media (max-width: 900px) {
  .guess-grid { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 600px) {
  .guess-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
