/*
 * @Description: Mock 拦截器
 *
 * 通过环境变量 VUE_APP_USE_MOCK=true 启用。
 * 启用后，所有 /api/* 请求会被拦截并返回 mock 数据，不会发送真实网络请求。
 *
 * 用法：
 *   npm run serve                     → 真实后端
 *   VUE_APP_USE_MOCK=true npm run serve → mock 数据
 *   或在 .env.mock 文件中配置后：npm run mock
 */
import {
  products, categories, carousel,
  mockUser, shoppingCart,
  orders, orderIdCounter,
  collectList,
  payments,
  IMG,
} from './data'

// 模拟网络延迟（ms）
const DELAY = 200

function delay(data) {
  return new Promise(resolve => {
    setTimeout(() => resolve({ data }), DELAY)
  })
}

// 路由匹配表：[urlPattern, handler]
// handler 接收 (requestData) 返回 responseData
const routes = [
  // ========== 用户 ==========
  ['/api/users/login', (req) => ({
    code: '001',
    msg: '登录成功',
    user: { ...mockUser },
  })],

  ['/api/users/register', () => ({
    code: '001', msg: '注册成功',
  })],

  ['/api/users/findUserName', () => ({
    code: '001', msg: '用户名可用',
  })],

  ['/api/users/isManager', () => ({
    code: '001', msg: '是管理员',
  })],

  ['/api/users/getDetails', () => ({
    code: '001',
    user_id: mockUser.user_id,
    userName: mockUser.userName,
    userPhoneNumber: '138****0000',
  })],

  ['/api/users/logout', () => ({ code: '200' })],

  // ========== 商品 ==========
  ['/api/resources/carousel', () => ({
    code: '001', carousel: carousel,
  })],

  ['/api/product/getAllProduct', (req) => {
    const page = req.currentPage || 1
    const size = req.pageSize || 15
    const start = (page - 1) * size
    return {
      code: '001',
      Product: products.slice(start, start + size),
      total: products.length,
    }
  }],

  ['/api/product/getCategory', () => ({
    code: '001', category: categories,
  })],

  ['/api/product/getProductByCategory', (req) => {
    const ids = req.categoryID || []
    const filtered = ids.length > 0
      ? products.filter(p => ids.includes(p.category_id))
      : products
    const page = req.currentPage || 1
    const size = req.pageSize || 15
    const start = (page - 1) * size
    return {
      code: '001',
      Product: filtered.slice(start, start + size),
      total: filtered.length,
    }
  }],

  ['/api/product/getProductBySearch', (req) => {
    const kw = (req.search || '').toLowerCase()
    const filtered = products.filter(p =>
      p.product_name.toLowerCase().includes(kw) ||
      p.product_title.toLowerCase().includes(kw)
    )
    const page = req.currentPage || 1
    const size = req.pageSize || 15
    const start = (page - 1) * size
    return {
      code: '001',
      Product: filtered.slice(start, start + size),
      total: filtered.length,
    }
  }],

  ['/api/product/getDetails', (req) => {
    const p = products.find(x => x.product_id === req.productID)
    return { code: '001', Product: p ? [p] : [] }
  }],

  ['/api/product/getDetailsPicture', (req) => ({
    code: '001',
    ProductPicture: [
      { product_picture: `${IMG}/pic${req.productID}a/560/560`, intro: '正面' },
      { product_picture: `${IMG}/pic${req.productID}b/560/560`, intro: '背面' },
    ],
  })],

  ['/api/product/getHotProduct', () => ({
    code: '001', category: products.slice(0, 7),
  })],

  ['/api/product/getPromotionProduct', () => ({
    code: '001', category: products.slice(0, 7),
  })],

  ['/api/product/getOneUserRecommendProduct', () => ({
    code: '001', category: products.slice(0, 7),
  })],

  ['/api/product/getAllUserRecommendProduct', () => ({
    code: '001', category: products.slice(0, 7),
  })],

  ['/api/product/getPhoneList', () => ({
    code: '001', category: products.filter(p => p.category_id === 1).slice(0, 7),
  })],

  ['/api/product/getProtectingShellList', () => ({
    code: '001', category: products.filter(p => p.category_id === 5).slice(0, 7),
  })],

  ['/api/product/getChargerList', () => ({
    code: '001', category: products.filter(p => p.category_id === 7).slice(0, 7),
  })],

  ['/api/product/setCategoryHotZero', () => ({
    code: '001', msg: '重置成功',
  })],

  // ========== 购物车 ==========
  ['/api/user/shoppingCart/getShoppingCart', () => ({
    code: '001', shoppingCartData: [...shoppingCart],
  })],

  ['/api/user/shoppingCart/addShoppingCart', (req) => {
    const p = products.find(x => x.product_id === req.product_id)
    if (!p) return { code: '002', msg: '商品不存在' }
    const existing = shoppingCart.find(c => c.productID === req.product_id)
    if (existing) {
      existing.num++
      return { code: '002', msg: '该商品已在购物车，数量+1' }
    }
    const item = {
      id: Date.now(),
      productID: p.product_id,
      productName: p.product_name,
      productImg: p.product_picture,
      price: p.product_selling_price,
      num: 1,
      maxNum: Math.floor(p.product_num / 2),
      check: false,
    }
    shoppingCart.unshift(item)
    return { code: '001', msg: '添加购物车成功', shoppingCartData: [item] }
  }],

  ['/api/user/shoppingCart/updateShoppingCart', (req) => {
    const item = shoppingCart.find(c => c.productID === req.product_id)
    if (item) item.num = req.num
    return { code: '001', msg: '更新成功' }
  }],

  ['/api/user/shoppingCart/deleteShoppingCart', (req) => {
    const idx = shoppingCart.findIndex(c => c.productID === req.product_id)
    if (idx >= 0) shoppingCart.splice(idx, 1)
    return { code: '001', msg: '删除成功' }
  }],

  ['/api/user/shoppingCart/isExistShoppingCart', (req) => {
    const exists = shoppingCart.some(c => c.productID === req.product_id)
    return { code: exists ? '001' : '002', msg: exists ? '已在购物车' : '不在购物车' }
  }],

  // ========== 订单 ==========
  ['/api/user/order/addOrder', (req) => {
    const orderId = Date.now() * 1000 + Math.floor(Math.random() * 1000)
    const items = (req.items || []).map((item, i) => {
      const p = products.find(x => x.product_id === item.product_id)
      return {
        id: Date.now() + i,
        order_id: orderId,
        user_id: 1,
        product_id: item.product_id,
        product_name: p ? p.product_name : 'Unknown',
        product_img: p ? p.product_picture : '',
        product_num: item.product_num,
        product_price: p ? p.product_selling_price : item.product_price,
        order_time: new Date().toISOString().replace('T', ' ').slice(0, 19),
        status: 0,
      }
    })
    let totalAmount = 0
    let itemCount = 0
    items.forEach(it => { totalAmount += it.product_price * it.product_num; itemCount += it.product_num })
    const group = {
      order_id: orderId, user_id: 1, status: 0,
      order_time: items[0] ? items[0].order_time : '',
      item_count: itemCount, total_amount: totalAmount,
      items: items,
    }
    orders.unshift(group)
    return { code: '200', msg: '下单成功', order_id: orderId }
  }],

  ['/api/user/order/getOrder', () => ({
    code: '200', orders: orders,
  })],

  ['/api/order/getDetails', (req) => {
    const found = orders.find(g => g.order_id === req.order_id)
    return { code: '200', order: found || {} }
  }],

  ['/api/order/deleteOrderById', (req) => {
    const idx = orders.findIndex(g => g.order_id === req.order_id)
    if (idx >= 0) orders.splice(idx, 1)
    return { code: '200', msg: '删除成功' }
  }],

  // ========== 收藏 ==========
  ['/api/user/collect/addCollect', (req) => {
    const p = products.find(x => x.product_id === req.product_id)
    if (p && !collectList.find(c => c.product_id === req.product_id)) {
      collectList.push({ ...p })
    }
    return { code: '001', msg: '收藏成功' }
  }],

  ['/api/user/collect/getCollect', () => ({
    code: '001', collectList: [...collectList],
  })],

  ['/api/user/collect/deleteCollect', (req) => {
    const idx = collectList.findIndex(c => c.product_id === req.product_id)
    if (idx >= 0) collectList.splice(idx, 1)
    return { code: '001', msg: '取消收藏成功' }
  }],

  // ========== 支付 ==========
  ['/api/payment/create', (req) => {
    const no = 'MOCK-PAY-' + Date.now()
    payments[no] = { status: 0, order_id: req.order_id }
    return { code: '200', payment_no: no, pay_url: '' }
  }],

  ['/api/payment/status', (req) => {
    const p = payments[req.payment_no]
    return {
      code: '200',
      payment_no: req.payment_no,
      status: p ? p.status : -1,
    }
  }],

  ['/api/payment/mock/pay', (req) => {
    if (payments[req.payment_no]) {
      payments[req.payment_no].status = 2
      // 同步更新订单状态
      const orderId = payments[req.payment_no].order_id
      orders.forEach(group => {
        if (group.order_id === orderId) {
          group.status = 1
          group.items.forEach(item => { item.status = 1 })
        }
      })
    }
    return { code: '200', msg: '支付成功' }
  }],

  ['/api/payment/list', () => ({
    code: '200', payments: [],
  })],

  ['/api/payment/refund', () => ({
    code: '200', refund_no: 'MOCK-REFUND-' + Date.now(),
  })],

  // ========== 管理后台 ==========
  ['/api/management/getAllOrders', () => {
    // Flatten grouped orders for management view
    const flat = []
    orders.forEach(g => {
      g.items.forEach(item => {
        flat.push({ ...item, user_name: 'user' + g.user_id, product_picture: item.product_img })
      })
    })
    return { code: '001', category: flat }
  }],

  ['/api/management/getOrdersByUserName', () => {
    const flat = []
    orders.forEach(g => {
      g.items.forEach(item => {
        flat.push({ ...item, user_name: 'user' + g.user_id, product_picture: item.product_img })
      })
    })
    return { code: '001', category: flat }
  }],

  ['/api/management/getAllUsers', () => ({
    code: '001',
    category: [
      { user_id: 1, user_name: 'testuser', user_phone_number: '138****0000' },
      { user_id: 2, user_name: 'admin', user_phone_number: '139****1111' },
    ],
  })],

  ['/api/management/getAllProducts', () => ({
    code: '001', category: products,
  })],

  ['/api/management/getProductsByCategoryName', (req) => {
    const cat = categories.find(c => c.category_name === req.category_name)
    const filtered = cat ? products.filter(p => p.category_id === cat.category_id) : []
    return { code: '001', category: filtered }
  }],

  ['/api/management/getAllDiscounts', () => ({
    code: '001', category: [],
  })],

  ['/api/management/addProduct', () => ({
    code: '001', msg: '添加成功',
  })],

  ['/api/management/addProductCombination', () => ({
    code: '001', msg: '添加成功',
  })],

  ['/api/management/deleteProductCombinationById', () => ({
    code: '001', msg: '删除成功',
  })],

  ['/api/management/getProductCombination', () => ({
    code: '001',
    category: [{
      main_product_id: 1, vice_product_id: 6,
      amountThreshold: 2, priceReductionRange: 20,
    }],
  })],

  ['/api/management/getCombinationProductList', (req) => {
    const p = products.find(x => x.product_id === req.product_id)
    return { code: '001', category: p ? [p] : [] }
  }],

  ['/api/product/updateProduct', () => ({
    code: '001', msg: '更新成功',
  })],

  ['/api/product/deleteProductById', () => ({
    code: '001', msg: '删除成功',
  })],

  ['/api/users/updateUser', () => ({
    code: '001', msg: '更新成功',
  })],

  ['/api/users/deleteUserById', () => ({
    code: '001', msg: '删除成功',
  })],

  ['/api/users/getUserByName', () => ({
    code: '001',
    category: [{ user_id: 1, user_name: 'testuser', user_phone_number: '138****0000' }],
  })],
]

/**
 * 安装 mock 拦截器到 Axios 实例
 */
export function setupMock(axios) {
  console.log('%c[Mock] Mock 模式已启用，所有 API 请求将返回模拟数据', 'color: #e6a23c; font-weight: bold')

  axios.interceptors.request.use(config => {
    const url = config.url
    const reqData = config.data || {}

    // 查找匹配的 mock 路由（最长前缀匹配）
    let matched = null
    let matchLen = 0
    for (const [pattern, handler] of routes) {
      if (url.startsWith(pattern) && pattern.length > matchLen) {
        matched = handler
        matchLen = pattern.length
      }
    }

    if (matched) {
      // 取消真实请求，通过 adapter 返回 mock 数据
      config.adapter = () => {
        const body = typeof reqData === 'string' ? JSON.parse(reqData) : reqData
        const result = matched(body)
        return delay(result).then(res => ({
          data: res.data,
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }))
      }
    }

    return config
  })
}
