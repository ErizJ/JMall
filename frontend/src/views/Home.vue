<!--
 * @Description: 首页
 -->
<template>
  <div class="home">
    <!-- Hero 区域：左侧分类 + 轮播图 -->
    <div class="hero">
      <div class="hero-inner">
        <!-- 左侧分类面板 -->
        <aside class="hero-sidebar">
          <router-link
            v-for="cat in categories"
            :key="cat.id"
            :to="{ path: '/goods', query: { categoryID: [cat.id] } }"
            class="sidebar-item"
          >
            <i :class="cat.icon"></i>
            <span>{{ cat.name }}</span>
            <i class="el-icon-arrow-right arrow"></i>
          </router-link>
        </aside>

        <!-- 轮播图 -->
        <div class="hero-carousel">
          <el-carousel height="400px" autoplay :interval="4000" arrow="hover" @click.native="carouselClick">
            <el-carousel-item v-for="item in carousel" :key="item.carousel_id">
              <div class="slide">
                <img :src="$target + item.imgPath" :alt="item.describes" />
              </div>
            </el-carousel-item>
          </el-carousel>
        </div>

        <!-- 右侧快捷入口 -->
        <aside class="hero-aside">
          <div class="aside-user" v-if="$store.getters.getUser">
            <el-avatar :size="48" src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png"></el-avatar>
            <p class="aside-name">Hi, {{ $store.getters.getUser.userName }}</p>
          </div>
          <div class="aside-user" v-else>
            <el-avatar :size="48" icon="el-icon-user-solid"></el-avatar>
            <p class="aside-name">Hi, 请登录</p>
            <div class="aside-btns">
              <el-button size="mini" type="primary" @click="$store.dispatch('setShowLogin', true)">登录</el-button>
              <el-button size="mini" @click="$parent.register = true">注册</el-button>
            </div>
          </div>
          <div class="aside-shortcuts">
            <router-link to="/order" class="shortcut"><i class="el-icon-document"></i><span>我的订单</span></router-link>
            <router-link to="/collect" class="shortcut"><i class="el-icon-star-off"></i><span>我的收藏</span></router-link>
            <router-link to="/shoppingCart" class="shortcut"><i class="el-icon-shopping-cart-2"></i><span>购物车</span></router-link>
            <router-link to="/goods" class="shortcut"><i class="el-icon-goods"></i><span>全部商品</span></router-link>
          </div>
        </aside>
      </div>
    </div>

    <!-- 主体内容 -->
    <div class="home-content">
      <!-- 限时特惠 -->
      <section class="section">
        <div class="section-head">
          <h2><i class="el-icon-time"></i> 限时特惠</h2>
          <router-link to="/goods" class="more-link">更多 <i class="el-icon-arrow-right"></i></router-link>
        </div>
        <MyList :list="promotionList" :isMore="false"></MyList>
      </section>

      <!-- 为你推荐 -->
      <section class="section">
        <div class="section-head">
          <h2><i class="el-icon-magic-stick"></i> 为你推荐</h2>
          <div class="head-tabs">
            <MyMenu :val="2" @fromChild="getChildMsg">
              <span slot="1">猜你喜欢</span>
              <span slot="2">热门 TOP 7</span>
            </MyMenu>
          </div>
        </div>
        <MyList :list="recommendList" :isMore="true"></MyList>
      </section>

      <!-- 手机三件套 -->
      <section class="section">
        <div class="section-head">
          <h2><i class="el-icon-mobile-phone"></i> 手机三件套</h2>
          <div class="head-tabs">
            <MyMenu :val="3" @fromChild="getChildMsg2">
              <span slot="1">手机</span>
              <span slot="2">保护套</span>
              <span slot="3">充电器</span>
            </MyMenu>
          </div>
        </div>
        <MyList :list="accessoryList" :isMore="true"></MyList>
      </section>

      <!-- 猜你喜欢 -->
      <GuessYouLike v-if="$store.getters.getUser" />
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
      carousel: [],
      recommendList: [],
      promotionList: [],
      accessoryList: [],
      recommendActive: 2,
      accessoryActive: 2,
      categories: [
        { id: 1, name: '手机', icon: 'el-icon-mobile-phone' },
        { id: 2, name: '电视机', icon: 'el-icon-monitor' },
        { id: 3, name: '笔记本', icon: 'el-icon-laptop' },
        { id: 4, name: '平板', icon: 'el-icon-tablet' },
        { id: 5, name: '手机壳', icon: 'el-icon-suitcase' },
        { id: 6, name: '耳机', icon: 'el-icon-headset' },
        { id: 7, name: '充电器', icon: 'el-icon-lightning' },
      ],
    }
  },
  watch: {
    deep: true,
    recommendActive(val) {
      if (val == 1) this.getOneUserRecommendProduct()
      if (val == 2) this.getAllUserRecommendProduct()
    },
    accessoryActive(val) {
      if (val == 1) this.getPhoneList()
      if (val == 2) this.getProtectingShellList()
      if (val == 3) this.getChargerList()
    },
  },
  created() {
    this.$axios.post('/api/resources/carousel', {}).then(r => { this.carousel = r.data.carousel }).catch(() => {})
    this.getPromotionProduct()
    this.getProtectingShellList()
    this.recommendActive = 1
    this.accessoryActive = 1
  },
  methods: {
    ...mapActions(['setUser', 'getUser']),
    getChildMsg(val) { this.recommendActive = val },
    getChildMsg2(val) { this.accessoryActive = val },
    getPhoneList() { this.$axios.post('/api/product/getPhoneList').then(r => { this.accessoryList = r.data.category }).catch(() => {}) },
    getProtectingShellList() { this.$axios.post('/api/product/getProtectingShellList').then(r => { this.accessoryList = r.data.category }).catch(() => {}) },
    getChargerList() { this.$axios.post('/api/product/getChargerList').then(r => { this.accessoryList = r.data.category }).catch(() => {}) },
    getPromotionProduct() { this.$axios.post('/api/product/getPromotionProduct').then(r => { this.promotionList = r.data.category }).catch(() => {}) },
    getOneUserRecommendProduct() { this.$axios.post('/api/product/getOneUserRecommendProduct').then(r => { this.recommendList = r.data.category }).catch(() => {}) },
    getAllUserRecommendProduct() { this.$axios.post('/api/product/getAllUserRecommendProduct').then(r => { this.recommendList = r.data.category }).catch(() => {}) },
    carouselClick() { this.$router.push({ path: '/goods' }) },
  },
}
</script>

<style scoped>
.home { background: var(--bg, #f5f5f5); }

/* ===== Hero 区域 ===== */
.hero {
  background: var(--bg-white, #fff);
  padding-bottom: 16px;
}
.hero-inner {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 12px 20px 0;
  display: flex;
  gap: 12px;
  height: 412px;
}

/* 左侧分类 */
.hero-sidebar {
  width: 190px;
  flex-shrink: 0;
  background: var(--bg, #f7f7f7);
  border-radius: 8px;
  padding: 8px 0;
  overflow-y: auto;
}
.sidebar-item {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  font-size: 13px;
  color: var(--text-secondary, #666);
  transition: all 0.15s;
  gap: 8px;
}
.sidebar-item i:first-child { font-size: 16px; color: var(--text-muted, #999); width: 18px; text-align: center; }
.sidebar-item span { flex: 1; }
.sidebar-item .arrow { font-size: 12px; color: var(--border, #ccc); }
.sidebar-item:hover {
  background: var(--bg-white, #fff);
  color: var(--primary, #ff6700);
  padding-left: 20px;
}
.sidebar-item:hover i { color: var(--primary, #ff6700); }

/* 轮播图 */
.hero-carousel {
  flex: 1;
  border-radius: 8px;
  overflow: hidden;
  min-width: 0;
}
.slide {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
}
.slide img { width: 100%; height: 100%; object-fit: cover; }

/* 右侧快捷入口 */
.hero-aside {
  width: 200px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.aside-user {
  background: var(--bg-white, #fff);
  border: 1px solid var(--border, #f0f0f0);
  border-radius: 8px;
  padding: 20px 16px;
  text-align: center;
}
.aside-name {
  font-size: 14px;
  color: var(--text, #333);
  margin-top: 8px;
}
.aside-btns {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  justify-content: center;
}
.aside-shortcuts {
  background: var(--bg-white, #fff);
  border: 1px solid var(--border, #f0f0f0);
  border-radius: 8px;
  padding: 8px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  flex: 1;
}
.shortcut {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 12px 0;
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-secondary, #666);
  transition: all 0.15s;
}
.shortcut i { font-size: 20px; color: var(--text-muted, #999); }
.shortcut:hover { background: var(--bg, #f9f9f9); color: var(--primary, #ff6700); }
.shortcut:hover i { color: var(--primary, #ff6700); }

/* ===== 主体内容 ===== */
.home-content {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px 40px;
}

/* 区块 */
.section { margin-top: 32px; }
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.section-head h2 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text, #333);
  display: flex;
  align-items: center;
  gap: 6px;
}
.section-head h2 i { color: var(--primary, #ff6700); font-size: 22px; }
.more-link {
  font-size: 13px;
  color: var(--text-muted, #999);
  display: flex;
  align-items: center;
  gap: 2px;
  transition: color 0.15s;
}
.more-link:hover { color: var(--primary, #ff6700); }
.head-tabs { display: flex; align-items: center; }
</style>
