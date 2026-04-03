<template>
  <div class="manage-page">
    <div class="manage-header">
      <h2>订单管理</h2>
      <el-input placeholder="按用户名搜索" v-model="search" size="small" style="width:260px" clearable @clear="searchit" @keyup.enter.native="searchit">
        <el-button slot="append" icon="el-icon-search" @click="searchit"></el-button>
      </el-input>
    </div>

    <!-- 统计 -->
    <div class="stat-row">
      <div class="stat-card">
        <i class="el-icon-s-order stat-icon"></i>
        <div><div class="stat-num">{{ groupedOrders.length }}</div><div class="stat-label">订单数</div></div>
      </div>
      <div class="stat-card">
        <i class="el-icon-document stat-icon" style="color:#67c23a"></i>
        <div><div class="stat-num">{{ rawData.length }}</div><div class="stat-label">商品行数</div></div>
      </div>
      <div class="stat-card">
        <i class="el-icon-money stat-icon" style="color:#ff6700"></i>
        <div><div class="stat-num price-num">¥{{ totalAmount }}</div><div class="stat-label">总金额</div></div>
      </div>
    </div>

    <!-- 订单列表（按 order_id 分组，每组一个卡片） -->
    <div v-loading="loading">
      <div v-if="pagedOrders.length > 0">
        <div class="order-block" v-for="order in pagedOrders" :key="order.orderId">
          <!-- 订单头 -->
          <div class="order-head">
            <div class="head-left">
              <span class="order-no">订单号：{{ order.orderId }}</span>
              <el-tag size="mini" :type="statusType(order.status)" class="status-tag">{{ statusLabel(order.status) }}</el-tag>
            </div>
            <div class="head-right">
              <span class="order-user"><i class="el-icon-user"></i> {{ order.userName }}</span>
              <span class="order-time"><i class="el-icon-time"></i> {{ order.time }}</span>
            </div>
          </div>

          <!-- 商品列表 -->
          <div class="order-products">
            <div class="product-row" v-for="(item, idx) in order.items" :key="idx">
              <div class="prod-img">
                <img v-if="item.product_picture" :src="$target + item.product_picture" />
                <div v-else class="img-placeholder"><i class="el-icon-picture-outline"></i></div>
              </div>
              <div class="prod-info">
                <p class="prod-name">{{ item.product_name || ('商品 #' + item.product_id) }}</p>
                <p class="prod-id">ID: {{ item.product_id }}</p>
              </div>
              <div class="prod-price">¥{{ item.product_price }}</div>
              <div class="prod-qty">x{{ item.product_num }}</div>
              <div class="prod-subtotal">¥{{ (item.product_price * item.product_num).toFixed(2) }}</div>
            </div>
          </div>

          <!-- 订单尾 -->
          <div class="order-foot">
            <div class="foot-left">
              共 <em>{{ order.totalNum }}</em> 件商品
            </div>
            <div class="foot-right">
              <span class="foot-total">合计：<em>¥{{ order.totalPrice.toFixed(2) }}</em></span>
              <el-button type="text" size="mini" @click="viewDetail(order)"><i class="el-icon-view"></i> 详情</el-button>
              <el-button type="text" size="mini" style="color:#f56c6c" @click="delOrder(order)"><i class="el-icon-delete"></i> 删除</el-button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <i class="el-icon-s-order"></i>
        <p>暂无订单数据</p>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagi" v-if="groupedOrders.length > pageSize">
      <el-pagination background layout="total,prev,pager,next" :total="groupedOrders.length" :page-size="pageSize" :current-page.sync="currentPage"></el-pagination>
    </div>

    <!-- 详情弹框 -->
    <el-dialog title="订单详情" :visible.sync="detailVisible" width="600px" top="6vh">
      <div v-if="detailOrder">
        <!-- 订单基本信息 -->
        <div class="detail-meta">
          <div class="meta-row"><span class="label">订单号</span><span class="mono">{{ detailOrder.orderId }}</span></div>
          <div class="meta-row"><span class="label">用户</span><span>{{ detailOrder.userName }}</span></div>
          <div class="meta-row"><span class="label">下单时间</span><span>{{ detailOrder.time }}</span></div>
          <div class="meta-row">
            <span class="label">状态</span>
            <el-tag size="small" :type="statusType(detailOrder.status)">{{ statusLabel(detailOrder.status) }}</el-tag>
          </div>
        </div>

        <el-divider content-position="left">商品明细</el-divider>

        <!-- 商品明细表格 -->
        <el-table :data="detailOrder.items" size="mini" border fit
          :header-cell-style="{ background:'#fafafa', color:'#333' }">
          <el-table-column label="商品" min-width="180">
            <template slot-scope="{ row }">
              <div class="detail-prod">
                <img v-if="row.product_picture" :src="$target + row.product_picture" class="detail-thumb" />
                <span>{{ row.product_name || ('ID: ' + row.product_id) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="单价" width="100" align="center">
            <template slot-scope="{ row }"><span>¥{{ row.product_price }}</span></template>
          </el-table-column>
          <el-table-column prop="product_num" label="数量" width="70" align="center"></el-table-column>
          <el-table-column label="小计" width="110" align="right">
            <template slot-scope="{ row }">
              <span class="price-text">¥{{ (row.product_price * row.product_num).toFixed(2) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- 合计 -->
        <div class="detail-summary">
          <span>共 {{ detailOrder.totalNum }} 件商品</span>
          <span>订单总额：<em class="price-text big">¥{{ detailOrder.totalPrice.toFixed(2) }}</em></span>
        </div>
      </div>
      <span slot="footer"><el-button type="primary" size="small" @click="detailVisible=false">关闭</el-button></span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  mounted() { this.getTableList() },
  data() {
    return {
      search: '', loading: false,
      rawData: [],
      currentPage: 1, pageSize: 5,
      detailVisible: false, detailOrder: null,
    }
  },
  computed: {
    // 按 order_id 分组
    groupedOrders() {
      const map = {}
      this.rawData.forEach(row => {
        const oid = row.order_id
        if (!map[oid]) {
          map[oid] = {
            orderId: oid,
            userId: row.user_id,
            userName: row.user_name || ('用户 ' + row.user_id),
            time: row.order_time,
            status: row.status,
            items: [],
            totalNum: 0,
            totalPrice: 0,
          }
        }
        map[oid].items.push(row)
        map[oid].totalNum += (row.product_num || 1)
        map[oid].totalPrice += (row.product_price || 0) * (row.product_num || 1)
      })
      // 按时间倒序
      return Object.values(map).sort((a, b) => {
        if (b.time > a.time) return 1
        if (b.time < a.time) return -1
        return 0
      })
    },
    pagedOrders() {
      const s = (this.currentPage - 1) * this.pageSize
      return this.groupedOrders.slice(s, s + this.pageSize)
    },
    totalAmount() {
      return this.groupedOrders.reduce((sum, o) => sum + o.totalPrice, 0).toFixed(2)
    },
  },
  methods: {
    searchit() { this.currentPage = 1; this.getTableList() },
    getTableList() {
      this.loading = true
      const api = this.search ? '/api/management/getOrdersByUserName' : '/api/management/getAllOrders'
      const data = this.search ? { user_name: this.search } : {}
      this.$axios.post(api, data).then(r => {
        this.rawData = r.data.category || []
      }).catch(() => {}).finally(() => { this.loading = false })
    },
    statusLabel(s) { return { 0: '待支付', 1: '已支付', 2: '已取消', 3: '已退款' }[s] || '—' },
    statusType(s) { return { 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info' },
    viewDetail(order) {
      this.detailOrder = order
      this.detailVisible = true
    },
    delOrder(order) {
      this.$confirm(
        `确定删除订单 ${order.orderId}？该订单包含 ${order.totalNum} 件商品。`,
        '删除订单',
        { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }
      ).then(() => {
        this.$axios.post('/api/order/deleteOrderById', { order_id: order.orderId }).then(() => {
          this.$message.success('已删除')
          this.getTableList()
        })
      }).catch(() => {})
    },
  },
}
</script>

<style scoped>
.manage-page {}
.manage-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.manage-header h2 { font-size: 18px; font-weight: 600; color: #333; }

/* 统计卡片 */
.stat-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card {
  background: #fff; border-radius: 8px; padding: 16px 20px;
  flex: 1; border: 1px solid #f0f0f0;
  display: flex; align-items: center; gap: 14px;
}
.stat-icon { font-size: 28px; color: #409eff; }
.stat-num { font-size: 22px; font-weight: 700; color: #333; }
.price-num { color: #ff6700; }
.stat-label { font-size: 12px; color: #999; margin-top: 2px; }

/* 订单卡片 */
.order-block {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
  margin-bottom: 12px;
  overflow: hidden;
  transition: box-shadow 0.2s;
}
.order-block:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); }

/* 订单头 */
.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}
.head-left { display: flex; align-items: center; gap: 10px; }
.order-no { font-size: 13px; color: #333; font-weight: 500; font-family: monospace; }
.status-tag { margin-left: 4px; }
.head-right { display: flex; align-items: center; gap: 16px; font-size: 12px; color: #999; }
.head-right i { margin-right: 2px; }

/* 商品行 */
.order-products { padding: 0 20px; }
.product-row {
  display: flex;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #f8f8f8;
}
.product-row:last-child { border-bottom: none; }
.prod-img {
  width: 56px; height: 56px; flex-shrink: 0; margin-right: 14px;
}
.prod-img img {
  width: 100%; height: 100%; object-fit: contain;
  border-radius: 6px; background: #f9f9f9; border: 1px solid #f0f0f0;
}
.img-placeholder {
  width: 100%; height: 100%; background: #f5f5f5; border-radius: 6px;
  display: flex; align-items: center; justify-content: center; color: #ddd; font-size: 20px;
}
.prod-info { flex: 1; min-width: 0; }
.prod-name { font-size: 13px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prod-id { font-size: 11px; color: #bbb; margin-top: 2px; }
.prod-price { width: 80px; text-align: center; font-size: 13px; color: #666; flex-shrink: 0; }
.prod-qty { width: 50px; text-align: center; font-size: 13px; color: #999; flex-shrink: 0; }
.prod-subtotal { width: 90px; text-align: right; font-size: 14px; font-weight: 600; color: #ff6700; flex-shrink: 0; }

/* 订单尾 */
.order-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
  font-size: 13px;
  color: #999;
}
.foot-left em { font-style: normal; color: #ff6700; font-weight: 500; }
.foot-right { display: flex; align-items: center; gap: 12px; }
.foot-total em { font-style: normal; font-size: 18px; font-weight: 700; color: #ff6700; }

/* 空状态 */
.empty-state { text-align: center; padding: 60px 0; color: #ccc; }
.empty-state i { font-size: 48px; display: block; margin-bottom: 8px; }
.empty-state p { font-size: 14px; }

/* 分页 */
.pagi { display: flex; justify-content: flex-end; padding: 16px 0; }

/* ===== 详情弹框 ===== */
.detail-meta {}
.meta-row {
  display: flex; justify-content: space-between;
  padding: 8px 0; font-size: 14px; color: #333;
  border-bottom: 1px solid #fafafa;
}
.meta-row .label { color: #999; }
.mono { font-family: monospace; font-size: 13px; }
.detail-prod { display: flex; align-items: center; gap: 8px; }
.detail-thumb { width: 32px; height: 32px; border-radius: 4px; object-fit: contain; background: #f9f9f9; flex-shrink: 0; }
.price-text { color: #ff6700; font-weight: 600; }
.detail-summary {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 0 0; margin-top: 12px;
  border-top: 1px solid #f0f0f0; font-size: 13px; color: #999;
}
.detail-summary .big { font-size: 20px; }
</style>
