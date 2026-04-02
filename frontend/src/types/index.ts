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
}

export interface AddOrderItem {
  product_id: number
  product_num: number
  product_price: number
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
