<template>
  <div id="app">
    <!-- ===== 顶部通栏（登录/注册） ===== -->
    <div class="top-bar">
      <div class="top-inner">
        <span class="welcome">欢迎来到 JMall</span>
        <div class="top-right">
          <template v-if="!$store.getters.getUser">
            <a href="javascript:;" class="top-link" @click="login">登录</a>
            <span class="top-sep">|</span>
            <a href="javascript:;" class="top-link" @click="register = true">免费注册</a>
            <span class="top-sep">|</span>
            <router-link to="/order" class="top-link">我的订单</router-link>
            <span class="top-sep">|</span>
            <router-link to="/collect" class="top-link">收藏夹</router-link>
          </template>
          <template v-else>
            <el-dropdown trigger="click" @command="handleCommand" size="small">
              <span class="top-user">
                <el-avatar :size="20" src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png"></el-avatar>
                {{ $store.getters.getUser.userName }}
                <i class="el-icon-arrow-down"></i>
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="order"><i class="el-icon-document"></i> 我的订单</el-dropdown-item>
                <el-dropdown-item command="collect"><i class="el-icon-star-off"></i> 我的收藏</el-dropdown-item>
                <el-dropdown-item command="cart"><i class="el-icon-shopping-cart-2"></i> 购物车</el-dropdown-item>
                <el-dropdown-item divided command="logout"><i class="el-icon-switch-button"></i> 退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
          <template v-if="!disable">
            <span class="top-sep">|</span>
            <router-link to="/manager" class="top-link admin-link"><i class="el-icon-setting"></i> 管理后台</router-link>
          </template>
          <span class="top-sep">|</span>
          <a href="javascript:;" class="top-link theme-toggle" @click="toggleTheme">
            <i :class="isDark ? 'el-icon-sunny' : 'el-icon-moon'"></i>
            {{ isDark ? '浅色' : '深色' }}
          </a>
        </div>
      </div>
    </div>

    <!-- ===== 主导航栏（Logo + 搜索 + 购物车） ===== -->
    <header class="main-header" :class="{ 'is-fixed': headerFixed }">
      <div class="header-inner">
        <router-link to="/" class="logo">
          <span class="logo-text">JMall</span>
          <span class="logo-sub">商城</span>
        </router-link>

        <!-- 分类导航 -->
        <nav class="cat-nav">
          <router-link to="/home" class="cat-item" :class="{ active: isHome }">首页</router-link>
          <router-link to="/goods" class="cat-item" :class="{ active: $route.path === '/goods' }">全部商品</router-link>
        </nav>

        <!-- 搜索框 -->
        <div class="search-wrap">
          <el-input
            v-model="search"
            placeholder="搜索你想要的商品"
            size="medium"
            @keyup.enter.native="searchClick"
            clearable
          >
            <el-button slot="append" @click="searchClick">搜索</el-button>
          </el-input>
        </div>

        <!-- 购物车入口 -->
        <router-link to="/shoppingCart" class="cart-entry" :class="{ active: $route.path === '/shoppingCart' }">
          <i class="el-icon-shopping-cart-2"></i>
          <span>购物车</span>
          <em v-if="getNum > 0" class="cart-num">{{ getNum }}</em>
        </router-link>
      </div>
    </header>

    <!-- 占位 -->
    <div class="header-space"></div>

    <!-- 登录 / 注册 -->
    <MyLogin></MyLogin>
    <MyRegister :register="register" @fromChild="isRegister"></MyRegister>

    <!-- 主内容 -->
    <main class="main-body">
      <keep-alive>
        <router-view></router-view>
      </keep-alive>
    </main>

    <!-- AI 智能助手 -->
    <AiChat />

    <!-- 底栏 -->
    <footer class="site-footer">
      <div class="footer-services">
        <div class="service-item"><i class="el-icon-refresh-left"></i><div><p>7天无理由</p><span>退换无忧</span></div></div>
        <div class="service-item"><i class="el-icon-truck"></i><div><p>满99免邮</p><span>极速配送</span></div></div>
        <div class="service-item"><i class="el-icon-circle-check"></i><div><p>品质保证</p><span>正品行货</span></div></div>
        <div class="service-item"><i class="el-icon-service"></i><div><p>在线客服</p><span>贴心服务</span></div></div>
      </div>
      <div class="footer-bottom">
        <div class="footer-links">
          <a href="javascript:;">关于JMall</a><span>|</span>
          <a href="javascript:;">隐私政策</a><span>|</span>
          <a href="javascript:;">使用条款</a><span>|</span>
          <a href="javascript:;">联系我们</a>
        </div>
        <p>Copyright © 2025 JMall Inc. 保留所有权利 · Power By ErizJ</p>
      </div>
    </footer>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import { mapGetters } from 'vuex'
import AiChat from './components/AiChat.vue'

export default {
  components: { AiChat },
  data() {
    return {
      search: '',
      register: false,
      flag: true,
      headerFixed: false,
      isDark: false,
    }
  },
  mounted() {
    if (localStorage.getItem('user')) {
      try { this.setUser(JSON.parse(localStorage.getItem('user'))) } catch (e) { /* */ }
    }
    // 恢复主题
    const saved = localStorage.getItem('jmall-theme')
    if (saved === 'dark') {
      this.isDark = true
      document.documentElement.setAttribute('data-theme', 'dark')
    }
    window.addEventListener('scroll', this.onScroll)
  },
  beforeDestroy() {
    window.removeEventListener('scroll', this.onScroll)
  },
  computed: {
    ...mapGetters(['getUser', 'getNum']),
    disable() { return this.flag },
    isHome() { return this.$route.path === '/' || this.$route.path === '/home' },
  },
  watch: {
    getUser(val) {
      if (val === '') {
        this.setShoppingCart([])
      } else {
        this.checkUserIsManager(val.userName)
        this.$axios.post('/api/user/shoppingCart/getShoppingCart', { user_id: val.user_id })
          .then((res) => { if (res.data.code === '001') this.setShoppingCart(res.data.shoppingCartData) })
          .catch(() => {})
      }
    },
  },
  methods: {
    ...mapActions(['setUser', 'setShowLogin', 'setShoppingCart']),
    onScroll() { this.headerFixed = window.scrollY > 32 },
    toggleTheme() {
      this.isDark = !this.isDark
      if (this.isDark) {
        document.documentElement.setAttribute('data-theme', 'dark')
        localStorage.setItem('jmall-theme', 'dark')
      } else {
        document.documentElement.removeAttribute('data-theme')
        localStorage.setItem('jmall-theme', 'light')
      }
    },
    login() { this.setShowLogin(true) },
    logout() {
      localStorage.setItem('user', '')
      this.setUser('')
      this.$axios.post('/api/product/setCategoryHotZero').catch(() => {})
      this.notifySucceed('已退出登录')
    },
    isRegister(val) { this.register = val },
    searchClick() {
      if (this.search.trim()) {
        this.$router.push({ path: '/goods', query: { search: this.search } })
        this.search = ''
      }
    },
    handleCommand(cmd) {
      if (cmd === 'logout') this.logout()
      else if (cmd === 'order') this.$router.push('/order')
      else if (cmd === 'collect') this.$router.push('/collect')
      else if (cmd === 'cart') this.$router.push('/shoppingCart')
    },
    checkUserIsManager(name) {
      this.$axios.post('/api/users/isManager', { user_name: name })
        .then((res) => { this.flag = res.data.code !== '001' })
        .catch(() => { this.flag = true })
    },
  },
}
</script>

<style lang="css">
@import './reset.css';

:root {
  --primary: #ff6700;
  --primary-light: #ff8533;
  --primary-dark: #e55d00;
  --text: #333;
  --text-secondary: #666;
  --text-muted: #999;
  --border: #e8e8e8;
  --bg: #f5f5f5;
  --bg-white: #fff;
  --top-bar-height: 32px;
  --header-height: 64px;
  --content-width: 1226px;
  --shadow-sm: 0 2px 8px rgba(0,0,0,0.06);
  --radius: 8px;
}

html {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
  background: var(--bg);
  color: var(--text);
  -webkit-font-smoothing: antialiased;
  transition: background-color 0.3s, color 0.3s;
}

/* ===== 深色模式 ===== */
html[data-theme="dark"] {
  --primary: #ff8533;
  --primary-light: #ffa366;
  --primary-dark: #e56b00;
  --text: #e0e0e0;
  --text-secondary: #aaa;
  --text-muted: #777;
  --border: #333;
  --bg: #1a1a1a;
  --bg-white: #242424;
  --shadow-sm: 0 2px 8px rgba(0,0,0,0.3);
}

/* 深色模式下 Element UI 组件覆盖 */
html[data-theme="dark"] .el-input__inner {
  background: #2c2c2c;
  border-color: #444;
  color: #e0e0e0;
}
html[data-theme="dark"] .el-input-group__append {
  background: var(--primary);
  border-color: var(--primary);
}
html[data-theme="dark"] .el-dialog {
  background: #2c2c2c;
}
html[data-theme="dark"] .el-dialog__title {
  color: #e0e0e0;
}
html[data-theme="dark"] .el-dialog__body {
  color: #ccc;
}
html[data-theme="dark"] .el-form-item__label {
  color: #aaa;
}
html[data-theme="dark"] .el-table {
  background: #242424;
  color: #ccc;
}
html[data-theme="dark"] .el-table th {
  background: #2c2c2c !important;
  color: #aaa !important;
}
html[data-theme="dark"] .el-table tr {
  background: #242424;
}
html[data-theme="dark"] .el-table--enable-row-hover .el-table__body tr:hover > td {
  background: #2f2f2f !important;
}
html[data-theme="dark"] .el-table td, html[data-theme="dark"] .el-table th {
  border-color: #333;
}
html[data-theme="dark"] .el-table--border::after,
html[data-theme="dark"] .el-table--group::after,
html[data-theme="dark"] .el-table::before {
  background: #333;
}
html[data-theme="dark"] .el-card {
  background: #242424;
  border-color: #333;
}
html[data-theme="dark"] .el-card__header {
  border-color: #333;
}
html[data-theme="dark"] .el-menu {
  background: transparent;
  border-color: #333;
}
html[data-theme="dark"] .el-menu-item {
  color: #aaa;
}
html[data-theme="dark"] .el-breadcrumb__inner {
  color: #888 !important;
}
html[data-theme="dark"] .el-pagination.is-background .el-pager li {
  background: #2c2c2c;
  color: #aaa;
}
html[data-theme="dark"] .el-pagination.is-background .el-pager li.active {
  background: var(--primary);
}
html[data-theme="dark"] .el-pagination.is-background .btn-prev,
html[data-theme="dark"] .el-pagination.is-background .btn-next {
  background: #2c2c2c;
  color: #aaa;
}
html[data-theme="dark"] .el-checkbox__label {
  color: #ccc;
}
html[data-theme="dark"] .el-input-number .el-input__inner {
  background: #2c2c2c;
}
html[data-theme="dark"] .el-tag {
  border-color: transparent;
}
html[data-theme="dark"] .el-select-dropdown {
  background: #2c2c2c;
  border-color: #444;
}
html[data-theme="dark"] .el-select-dropdown__item {
  color: #ccc;
}
html[data-theme="dark"] .el-dropdown-menu {
  background: #2c2c2c;
  border-color: #444;
}
html[data-theme="dark"] .el-dropdown-menu__item {
  color: #ccc;
}
html[data-theme="dark"] .el-dropdown-menu__item:hover {
  background: #363636;
  color: var(--primary);
}
html[data-theme="dark"] .el-drawer {
  background: #242424;
}
html[data-theme="dark"] .el-drawer__header {
  color: #e0e0e0;
}
html[data-theme="dark"] .el-popover {
  background: #2c2c2c;
  border-color: #444;
  color: #ccc;
}
html[data-theme="dark"] .el-divider {
  border-color: #333;
}
html[data-theme="dark"] .el-tabs__item {
  color: #aaa;
}
html[data-theme="dark"] .el-tabs__item.is-active {
  color: var(--primary);
}

/* 深色模式 — 全局通用类覆盖 */
html[data-theme="dark"] .manage-card,
html[data-theme="dark"] .order-block,
html[data-theme="dark"] .order-card,
html[data-theme="dark"] .combo-card,
html[data-theme="dark"] .collect-card,
html[data-theme="dark"] .stat-card,
html[data-theme="dark"] .form-card,
html[data-theme="dark"] .upload-card {
  background: #242424;
  border-color: #333;
}
html[data-theme="dark"] .order-head,
html[data-theme="dark"] .order-foot,
html[data-theme="dark"] .card-head,
html[data-theme="dark"] .card-foot,
html[data-theme="dark"] .cart-table-header {
  background: #2a2a2a;
  border-color: #333;
}
html[data-theme="dark"] .product-row,
html[data-theme="dark"] .item-row,
html[data-theme="dark"] .cart-item {
  border-color: #333;
}
html[data-theme="dark"] .cart-item.checked {
  background: #2d2520;
}
html[data-theme="dark"] .cart-footer {
  background: #242424;
  box-shadow: 0 -2px 8px rgba(0,0,0,0.3);
}
html[data-theme="dark"] .product-img,
html[data-theme="dark"] .prod-img img,
html[data-theme="dark"] .item-img,
html[data-theme="dark"] .thumb,
html[data-theme="dark"] .product-thumb,
html[data-theme="dark"] .detail-thumb {
  background: #2c2c2c;
  border-color: #3a3a3a;
}
html[data-theme="dark"] .img-placeholder {
  background: #2c2c2c;
  color: #555;
}
html[data-theme="dark"] .combo-item {
  background: #2c2c2c;
}
html[data-theme="dark"] .combo-badge {
  background: var(--primary);
}
html[data-theme="dark"] .threshold {
  background: #3d3020;
  color: #fa8c16;
}
html[data-theme="dark"] .reduction {
  background: #3d2020;
  color: #f5222d;
}
html[data-theme="dark"] .hero-sidebar {
  background: #2a2a2a;
}
html[data-theme="dark"] .sidebar-item:hover {
  background: #333;
}
html[data-theme="dark"] .filter-bar {
  background: #242424;
  border-color: #333;
}
html[data-theme="dark"] .cat-tag {
  color: #aaa;
}
html[data-theme="dark"] .cat-tag:hover {
  color: var(--primary);
  background: rgba(255, 133, 51, 0.1);
}
html[data-theme="dark"] .cat-tag.active {
  color: #fff;
  background: var(--primary);
}
html[data-theme="dark"] .confirmOrder .content,
html[data-theme="dark"] .payment .content {
  background: #242424;
}
html[data-theme="dark"] .confirmOrder .confirmOrder-header,
html[data-theme="dark"] .payment .payment-header {
  background: #242424;
}
html[data-theme="dark"] .section-shipment,
html[data-theme="dark"] .section-invoice,
html[data-theme="dark"] .section-count,
html[data-theme="dark"] .section-bar,
html[data-theme="dark"] .section-goods .goods-list,
html[data-theme="dark"] .address-body li {
  border-color: #333;
}
html[data-theme="dark"] .detail-row,
html[data-theme="dark"] .info-row,
html[data-theme="dark"] .meta-row {
  border-color: #333;
}
html[data-theme="dark"] .el-button--text {
  color: var(--primary);
}
html[data-theme="dark"] .status-box.pending { background: #3d3020; }
html[data-theme="dark"] .status-box.success { background: #203d20; }
html[data-theme="dark"] .status-box.fail { background: #3d2020; }
html[data-theme="dark"] .channel-item {
  border-color: #444;
  background: #2c2c2c;
}
html[data-theme="dark"] .channel-item.active {
  border-color: var(--primary);
  background: #2d2520;
}
html[data-theme="dark"] .admin-layout {
  background: #1a1a1a;
}
html[data-theme="dark"] .el-popconfirm {
  background: #2c2c2c;
}
body { line-height: 1.6; }
a { text-decoration: none; color: inherit; }
#app { min-width: 1000px; }

/* ===== 顶部通栏 ===== */
.top-bar {
  height: var(--top-bar-height);
  background: var(--bg);
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-muted);
  line-height: var(--top-bar-height);
}
.top-inner {
  max-width: var(--content-width);
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
}
.top-right { display: flex; align-items: center; gap: 0; }
.top-link { color: var(--text-secondary); transition: color 0.15s; padding: 0 2px; }
.top-link:hover { color: var(--primary); }
.top-sep { color: var(--border); margin: 0 8px; }
.top-user {
  display: flex; align-items: center; gap: 4px;
  color: var(--text-secondary); cursor: pointer; font-size: 12px;
}
.top-user:hover { color: var(--primary); }
.admin-link { color: var(--primary); }
.admin-link:hover { color: var(--primary-dark); }
.theme-toggle i { margin-right: 2px; }

/* ===== 主导航栏 ===== */
.main-header {
  height: var(--header-height);
  background: var(--bg-white);
  border-bottom: 2px solid var(--primary);
  position: sticky;
  top: 0;
  z-index: 1000;
  transition: box-shadow 0.2s, background-color 0.3s;
}
.main-header.is-fixed {
  box-shadow: var(--shadow-sm);
}
.header-inner {
  max-width: var(--content-width);
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 28px;
}

/* Logo */
.logo { display: flex; align-items: baseline; gap: 4px; flex-shrink: 0; }
.logo-text { font-size: 30px; font-weight: 800; color: var(--primary); letter-spacing: -1px; }
.logo-sub { font-size: 13px; color: var(--text-muted); font-weight: 400; }

/* 分类导航 */
.cat-nav { display: flex; gap: 4px; flex-shrink: 0; }
.cat-item {
  padding: 6px 18px;
  font-size: 15px;
  color: var(--text);
  font-weight: 500;
  border-radius: 4px;
  transition: all 0.15s;
}
.cat-item:hover { color: var(--primary); background: rgba(255,103,0,0.04); }
.cat-item.active { color: var(--primary); }

/* 搜索框 */
.search-wrap { flex: 1; max-width: 520px; }
.search-wrap .el-input__inner {
  border-color: var(--primary) !important;
  border-right: none;
  height: 38px;
}
.search-wrap .el-input__inner:focus {
  box-shadow: 0 0 0 2px rgba(255,103,0,0.15);
}
.search-wrap .el-input-group__append {
  background: var(--primary);
  border-color: var(--primary);
  color: #fff;
  font-size: 14px;
  padding: 0 20px;
  font-weight: 500;
}
.search-wrap .el-input-group__append:hover {
  background: var(--primary-dark);
}

/* 购物车入口 */
.cart-entry {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 20px;
  border: 1px solid var(--border);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.15s;
  position: relative;
  flex-shrink: 0;
  white-space: nowrap;
}
.cart-entry i { font-size: 18px; }
.cart-entry:hover, .cart-entry.active {
  border-color: var(--primary);
  color: var(--primary);
}
.cart-num {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  font-size: 11px;
  font-style: normal;
  color: #fff;
  background: #f56c6c;
  border-radius: 9px;
  padding: 0 5px;
}

/* 占位 */
.header-space { height: calc(var(--top-bar-height) + 2px); }

/* 主内容 */
.main-body { min-height: calc(100vh - var(--header-height) - var(--top-bar-height) - 200px); }

/* ===== 底栏 ===== */
.site-footer {
  background: var(--bg-white);
  margin-top: 40px;
  border-top: 1px solid var(--border);
  transition: background-color 0.3s;
}
.footer-services {
  max-width: var(--content-width);
  margin: 0 auto;
  display: flex;
  justify-content: space-around;
  padding: 36px 20px;
  border-bottom: 1px solid var(--border);
}
.service-item {
  display: flex;
  align-items: center;
  gap: 12px;
}
.service-item i {
  font-size: 32px;
  color: var(--primary);
}
.service-item p {
  font-size: 15px;
  color: var(--text);
  font-weight: 500;
  line-height: 1.3;
}
.service-item span {
  font-size: 12px;
  color: var(--text-muted);
}
.footer-bottom {
  text-align: center;
  padding: 24px 20px;
  color: var(--text-muted);
  font-size: 13px;
}
.footer-links { margin-bottom: 8px; }
.footer-links a { color: var(--text-muted); transition: color 0.15s; }
.footer-links a:hover { color: var(--primary); }
.footer-links span { margin: 0 10px; color: #ddd; }
</style>
