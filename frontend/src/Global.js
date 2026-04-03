/*
 * @Description: 全局变量
 */
exports.install = function (Vue) {

  // Mock 模式下图片已经是完整 URL，不需要前缀
  if (process.env.VUE_APP_USE_MOCK === 'true') {
    Vue.prototype.$target = "";
  } else {
    // http://101.132.181.9:3000/ 线上后端地址
    Vue.prototype.$target = "http://localhost:3000/"; // 本地后端地址
  }


  // 封装提示成功的弹出框
  Vue.prototype.notifySucceed = function (msg) {

    //调用 Element-UI 中的消息通知组件现实当前登录的状态
    this.$notify({
      title: "成功",
      message: msg,
      type: "success",
      offset: 100
    });
  };
  // 封装提示失败的弹出框
  Vue.prototype.notifyError = function (msg) {
    this.$notify.error({
      title: "错误",
      message: msg,
      offset: 100
    });
  };
}