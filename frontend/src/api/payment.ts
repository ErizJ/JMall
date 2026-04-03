import request from '@/utils/request'

export const paymentApi = {
  create: (orderId: number, channel: string = 'mock') =>
    request.post('/payment/create', { order_id: orderId, channel }),

  getStatus: (paymentNo: string) =>
    request.post('/payment/status', { payment_no: paymentNo }),

  mockPay: (paymentNo: string) =>
    request.post('/payment/mock/pay', { payment_no: paymentNo }),

  refund: (paymentNo: string, refundAmount: number, reason: string) =>
    request.post('/payment/refund', { payment_no: paymentNo, refund_amount: refundAmount, reason }),

  list: (userId: number) =>
    request.post('/payment/list', { user_id: userId }),
}
