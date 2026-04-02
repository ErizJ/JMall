import request from '@/utils/request'

export const cartApi = {
  getCart: (userId: number) =>
    request.post('/user/shoppingCart/getShoppingCart', { user_id: userId }),
  addCart: (data: { user_id: number; product_id: number; num: number }) =>
    request.post('/user/shoppingCart/addShoppingCart', data),
  updateCart: (data: { user_id: number; product_id: number; num: number }) =>
    request.post('/user/shoppingCart/updateShoppingCart', data),
  deleteCart: (data: { user_id: number; product_id: number }) =>
    request.post('/user/shoppingCart/deleteShoppingCart', data),
  isExist: (data: { user_id: number; product_id: number }) =>
    request.post('/user/shoppingCart/isExistShoppingCart', data),
}
