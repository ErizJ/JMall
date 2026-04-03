// User
export interface User {
  userId: number
  userName: string
}

// Product
export interface Product {
  product_id: number
  product_name: string
  category_id: number
  product_title: string
  product_intro: string
  product_picture: string
  product_price: number
  product_selling_price: number
  product_num: number
  product_sales: number
  product_isPromotion: number
  product_hot: number
}

export interface Category {
  category_id: number
  category_name: string
  category_hot: number
}

export interface ProductPicture {
  id: number
  product_id: number
  product_picture: string
  intro: string
}

// Shopping Cart
export interface CartItem {
  id: number
  user_id: number
  product_id: number
  productName: string
  productImg: string
  price: number
  num: number
  maxNum: number
  check?: boolean
}

// Order
export interface OrderItem {
  id: number
  order_id: number
  user_id: number
  product_id: number
  productName: string
  productImg: string
  product_num: number
  product_price: number
  order_time: string
  status: number // 0=待支付 1=已支付 2=已取消 3=已退款
}

export interface AddOrderItem {
  product_id: number
  product_num: number
  product_price: number
}

// Payment
export interface PaymentItem {
  payment_no: string
  order_id: number
  amount: number
  channel: string
  status: number // 0=待支付 1=支付中 2=成功 3=失败 4=关闭 5=退款
  created_at: number
}

// Collect
export interface CollectItem {
  id: number
  user_id: number
  product_id: number
  category: string
  collect_time: string
}

// Carousel
export interface CarouselItem {
  carousel_id: number
  imgPath: string
  describes: string
}

// Discount combination
export interface CombinationItem {
  id: number
  main_product_id: number
  vice_product_id: number
  amountThreshold: number
  priceReductionRange: number
}

// API base response
export interface ApiResp {
  code: string
  message?: string
}

// Order status helpers
export const ORDER_STATUS_MAP: Record<number, string> = {
  0: '待支付',
  1: '已支付',
  2: '已取消',
  3: '已退款',
}

export const PAYMENT_STATUS_MAP: Record<number, string> = {
  0: '待支付',
  1: '支付中',
  2: '支付成功',
  3: '支付失败',
  4: '已关闭',
  5: '已退款',
}
