import request from '@/utils/request'

export const userApi = {
  login: (data: { userName: string; password: string }) =>
    request.post('/users/login', data),

  register: (data: { userName: string; password: string; userPhoneNumber?: string }) =>
    request.post('/users/register', data),

  findUserName: (userName: string) =>
    request.post('/users/findUserName', { userName }),

  isManager: (userId: number) =>
    request.post('/users/isManager', { user_id: userId }),

  getDetails: (userId: number) =>
    request.post('/users/getDetails', { user_id: userId }),

  updateUser: (data: { user_id: number; userName?: string; userPhoneNumber?: string }) =>
    request.post('/users/updateUser', data),

  deleteUser: (userId: number) =>
    request.post('/users/deleteUserById', { user_id: userId }),

  logout: () =>
    request.post('/users/logout'),
}
