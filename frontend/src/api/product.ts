import request from '@/utils/request'

export const productApi = {
  getAllProduct: () => request.post('/product/getAllProduct'),
  getCategory: () => request.post('/product/getCategory'),
  getProductByCategory: (categoryId: number) =>
    request.post('/product/getProductByCategory', { category_id: categoryId }),
  search: (keyword: string) =>
    request.post('/product/getProductBySearch', { keyword }),
  getDetails: (productId: number) =>
    request.post('/product/getDetails', { product_id: productId }),
  getDetailsPicture: (productId: number) =>
    request.post('/product/getDetailsPicture', { product_id: productId }),
  getPromoProduct: (categoryName: string) =>
    request.post('/product/getPromoProduct', { category_name: categoryName }),
  getPromotionProduct: () => request.post('/product/getPromotionProduct'),
  getHotProduct: () => request.post('/product/getHotProduct'),
  getOneUserRecommend: (userId: number) =>
    request.post('/product/getOneUserRecommendProduct', { user_id: userId }),
  getAllUserRecommend: () => request.post('/product/getAllUserRecommendProduct'),
  getPhoneList: () => request.post('/product/getPhoneList'),
  getProtectingShellList: () => request.post('/product/getProtectingShellList'),
  getChargerList: () => request.post('/product/getChargerList'),
  setCategoryHotZero: () => request.post('/product/setCategoryHotZero'),
}
