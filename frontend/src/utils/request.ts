import axios from 'axios'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: true,
})

request.interceptors.request.use(
  (config) => {
    NProgress.start()
    return config
  },
  (error) => {
    NProgress.done()
    return Promise.reject(error)
  },
)

request.interceptors.response.use(
  (response) => {
    NProgress.done()
    const data = response.data
    if (data.code === '401') {
      const userStore = useUserStore()
      userStore.setShowLogin(true)
    }
    if (data.code === '500') {
      router.push('/error')
    }
    return response
  },
  (error) => {
    NProgress.done()
    return Promise.reject(error)
  },
)

export default request
