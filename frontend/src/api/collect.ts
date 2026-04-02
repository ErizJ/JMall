import request from '@/utils/request'

export const collectApi = {
  getCollect: (userId: number) =>
    request.post('/user/collect/getCollect', { user_id: userId }),
  addCollect: (data: { user_id: number; product_id: number; category: string }) =>
    request.post('/user/collect/addCollect', data),
  deleteCollect: (data: { user_id: number; product_id: number }) =>
    request.post('/user/collect/deleteCollect', data),
}
