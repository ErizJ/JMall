<!--
 * @Description: 全部商品页面
 -->
<template>
  <div class="goods-page">
    <!-- 顶部筛选栏 -->
    <div class="filter-bar">
      <div class="filter-inner">
        <!-- 面包屑 -->
        <el-breadcrumb separator="/" class="crumb">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item>全部商品</el-breadcrumb-item>
          <el-breadcrumb-item v-if="search">搜索：{{ search }}</el-breadcrumb-item>
        </el-breadcrumb>

        <!-- 分类标签 -->
        <div class="category-row">
          <span class="cat-label">分类</span>
          <div class="cat-tags">
            <span
              v-for="item in categoryList"
              :key="item.category_id"
              :class="['cat-tag', activeName === '' + item.category_id ? 'active' : '']"
              @click="switchCategory(item.category_id)"
            >{{ item.category_name }}</span>
          </div>
        </div>

        <!-- 搜索结果提示 -->
        <div class="search-tip" v-if="search">
          <span>搜索 "<em>{{ search }}</em>" 共找到 <em>{{ total }}</em> 件商品</span>
          <el-button type="text" size="mini" @click="clearSearch">清除搜索</el-button>
        </div>

        <!-- 排序 & 统计 -->
        <div class="sort-row">
          <div class="sort-left">
            <span class="result-count">共 <em>{{ total }}</em> 件商品</span>
          </div>
          <div class="sort-right">
            <span class="view-label">每页</span>
            <el-select v-model="pageSize" size="mini" style="width:80px" @change="onPageSizeChange">
              <el-option :value="15" label="15"></el-option>
              <el-option :value="30" label="30"></el-option>
              <el-option :value="60" label="60"></el-option>
            </el-select>
            <span class="view-label">件</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 商品列表 -->
    <div class="goods-content">
      <div class="goods-inner">
        <div v-if="product && product.length > 0" class="product-area">
          <MyList :list="product"></MyList>
        </div>
        <div v-else class="empty-state">
          <i class="el-icon-search"></i>
          <p>抱歉，没有找到相关商品</p>
          <el-button type="primary" size="small" round @click="clearSearch">查看全部商品</el-button>
        </div>

        <!-- 分页 -->
        <div class="pagination" v-if="total > 0">
          <el-pagination
            background
            layout="prev, pager, next, jumper"
            :page-size="pageSize"
            :total="total"
            :current-page.sync="currentPage"
            @current-change="currentChange"
          ></el-pagination>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      categoryList: [],
      categoryID: [],
      product: [],
      total: 0,
      pageSize: 15,
      currentPage: 1,
      activeName: '-1',
      search: '',
    }
  },
  created() {
    this.getCategory()
  },
  activated() {
    this.activeName = '-1'
    this.total = 0
    this.currentPage = 1
    if (Object.keys(this.$route.query).length === 0) {
      this.categoryID = []
      this.activeName = '0'
      return
    }
    if (this.$route.query.categoryID != undefined) {
      this.categoryID = this.$route.query.categoryID
      if (this.categoryID.length === 1) {
        this.activeName = '' + this.categoryID[0]
      }
      return
    }
    if (this.$route.query.search != undefined) {
      this.search = this.$route.query.search
    }
  },
  watch: {
    activeName(val) {
      if (val === '-1') return
      this.categoryID = val == 0 ? [] : [Number(val)]
      this.total = 0
      this.currentPage = 1
      this.$router.push({ path: '/goods', query: { categoryID: this.categoryID } })
    },
    search(val) {
      if (val) this.getProductBySearch()
    },
    categoryID() {
      this.getData()
      this.search = ''
    },
    $route(val) {
      if (val.path === '/goods' && val.query.search != undefined) {
        this.activeName = '-1'
        this.currentPage = 1
        this.total = 0
        this.search = val.query.search
      }
    },
  },
  methods: {
    switchCategory(id) {
      this.activeName = '' + id
    },
    clearSearch() {
      this.search = ''
      this.activeName = '0'
      this.$router.push({ path: '/goods' })
    },
    onPageSizeChange() {
      this.currentPage = 1
      if (this.search) this.getProductBySearch()
      else this.getData()
    },
    backtop() {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    currentChange(page) {
      this.currentPage = page
      if (this.search) this.getProductBySearch()
      else this.getData()
      this.backtop()
    },
    getCategory() {
      this.$axios.post('/api/product/getCategory', {}).then((res) => {
        const cate = res.data.category || []
        cate.unshift({ category_id: 0, category_name: '全部' })
        this.categoryList = cate
      }).catch(() => {})
    },
    getData() {
      const api = this.categoryID.length === 0
        ? '/api/product/getAllProduct'
        : '/api/product/getProductByCategory'
      this.$axios.post(api, {
        categoryID: this.categoryID,
        currentPage: this.currentPage,
        pageSize: this.pageSize,
      }).then((res) => {
        this.product = res.data.Product || []
        this.total = res.data.total || 0
      }).catch(() => {})
    },
    getProductBySearch() {
      this.$axios.post('/api/product/getProductBySearch', {
        search: this.search,
        currentPage: this.currentPage,
        pageSize: this.pageSize,
      }).then((res) => {
        this.product = res.data.Product || []
        this.total = res.data.total || 0
      }).catch(() => {})
    },
  },
}
</script>

<style scoped>
.goods-page {
  background: var(--bg, #f5f5f5);
  min-height: calc(100vh - 260px);
}

/* ===== 顶部筛选栏 ===== */
.filter-bar {
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  padding: 16px 0 0;
}
.filter-inner {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px;
}

/* 面包屑 */
.crumb {
  margin-bottom: 16px;
}

/* 分类标签 */
.category-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f5f5f5;
}
.cat-label {
  font-size: 13px;
  color: #999;
  flex-shrink: 0;
  line-height: 30px;
}
.cat-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.cat-tag {
  display: inline-block;
  padding: 4px 16px;
  font-size: 13px;
  color: #666;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}
.cat-tag:hover {
  color: var(--primary, #ff6700);
  background: rgba(255, 103, 0, 0.04);
}
.cat-tag.active {
  color: #fff;
  background: var(--primary, #ff6700);
  border-color: var(--primary, #ff6700);
}

/* 搜索提示 */
.search-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  font-size: 13px;
  color: #666;
}
.search-tip em {
  font-style: normal;
  color: var(--primary, #ff6700);
  font-weight: 500;
}

/* 排序行 */
.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
}
.result-count {
  font-size: 13px;
  color: #999;
}
.result-count em {
  font-style: normal;
  color: var(--primary, #ff6700);
  font-weight: 600;
}
.sort-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.view-label {
  font-size: 12px;
  color: #999;
}

/* ===== 商品列表区 ===== */
.goods-content {
  padding: 20px 0 40px;
}
.goods-inner {
  max-width: var(--content-width, 1226px);
  margin: 0 auto;
  padding: 0 20px;
}
.product-area {
  min-height: 400px;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 0;
}
.empty-state i {
  font-size: 56px;
  color: #ddd;
}
.empty-state p {
  font-size: 15px;
  color: #999;
  margin: 12px 0 20px;
}

/* 分页 */
.pagination {
  text-align: center;
  padding: 24px 0 0;
}
</style>
