/*
 * @Description: Mock 数据
 * 所有 mock 响应的数据集中在这里，方便维护
 */

// 占位图基础 URL
const IMG = 'https://picsum.photos/seed'

// ========== 商品数据 ==========
const products = [
  {
    product_id: 1,
    product_name: 'Redmi K30',
    category_id: 1,
    product_title: '120Hz流速屏，全速热爱',
    product_intro: '120Hz高帧率流速屏 / 索尼6400万前后六摄',
    product_picture: `${IMG}/p1/200/200`,
    product_price: 1999,
    product_selling_price: 1599,
    product_num: 100,
    product_sales: 50,
  },
  {
    product_id: 2,
    product_name: 'Redmi K30 5G',
    category_id: 1,
    product_title: '双模5G,120Hz流速屏',
    product_intro: '双模5G / 高通骁龙765G / 120Hz高帧率流速屏',
    product_picture: `${IMG}/p2/200/200`,
    product_price: 2599,
    product_selling_price: 2599,
    product_num: 80,
    product_sales: 30,
  },
  {
    product_id: 3,
    product_name: '小米CC9 Pro',
    category_id: 1,
    product_title: '1亿像素,五摄四闪',
    product_intro: '1亿像素主摄 / 全场景五摄像头',
    product_picture: `${IMG}/p3/200/200`,
    product_price: 2799,
    product_selling_price: 2599,
    product_num: 60,
    product_sales: 20,
  },
  {
    product_id: 4,
    product_name: '小米电视4A 32英寸',
    category_id: 2,
    product_title: '人工智能系统，高清液晶屏',
    product_intro: '64位四核处理器 / 1GB+4GB大内存',
    product_picture: `${IMG}/p4/200/200`,
    product_price: 799,
    product_selling_price: 799,
    product_num: 200,
    product_sales: 90,
  },
  {
    product_id: 5,
    product_name: '小米笔记本Pro 15',
    category_id: 3,
    product_title: '全面屏，超轻薄',
    product_intro: '15.6英寸全面屏 / 第十代英特尔酷睿',
    product_picture: `${IMG}/p5/200/200`,
    product_price: 5999,
    product_selling_price: 5499,
    product_num: 50,
    product_sales: 15,
  },
  {
    product_id: 6,
    product_name: '小米USB充电器30W',
    category_id: 7,
    product_title: '多一种接口，多一种选择',
    product_intro: '双口输出 / 30W输出 / 可折叠插脚',
    product_picture: `${IMG}/p6/200/200`,
    product_price: 59,
    product_selling_price: 59,
    product_num: 500,
    product_sales: 200,
  },
  {
    product_id: 7,
    product_name: 'Redmi K20 保护壳',
    category_id: 5,
    product_title: '怪力魔王专属定制',
    product_intro: '优选PC材料，强韧张力，经久耐用',
    product_picture: `${IMG}/p7/200/200`,
    product_price: 39,
    product_selling_price: 39,
    product_num: 300,
    product_sales: 80,
  },
]

const categories = [
  { category_id: 1, category_name: '手机', category_hot: 10 },
  { category_id: 2, category_name: '电视机', category_hot: 5 },
  { category_id: 3, category_name: '笔记本', category_hot: 3 },
  { category_id: 4, category_name: '平板', category_hot: 2 },
  { category_id: 5, category_name: '手机壳', category_hot: 4 },
  { category_id: 6, category_name: '耳机', category_hot: 1 },
  { category_id: 7, category_name: '充电器', category_hot: 6 },
]

const carousel = [
  { carousel_id: 1, imgPath: `${IMG}/c1/1200/400`, describes: '新品首发' },
  { carousel_id: 2, imgPath: `${IMG}/c2/1200/400`, describes: '限时特惠' },
  { carousel_id: 3, imgPath: `${IMG}/c3/1200/400`, describes: '爆款推荐' },
]

// ========== 用户数据 ==========
const mockUser = {
  user_id: 1,
  userName: 'testuser',
  token: 'mock-jwt-token-for-testing',
}

// ========== 购物车 ==========
let cartIdCounter = 100
const shoppingCart = [
  {
    id: 1, productID: 1, productName: 'Redmi K30',
    productImg: `${IMG}/p1/200/200`, price: 1599, num: 1, maxNum: 50, check: false,
  },
  {
    id: 2, productID: 3, productName: '小米CC9 Pro',
    productImg: `${IMG}/p3/200/200`, price: 2599, num: 2, maxNum: 30, check: false,
  },
]

// ========== 订单 ==========
let orderIdCounter = Date.now() * 1000

const orders = [
  [
    {
      id: 1, order_id: 1001001, user_id: 1, product_id: 1,
      product_name: 'Redmi K30', product_picture: `${IMG}/p1/200/200`,
      product_num: 1, product_price: 1599,
      order_time: Date.now() - 86400000, status: 1,
    },
  ],
  [
    {
      id: 2, order_id: 1002001, user_id: 1, product_id: 3,
      product_name: '小米CC9 Pro', product_picture: `${IMG}/p3/200/200`,
      product_num: 2, product_price: 2599,
      order_time: Date.now() - 3600000, status: 0,
    },
  ],
]

// ========== 收藏 ==========
const collectList = [
  { ...products[0] },
  { ...products[4] },
]

// ========== 支付 ==========
let paymentNoCounter = 1
const payments = {}

export {
  products, categories, carousel,
  mockUser, shoppingCart, cartIdCounter,
  orders, orderIdCounter,
  collectList,
  payments, paymentNoCounter,
  IMG,
}
