<template>
  <div id="app" name="app">
    <el-container style="min-width: 1570px">
      <!-- 顶部导航栏 -->
      <div class="j-header" style="overflow: hidden">
        <el-row :gutter="4">
          <el-col :span="4">
            <router-link to="/">
              <div class="logo">
                <img src="./assets/imgs/logo-white.png" alt="logo" />
              </div>
            </router-link>
          </el-col>

          <el-col :span="5">
            <div class="so">
              <el-input placeholder="请输入搜索内容" v-model="search">
                <el-button slot="append" @click="searchClick">搜索</el-button>
              </el-input>
            </div>
          </el-col>

          <el-col :span="1">
            <div class="grid-content"></div>
          </el-col>

          <el-col :span="10">
            <el-menu
              :default-active="activeIndex"
              class="el-menu-demo"
              mode="horizontal"
              router
            >
              <!-- router -->
              <el-menu-item index="home">
                <i class="el-icon-goods el-icon--left"></i>首页
              </el-menu-item>
              <el-menu-item index="collect">
                <i class="el-icon-star-off el-icon--left"></i>我的收藏
              </el-menu-item>
              <el-menu-item index="shoppingCart">
                <i class="el-icon-shopping-cart-2 el-icon--left"></i>购物车
              </el-menu-item>
              <el-menu-item index="order">
                <i class="el-icon-toilet-paper el-icon--left"></i>我的订单
              </el-menu-item>
              <el-menu-item index="manager" :disabled="disable">
                <i class="el-icon-setting el-icon--left"></i>系统管理
              </el-menu-item>
            </el-menu>
          </el-col>

          <el-col :span="2" style="height: 60px">
            <div class="j-userBar" style="height: 60px">
              <ul>
                <li v-if="!this.$store.getters.getUser">
                  <el-button type="text" @click="login">登录</el-button>
                  <span class="j-separator">&nbsp; | &nbsp;</span>
                  <el-button type="text" @click="register = true"
                    >注册</el-button
                  >
                </li>
                <li v-else>
                  <span>欢迎您，</span>
                  <el-popover placement="top" width="190" v-model="visible">
                    <p>确定退出登录吗？</p>
                    <div style="text-align: right; margin: 10px 0 0">
                      <el-button
                        size="mini"
                        type="text"
                        @click="visible = false"
                        >取消</el-button
                      >
                      <el-button type="primary" size="mini" @click="logout"
                        >确定</el-button
                      >
                    </div>
                    <el-button type="text" slot="reference">{{
                      this.$store.getters.getUser.userName
                    }}</el-button>
                  </el-popover>
                </li>
              </ul>
            </div>
          </el-col>

          <el-col :span="1">
            <div class="j-userAvatar">
              <el-avatar
                src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png"
              ></el-avatar>
            </div>
          </el-col>

          <el-col :span="1">
            <div class="grid-content"></div>
          </el-col>
        </el-row>
      </div>
      <!-- 顶部导航栏END -->

      <!-- 顶栏容器 -->
      <el-header>
        <el-menu
          :default-active="activeIndex"
          class="el-menu-demo"
          mode="horizontal"
          active-text-color="#409eff"
        ></el-menu>
      </el-header>
      <!-- 顶栏容器END -->

      <!-- 登录模块 -->
      <MyLogin></MyLogin>
      <!-- 注册模块 -->
      <MyRegister :register="register" @fromChild="isRegister"></MyRegister>

      <!-- 主要区域容器 -->
      <el-main style="padding: 0">
        <keep-alive>
          <router-view></router-view>
        </keep-alive>
      </el-main>
      <!-- 主要区域容器END -->

      <!-- 底栏容器 -->
      <el-footer style="padding: 0px">
        <div class="footer">
          <div class="j-promise-box">
            <div class="j-promise">
              <p class="text">
                <a class="icon1" href="http://github.com/ErizJ"
                  >7天无理由退换货</a
                >
                <a class="icon2" href="http://github.com/ErizJ"
                  >满99元全场免邮</a
                >
                <a
                  class="icon3"
                  style="margin-right: 0"
                  href="http://github.com/ErizJ"
                  >100%品质保证</a
                >
              </p>
            </div>
          </div>
          <div class="github">
            <div class="github-but"></div>
          </div>
          <div class="j-footer-help">
            <p>
              <a href="http://github.com/ErizJ">隐私政策</a>
              <span>|</span>
              <a href="http://github.com/ErizJ">使用条款</a>
              <span>|</span>
              <a href="http://github.com/ErizJ">法律信息</a>
            </p>
            <p class="coty">Copyright &copy; 2022 JMall Inc. 保留所有权利</p>
            <p class="coty">Power By ErizJ</p>
          </div>
        </div>
      </el-footer>
      <!-- 底栏容器END -->
      <!-- 回到顶部快捷键 -->
    </el-container>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import { mapGetters } from 'vuex'

export default {
  beforeUpdate() {
    this.activeIndex = this.$route.path
  },
  data() {
    return {
      activeIndex: '1', // 头部导航栏选中的标签
      search: '', // 搜索条件
      register: false, // 是否显示注册组件
      visible: false, // 是否退出登录
      flag: true,
      name:'',
    }
  },
  mounted() {
    
    // 获取浏览器localStorage，判断用户是否已经登录
    if (localStorage.getItem('user')) {
      // 如果已经登录，设置vuex登录状态
      this.setUser(JSON.parse(localStorage.getItem('user')))
      console.log(localStorage.getItem('user'));
      //this.checkUserIsManager(localStorage.getItem('user').userName)
    }

    // window.setTimeout(() => {
    //   this.$message({
    //     duration: 0,
    //     showClose: true,
    //     message: `
    //     <p>ohhh，亲亲</p>
    //     <p style="padding:10px 0">你已经看了很久了，喜欢就入手吧！</p>
    //     `,
    //     dangerouslyUseHTMLString: true,
    //     type: 'success',
    //   })
    // }, 10000 * 60)
  },

  computed: {
    ...mapGetters(['getUser', 'getNum']),
    disable(){
      console.log(this.flag);
      return this.flag
    }
  },

  watch: {
    // 获取vuex的登录状态
    getUser: function (val) {
      if (val === '') {
        // 用户没有登录
        this.setShoppingCart([])
      } else {
        // 用户已经登录,获取该用户的购物车信息
        this.checkUserIsManager(val.userName)
        this.$axios
          .post('/api/user/shoppingCart/getShoppingCart', {
            user_id: val.user_id,
          })
          .then((res) => {
            if (res.data.code === '001') {
              // 001 为成功, 更新vuex购物车状态
              this.setShoppingCart(res.data.shoppingCartData)
            } else {
              // 提示失败信息
              this.notifyError(res.data.msg)
            }
          })
          .catch((err) => {
            return Promise.reject(err)
          })
          
      }
    },
  },
  methods: {
    ...mapActions(['setUser', 'setShowLogin', 'setShoppingCart']),
    login() {
      // 点击登录按钮, 通过更改vuex的showLogin值显示登录组件
      this.setShowLogin(true)
      // this.$router.push({ path: '/login' })
    },
    // 退出登录
    logout() {
      this.visible = false
      // 清空本地登录信息
      localStorage.setItem('user', '')
      // 清空vuex登录信息
      this.setUser('')

      //发请求清除用户喜好商品类别数据
      this.$axios
        .post('/api/product/setCategoryHotZero')
        .then((res) => {
          if (res.data.code === '001') {
            this.notifySucceed('已成功删除用户喜好信息')
          } else {
            // 提示失败信息
            this.notifyError(res.data.msg)
          }
        })
        .catch((err) => {
          return Promise.reject(err)
        })

      this.notifySucceed('成功退出登录')
    },
    // 接收注册子组件传过来的数据
    isRegister(val) {
      this.register = val
    },
    // 点击搜索按钮
    searchClick() {
      if (this.search != '') {
        // 跳转到全部商品页面,并传递搜索条件
        this.$router.push({ path: '/goods', query: { search: this.search } })
        this.search = ''
      }
    },
    toShoppingCar() {
      this.$router.push({ path: '/shoppingCart' })
    },
    checkUserIsManager(name) {
      //向后端发请求，若此名称存在数据库的系统管理表中，则把flag设为true
      this.$axios
        .post('/api/users/isManager', {
          user_name: name,
        })
        .then((res) => {
          if (res.data.code === '001') {
            // 001 为成功
            this.flag = false
            console.log('----111', res.data.msg)
          } else {
            console.log('该用户不是系统管理者')
            this.flag = true
          }
          
        })
    },
  },
}
</script>

<style lang="css">
@import './reset.css';
html {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  background-color: #fff;
  min-width: 600px;
  /* overflow-x: hidden; */
}

.el-row {
  height: 60px;
  margin-bottom: 0px;
}

.el-col {
  overflow: hidden;
}

.grid-content {
  min-height: 60px;
  background-color: #fff;
}

/* 全局CSS */

#app .el-main {
  height: auto;
  width: auto;
}

a,
a:hover {
  text-decoration: none;
}
/* 全局CSS END */

/* 顶部导航栏CSS */
.j-header {
  background-color: #fff;
  width: 100%;
  height: 60px;
  position: fixed;
  top: 0;
  z-index: 9999;
  border-bottom: 2px solid rgb(230, 230, 230);
}

.j-userBar {
  height: 60px;
  background-color: #fff;
  float: right;
}

.j-userBar ul {
  float: left;
}
.j-userBar li {
  float: left;
  height: 60px;
  font-size: 16px;
  text-align: left;
  line-height: 60px;
  margin-right: 1px;
}
.j-userBar .j-separator {
  color: #b0b0b0;
  padding: 0 5px;
}

.j-userBar a:hover {
  color: #fff;
}

.j-userAvatar {
  height: 100%;
  padding: 10px 0px;
  background-color: transparent;
}

.logo {
  display: flex;
  width: 190px;
}

.logo img {
  display: flex;
  width: 100%;
}

.so {
  margin-top: 10px;
  width: auto;
}

.el-menu-demo {
  float: right;
}
/* 顶部导航栏CSS END */

/* 底栏容器CSS */
.footer {
  width: 100%;
  height: auto;
  text-align: center;
  background-color: rgb(247, 251, 253);
  padding-bottom: 20px;
}
.footer .j-promise-box {
  border-bottom: 1px solid #409eff;
  line-height: 145px;
}
.footer .j-promise-box {
  margin: 0 auto;
  border-bottom: 1px solid #409eff;
  line-height: 145px;
}
.footer .j-promise-box .j-promise p a {
  color: #409eff;
  font-size: 20px;
  margin-right: 210px;
  padding-left: 44px;
  height: 40px;
  display: inline-block;
  line-height: 40px;
  text-decoration: none;
  background: url('./assets/imgs/us-icon.png') no-repeat left 0;
}
.footer .github {
  height: 60px;
  line-height: 60px;
  margin-top: 20px;
}
.footer .github .github-but {
  width: 190px;
  text-align: center;
  height: 60px;
  margin: 0 auto;
  background: url('./assets/imgs/logo.png') no-repeat;
}
.footer .j-footer-help {
  text-align: center;
  color: #409eff;
}
.footer .j-footer-help p {
  margin: 20px 0 16px 0;
}

.footer .j-footer-help p a {
  color: #409eff;
  text-decoration: none;
}
.footer .j-footer-help p a:hover {
  color: #409eff;
}
.footer .j-footer-help p span {
  padding: 0 22px;
}
/* 底栏容器CSS END */
</style>
