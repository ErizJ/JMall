<template>
  <div class="admin-layout">
    <!-- 浅色侧边栏 -->
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <i class="el-icon-s-platform"></i>
        <span>管理后台</span>
      </div>
      <el-menu
        :default-active="$route.path"
        class="sidebar-menu"
        router
        unique-opened
      >
        <el-menu-item index="/goodsmanage">
          <i class="el-icon-s-goods"></i>
          <span slot="title">商品管理</span>
        </el-menu-item>
        <el-menu-item index="/ordersmanage">
          <i class="el-icon-s-order"></i>
          <span slot="title">订单管理</span>
        </el-menu-item>
        <el-menu-item index="/discountsmanage">
          <i class="el-icon-s-ticket"></i>
          <span slot="title">促销管理</span>
        </el-menu-item>
        <el-menu-item index="/usersmanage">
          <i class="el-icon-user-solid"></i>
          <span slot="title">用户管理</span>
        </el-menu-item>
      </el-menu>
      <div class="sidebar-footer">
        <el-button type="text" class="about-btn" @click="showAbout = true">
          <i class="el-icon-info"></i> 关于系统
        </el-button>
      </div>
    </aside>

    <!-- 主内容 -->
    <main class="admin-main">
      <router-view></router-view>
    </main>

    <!-- 关于系统弹框 -->
    <el-dialog title="关于 JMall" :visible.sync="showAbout" width="480px" center>
      <div class="about-dialog">
        <div class="about-logo">JMall</div>
        <p class="about-desc">小米风格电商平台 · 管理后台</p>
        <el-divider></el-divider>
        <div class="about-info">
          <div class="info-row"><span class="label">前端框架</span><span>Vue 2.6 + Element UI 2.x</span></div>
          <div class="info-row"><span class="label">后端框架</span><span>Go 1.23 + go-zero 1.10</span></div>
          <div class="info-row"><span class="label">数据库</span><span>MySQL 8.0 + Redis 7</span></div>
          <div class="info-row"><span class="label">版本</span><span>v1.0.0</span></div>
          <div class="info-row"><span class="label">作者</span><span>ErizJ</span></div>
        </div>
        <el-divider></el-divider>
        <p class="about-copy">Copyright © 2025 JMall Inc.</p>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" @click="showAbout = false">知道了</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  data() {
    return { showAbout: false }
  },
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: calc(100vh - var(--header-height, 60px) - 40px);
  background: #f0f2f5;
}

/* 浅色侧边栏 */
.admin-sidebar {
  width: 200px;
  background: var(--bg-white, #fff);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border, #e8e8e8);
}
.sidebar-brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text, #333);
  border-bottom: 1px solid var(--border, #f0f0f0);
}
.sidebar-brand i {
  font-size: 20px;
  color: var(--primary, #ff6700);
}

.sidebar-menu {
  border-right: none;
  flex: 1;
}
.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  font-size: 14px;
  color: var(--text-secondary, #666);
  margin: 2px 8px;
  border-radius: 6px;
}
.sidebar-menu .el-menu-item i {
  color: var(--text-muted, #999);
  margin-right: 6px;
}
.sidebar-menu .el-menu-item:hover {
  background: var(--bg, #f5f5f5);
  color: var(--primary, #ff6700);
}
.sidebar-menu .el-menu-item:hover i {
  color: var(--primary, #ff6700);
}
.sidebar-menu .el-menu-item.is-active {
  background: rgba(255, 103, 0, 0.08);
  color: var(--primary, #ff6700);
  font-weight: 500;
}
.sidebar-menu .el-menu-item.is-active i {
  color: var(--primary, #ff6700);
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border, #f0f0f0);
}
.about-btn {
  color: var(--text-muted, #999) !important;
  font-size: 13px;
  padding: 0;
}
.about-btn:hover { color: var(--primary, #ff6700) !important; }

/* 主内容 */
.admin-main {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  min-width: 0;
}

/* 关于弹框 */
.about-dialog { text-align: center; }
.about-logo {
  font-size: 32px;
  font-weight: 700;
  color: var(--primary, #ff6700);
  letter-spacing: -1px;
}
.about-desc {
  color: var(--text-muted, #999);
  font-size: 14px;
  margin-top: 4px;
}
.about-info {
  text-align: left;
  max-width: 280px;
  margin: 0 auto;
}
.about-info .info-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
  color: var(--text, #333);
}
.about-info .info-row .label {
  color: var(--text-muted, #999);
}
.about-copy {
  font-size: 12px;
  color: #bbb;
}
</style>
