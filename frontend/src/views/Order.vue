<!--
 * @Description: 我的订单页面组件
 -->
<template>
  <div class="order">
    <!-- 我的订单头部 -->
    <div class="order-header">
      <div class="order-header-content">
        <p>
          <i class="el-icon-s-order" style="font-size: 30px;color: #409EFF;"></i>
          我的订单
        </p>
      </div>
    </div>
    <!-- 我的订单头部END -->

    <!-- 我的订单主要内容 -->
    <div class="order-content" v-if="orders.length>0">
      <div class="content" v-for="(item,index) in orders" :key="index">
        <ul>
          <!-- 我的订单表头 -->
          <li class="order-info">
            <div class="order-id">订单编号: {{item[0].order_id}}</div>
            <div class="order-status">
              <el-tag :type="statusTagType(item[0].status)" size="small">
                {{ statusText(item[0].status) }}
              </el-tag>
            </div>
            <div class="order-time">订单时间: {{item[0].order_time | dateFormat}}</div>
          </li>
          <li class="header">
            <div class="pro-img"></div>
            <div class="pro-name">商品名称</div>
            <div class="pro-price">单价</div>
            <div class="pro-num">数量</div>
            <div class="pro-total">小计</div>
          </li>
          <!-- 我的订单表头END -->

          <!-- 订单列表 -->
          <li class="product-list" v-for="(product,i) in item" :key="i">
            <div class="pro-img">
              <router-link :to="{ path: '/goods/details', query: {productID:product.product_id} }">
                <img :src="$target + product.product_picture" />
              </router-link>
            </div>
            <div class="pro-name">
              <router-link
                :to="{ path: '/goods/details', query: {productID:product.product_id} }"
              >{{product.product_name}}</router-link>
            </div>
            <div class="pro-price">{{product.product_price}}元</div>
            <div class="pro-num">{{product.product_num}}</div>
            <div class="pro-total pro-total-in">{{product.product_price*product.product_num}}元</div>
          </li>
        </ul>
        <div class="order-bar">
          <div class="order-bar-left">
            <span class="order-total">
              共
              <span class="order-total-num">{{total[index].totalNum}}</span> 件商品
            </span>
          </div>
          <div class="order-bar-right">
            <span>
              <span class="total-price-title">合计：</span>
              <span class="total-price">{{total[index].totalPrice}}元</span>
            </span>
            <router-link
              v-if="item[0].status === 0"
              :to="{ path: '/payment', query: { orderId: item[0].order_id, totalPrice: total[index].totalPrice } }"
              class="btn-pay"
            >去支付</router-link>
          </div>
        </div>
      </div>
      <div style="margin-top:-40px;"></div>
    </div>
    <!-- 我的订单主要内容END -->

    <!-- 订单为空的时候显示的内容 -->
    <div v-else class="order-empty">
      <div class="empty">
        <h2>您的订单还是空的！</h2>
        <p>快去购物吧！</p>
      </div>
    </div>
    <!-- 订单为空的时候显示的内容END -->
  </div>
</template>
<script>
export default {
  data() {
    return {
      orders: [],
      total: [],
    }
  },
  activated() {
    this.$axios
      .post('/api/user/order/getOrder', {
        user_id: this.$store.getters.getUser.user_id,
      })
      .then((res) => {
        if (res.data.code === '001' || res.data.code === '200') {
          this.orders = res.data.orders
        } else {
          this.notifyError(res.data.msg)
        }
      })
      .catch((err) => {
        return Promise.reject(err)
      })
  },
  watch: {
    orders: function (val) {
      let total = []
      for (let i = 0; i < val.length; i++) {
        const element = val[i]
        let totalNum = 0
        let totalPrice = 0
        for (let j = 0; j < element.length; j++) {
          const temp = element[j]
          totalNum += temp.product_num
          totalPrice += temp.product_price * temp.product_num
        }
        total.push({ totalNum, totalPrice })
      }
      this.total = total
    },
  },
  methods: {
    statusText(status) {
      const map = { 0: '待支付', 1: '已支付', 2: '已取消', 3: '已退款' }
      return map[status] !== undefined ? map[status] : '未知'
    },
    statusTagType(status) {
      const map = { 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }
      return map[status] || 'info'
    },
  },
}
</script>
<style scoped>
.order {
  background-color: #f5f5f5;
  padding-bottom: 20px;
  height: auto;
}
.order .order-header {
  height: 64px;
  border-bottom: 2px solid #409EFF;
  background-color: #fff;
  margin-bottom: 20px;
}
.order .order-header .order-header-content {
  width: 1225px;
  margin: 0 auto;
}
.order .order-header p {
  font-size: 28px;
  line-height: 58px;
  float: left;
  font-weight: normal;
  color: #424242;
}
.order .content {
  width: 1225px;
  margin: 0 auto;
  background-color: #fff;
  margin-bottom: 50px;
}
.order .content ul {
  background-color: #fff;
  color: #424242;
  line-height: 85px;
}
.order .content ul .order-info {
  height: 60px;
  line-height: 60px;
  padding: 0 26px;
  color: #424242;
  border-bottom: 1px solid #409EFF;
}
.order .content ul .order-info .order-id {
  float: left;
  color: #409EFF;
}
.order .content ul .order-info .order-status {
  float: left;
  margin-left: 20px;
  line-height: 60px;
}
.order .content ul .order-info .order-time {
  float: right;
}
.order .content ul .header {
  height: 85px;
  padding-right: 26px;
  color: #424242;
}
.order .content ul .product-list {
  height: 85px;
  padding: 15px 26px 15px 0;
  border-top: 1px solid #e0e0e0;
}
.order .content ul .pro-img {
  float: left;
  height: 85px;
  width: 120px;
  padding-left: 80px;
}
.order .content ul .pro-img img {
  height: 80px;
  width: 80px;
}
.order .content ul .pro-name {
  float: left;
  width: 380px;
}
.order .content ul .pro-name a {
  color: #424242;
}
.order .content ul .pro-name a:hover {
  color: #409EFF;
}
.order .content ul .pro-price {
  float: left;
  width: 160px;
  padding-right: 18px;
  text-align: center;
}
.order .content ul .pro-num {
  float: left;
  width: 190px;
  text-align: center;
}
.order .content ul .pro-total {
  float: left;
  width: 160px;
  padding-right: 81px;
  text-align: right;
}
.order .content ul .pro-total-in {
  color: #409EFF;
}
.order .order-bar {
  width: 1185px;
  padding: 0 20px;
  border-top: 1px solid #409EFF;
  height: 50px;
  line-height: 50px;
  background-color: #fff;
}
.order .order-bar .order-bar-left {
  float: left;
}
.order .order-bar .order-bar-left .order-total {
  color: #757575;
}
.order .order-bar .order-bar-left .order-total-num {
  color: #409EFF;
}
.order .order-bar .order-bar-right {
  float: right;
  display: flex;
  align-items: center;
  height: 50px;
}
.order .order-bar .order-bar-right .total-price-title {
  color: #409EFF;
  font-size: 14px;
}
.order .order-bar .order-bar-right .total-price {
  color: #409EFF;
  font-size: 30px;
}
.order .order-bar .order-bar-right .btn-pay {
  display: inline-block;
  margin-left: 20px;
  padding: 0 24px;
  height: 36px;
  line-height: 36px;
  background: #409EFF;
  color: #fff;
  border-radius: 4px;
  font-size: 14px;
  text-decoration: none;
}
.order .order-bar .order-bar-right .btn-pay:hover {
  background: #66b1ff;
}
.order .order-empty {
  width: 1225px;
  margin: 0 auto;
}
.order .order-empty .empty {
  height: 300px;
  padding: 0 0 130px 558px;
  margin: 65px 0 0;
  background: url(../assets/imgs/cart-empty.png) no-repeat 124px 0;
  color: #b0b0b0;
  overflow: hidden;
}
.order .order-empty .empty h2 {
  margin: 70px 0 15px;
  font-size: 36px;
}
.order .order-empty .empty p {
  margin: 0 0 20px;
  font-size: 20px;
}
</style>
