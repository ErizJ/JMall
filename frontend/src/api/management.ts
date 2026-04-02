import request from '@/utils/request'
import type { CombinationItem } from '@/types'

export const managementApi = {
  getCarousel: () => request.post('/resources/carousel'),
  addProduct: (data: object) => request.post('/management/addProduct', data),
  deleteProduct: (productId: number) =>
    request.post('/product/deleteProductById', { product_id: productId }),
  updateProduct: (data: object) => request.post('/product/updateProduct', data),
  getAllDiscounts: () => request.post('/management/getAllDiscounts'),
  addCombination: (data: Omit<CombinationItem, 'id'>) =>
    request.post('/management/addProductCombination', data),
  deleteCombination: (id: number) =>
    request.post('/management/deleteProductCombinationById', { id }),
  getProductsByCategoryName: (categoryName: string) =>
    request.post('/management/getProductsByCategoryName', {
      category_name: categoryName,
    }),
  getOrdersByUserName: (userName: string) =>
    request.post('/management/getOrdersByUserName', { userName }),
}
