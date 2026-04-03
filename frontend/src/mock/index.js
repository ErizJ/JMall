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
  userBehaviors,
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

  // ========== 推荐系统 ==========
  ['/api/recommend/guessYouLike', (req) => {
    const page = req.page || 1
    const pageSize = req.page_size || 20

    // 模拟推荐逻辑：根据用户行为偏好排序 + 热门兜底
    // 统计用户行为中的分类偏好
    const catScores = {}
    userBehaviors.forEach(b => {
      const weight = { 1: 1, 2: 2, 3: 3, 4: 5, 5: 4 }[b.behavior_type] || 1
      catScores[b.category_id] = (catScores[b.category_id] || 0) + weight
    })

    // 按偏好分类排序商品，偏好分类的商品排前面
    const reasons = ['猜你喜欢', '相似商品推荐', '和你口味相似的人也在看', '热门推荐']
    const scored = products.map((p, idx) => {
      const catScore = catScores[p.category_id] || 0
      const hotScore = (p.product_sales || 0) * 0.3
      const diversityScore = Math.random() * 20
      const finalScore = catScore * 3 + hotScore + diversityScore
      return {
        product_id: p.product_id,
        product_name: p.product_name,
        category_id: p.category_id,
        product_title: p.product_title,
        product_picture: p.product_picture,
        product_price: p.product_price,
        product_selling_price: p.product_selling_price,
        product_sales: p.product_sales || 0,
        product_hot: p.product_sales || 0,
        recommend_reason: catScore > 0 ? reasons[idx % 3] : reasons[3],
        score: Math.round(finalScore * 100) / 100,
      }
    })

    scored.sort((a, b) => b.score - a.score)

    const start = (page - 1) * pageSize
    const end = start + pageSize
    const pageItems = scored.slice(start, end)
    const hasMore = end < scored.length

    return {
      code: '200',
      recommendations: pageItems,
      has_more: hasMore,
    }
  }],

  ['/api/recommend/reportBehavior', (req) => {
    // 记录行为到内存（mock 模式下不持久化）
    if (req.product_id && req.behavior_type) {
      userBehaviors.push({
        user_id: 1,
        product_id: req.product_id,
        category_id: req.category_id || 0,
        behavior_type: req.behavior_type,
        behavior_time: Date.now(),
      })
    }
    return { code: '200' }
  }],

  ['/api/recommend/fillup', () => ({
    code: '200',
    cart_total: 4198,
    nearest_rule: { threshold: 5000, reduction: 500 },
    gap: 802,
    recommendations: products.slice(0, 6).map(p => ({
      product_id: p.product_id,
      product_name: p.product_name,
      category_id: p.category_id,
      product_title: p.product_title,
      product_picture: p.product_picture,
      product_price: p.product_price,
      product_selling_price: p.product_selling_price,
      product_sales: p.product_sales || 0,
      product_hot: p.product_sales || 0,
      recommend_reason: '差额精准推荐',
      score: 75.5,
    })),
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

  // ========== AI 聊天 ==========
  ['/api/aichat/chat', (req) => {
    const msg = (req.message || '').toLowerCase()
    let reply = ''

    if (msg.includes('热门') || msg.includes('推荐') || msg.includes('火')) {
      const hot = products.slice(0, 5)
      reply = '🔥 以下是目前最热门的商品推荐：\n\n'
      hot.forEach((p, i) => {
        reply += `${i + 1}. **${p.product_name}** - ${p.product_title}\n`
        reply += `   售价：¥${p.product_selling_price}`
        if (p.product_price > p.product_selling_price) {
          reply += `（原价 ¥${p.product_price}，立省 ¥${p.product_price - p.product_selling_price}）`
        }
        reply += `\n   库存：${p.product_num}件 | 已售：${p.product_sales || 0}件\n\n`
      })
      reply += '需要了解某个商品的详细信息吗？直接告诉我商品名就行~'
    } else if (msg.includes('促销') || msg.includes('打折') || msg.includes('优惠') || msg.includes('便宜')) {
      const promos = products.filter(p => p.product_price > p.product_selling_price)
      if (promos.length > 0) {
        reply = '🏷️ 当前正在促销的商品：\n\n'
        promos.forEach((p, i) => {
          const discount = Math.round((p.product_selling_price / p.product_price) * 100)
          reply += `${i + 1}. **${p.product_name}**\n`
          reply += `   原价 ¥${p.product_price} → 现价 ¥${p.product_selling_price}（${discount}折，省 ¥${p.product_price - p.product_selling_price}）\n\n`
        })
        reply += '心动不如行动，赶紧下单吧！'
      } else {
        reply = '目前暂时没有促销活动，不过我们的商品价格一直很实惠哦~ 需要我帮你找找特定商品吗？'
      }
    } else if (msg.includes('分类') || msg.includes('类别') || msg.includes('有什么')) {
      reply = '📦 我们商城目前有以下商品分类：\n\n'
      categories.forEach((c, i) => {
        const count = products.filter(p => p.category_id === c.category_id).length
        reply += `${i + 1}. **${c.category_name}** - ${count}件商品\n`
      })
      reply += '\n想看哪个分类的商品？告诉我分类名就行~'
    } else if (msg.includes('手机')) {
      const phones = products.filter(p => p.category_id === 1)
      if (phones.length > 0) {
        reply = '📱 为你找到以下手机：\n\n'
        phones.forEach((p, i) => {
          reply += `${i + 1}. **${p.product_name}** - ${p.product_title}\n`
          reply += `   售价：¥${p.product_selling_price} | 库存：${p.product_num}件\n\n`
        })
      } else {
        reply = '暂时没有找到手机类商品，要不看看其他分类？'
      }
    } else if (msg.includes('电视')) {
      const tvs = products.filter(p => p.category_id === 2)
      if (tvs.length > 0) {
        reply = '📺 为你找到以下电视：\n\n'
        tvs.forEach((p, i) => {
          reply += `${i + 1}. **${p.product_name}** - ${p.product_title}\n`
          reply += `   售价：¥${p.product_selling_price} | 库存：${p.product_num}件\n\n`
        })
      } else {
        reply = '暂时没有找到电视类商品。'
      }
    } else if (msg.includes('充电') || msg.includes('配件')) {
      const chargers = products.filter(p => p.category_id === 7)
      if (chargers.length > 0) {
        reply = '🔌 为你找到以下充电配件：\n\n'
        chargers.forEach((p, i) => {
          reply += `${i + 1}. **${p.product_name}** - ${p.product_title}\n`
          reply += `   售价：¥${p.product_selling_price} | 库存：${p.product_num}件\n\n`
        })
      } else {
        reply = '暂时没有找到充电配件。'
      }
    } else if (msg.includes('价格') || msg.includes('多少钱') || msg.includes('贵')) {
      // 尝试匹配具体商品名
      const matched = products.find(p => msg.includes(p.product_name.toLowerCase()))
      if (matched) {
        reply = `**${matched.product_name}** 的价格信息：\n\n`
        reply += `- 原价：¥${matched.product_price}\n`
        reply += `- 售价：¥${matched.product_selling_price}\n`
        if (matched.product_price > matched.product_selling_price) {
          reply += `- 优惠：立省 ¥${matched.product_price - matched.product_selling_price}\n`
        }
        reply += `- 库存：${matched.product_num}件\n`
        reply += `- 已售：${matched.product_sales || 0}件`
      } else {
        reply = '请告诉我具体的商品名称，我来帮你查价格~ 比如"Redmi K30 多少钱"'
      }
    } else {
      // 通用搜索：尝试在商品名/标题中匹配关键词
      const kw = msg.replace(/[？?！!。，,\s]/g, '')
      const matched = products.filter(p =>
        p.product_name.toLowerCase().includes(kw) ||
        p.product_title.toLowerCase().includes(kw) ||
        p.product_intro.toLowerCase().includes(kw)
      )
      if (matched.length > 0) {
        reply = `🔍 为你找到 ${matched.length} 个相关商品：\n\n`
        matched.forEach((p, i) => {
          reply += `${i + 1}. **${p.product_name}** - ${p.product_title}\n`
          reply += `   售价：¥${p.product_selling_price} | 库存：${p.product_num}件\n\n`
        })
      } else {
        reply = `你好！我是 JMall 智能购物助手 🛒\n\n我可以帮你：\n- 🔍 搜索商品（如"帮我找手机"）\n- 💰 查询价格（如"Redmi K30 多少钱"）\n- 🔥 推荐热门商品\n- 🏷️ 查看促销活动\n- 📦 浏览商品分类\n\n试试跟我说说你想找什么吧~`
      }
    }

    return { code: '200', reply: reply }
  }],

  ['/api/aichat/stream', (req) => {
    // stream 接口在 mock 模式下不会被调用（AiChat 组件会走 /api/aichat/chat）
    // 但为了完整性还是加上
    return { code: '200', reply: 'Mock 模式下请使用非流式接口' }
  }],
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
