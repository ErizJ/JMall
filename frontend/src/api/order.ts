import request from '@/utils/request'
import type { AddOrderItem } from '@/types'

export const orderApi = {
  getOrder: (userId: number) =>
    request.post('/user/order/getOrder', { user_id: userId }),
  addOrder: (data: { user_id: number; items: AddOrderItem[] }) =>
    request.post('/user/order/addOrder', data),
  getDetails: (orderId: number) =>
    request.post('/order/getDetails', { order_id: orderId }),
  deleteOrder: (orderId: number) =>
    request.post('/order/deleteOrderById', { order_id: orderId }),
}
