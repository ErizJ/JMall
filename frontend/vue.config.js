/*
 * @Description: 配置文件
 */
module.exports = {
  publicPath: './',
  devServer: {
    open: true,
    proxy: {
      '/api/users': {
        target: 'http://localhost:8881/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/resources': {
        target: 'http://localhost:8886/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/product': {
        target: 'http://localhost:8882/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/user/shoppingCart': {
        target: 'http://localhost:8883/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/user/order': {
        target: 'http://localhost:8884/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/order': {
        target: 'http://localhost:8884/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/user/collect': {
        target: 'http://localhost:8885/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/management': {
        target: 'http://localhost:8886/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/payment': {
        target: 'http://localhost:8887/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      },
      '/api/aichat': {
        target: 'http://localhost:8888/',
        changeOrigin: true,
        pathRewrite: { '^/api': '' }
      }
    }
  }
}
