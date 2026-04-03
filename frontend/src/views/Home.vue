<!--
 * @Description: 首页组件
 -->
<template>
  <div class="home" id="home" name="home">
    <!-- 轮播图 -->
    <div class="block" style="padding-top: 20px">
      <el-carousel height="460px" autoplay :interval="2000" loop @click.native="carouselClick">
        <el-carousel-item v-for="item in carousel" :key="item.carousel_id">
          <img style="height: 460px" :src="$target + item.imgPath" :alt="item.describes" />
        </el-carousel-item>
      </el-carousel>
    </div>
    <!-- 轮播图END -->

    <div class="main-box">
      <div class="main">
        <!-- 促销商品展示区域 -->
        <div class="phone">
          <div class="box-hd">
            <div class="title">JMall 大促</div>
          </div>
          <div class="box-bd">
            <div class="promo-list">
              <router-link to>
                <img :src="$target + 'public/imgs/phone/phone.png'" />
              </router-link>
            </div>
            <div class="list">
              <MyList :list="promotionList" :isMore="true"></MyList>
            </div>
          </div>
        </div>
        <!-- 促销商品展示区域END -->

        <!-- 推荐商品展示区域 -->
        <div class="appliance" id="promo-menu">
          <div class="box-hd">
            <div class="title">JMall 推荐</div>
            <div class="more" id="more">
              <MyMenu :val="2" @fromChild="getChildMsg">
                <span slot="1">猜你喜欢</span>
                <span slot="2">平台热购TOP 7</span>
              </MyMenu>
            </div>
          </div>
          <div class="box-bd">
            <div class="promo-list">
              <ul>
                <li>
                  <img
                    :src="
                      $target + 'public/imgs/appliance/appliance-promo1.png'
                    "
                  />
                </li>
                <li>
                  <img
                    :src="
                      $target + 'public/imgs/appliance/appliance-promo2.png'
                    "
                  />
                </li>
              </ul>
            </div>
            <div class="list">
              <MyList :list="recommendList" :isMore="true"></MyList>
            </div>
          </div>
        </div>
        <!-- 推荐商品展示区域END -->

        <!-- 手机三件套展示区域 -->
        <div class="accessory" id="promo-menu">
          <div class="box-hd">
            <div class="title">手机三件套</div>
            <div class="more" id="more">
              <MyMenu :val="3" @fromChild="getChildMsg2">
                <span slot="1">手机</span>
                <span slot="2">保护套</span>
                <span slot="3">充电器</span>
              </MyMenu>
            </div>
          </div>
          <div class="box-bd">
            <div class="promo-list">
              <ul>
                <li>
                  <img
                    :src="
                      $target + 'public/imgs/accessory/accessory-promo1.png'
                    "
                    alt
                  />
                </li>
                <li>
                  <img
                    :src="
                      $target + 'public/imgs/accessory/accessory-promo2.png'
                    "
                    alt
                  />
                </li>
              </ul>
            </div>
            <div class="list">
              <MyList :list="accessoryList" :isMore="true"></MyList>
            </div>
          </div>
        </div>
        <!-- 手机三件套展示区域END -->
      </div>
    </div>
  </div>
</template>
<script>
import { mapActions } from 'vuex'

export default {
  data() {
    return {
      carousel: [], // 轮播图数据

      recommendList: [], // JMall推荐商品笼统称
      promotionList: [], //促销商品列表

      accessoryList: [], //配件商品列表
      //chargerList: '', //充电器商品列表
      //phoneList: '', // 手机商品列表
      //protectingShellList: '', // 保护套商品列表

      recommendActive: 2, // 家电当前选中的商品分类
      accessoryActive: 2, // 配件当前选中的商品分类
    }
  },
  watch: {
    deep: true,
    // JMall推荐的商品分类，响应不同的商品数据
    recommendActive: function (val) {
      if (val == 1) {
        // 1为个人推荐商品
        //console.log('Now is 猜你喜欢')
        this.getOneUserRecommendProduct()
        //this.recommendList = this.oneUserRecommendList
        return
      }
      if (val == 2) {
        // 2为大众推荐商品
        //console.log('Now is 大众推荐')
        this.getAllUserRecommendProduct()
        //this.recommendList = this.allUserRecommendList
        return
      }
    },
    accessoryActive: function (val) {
      if (val == 1) {
        // 1为手机商品
        this.getPhoneList()
        //this.accessoryList = this.phoneList
        return
      }
      if (val == 2) {
        // 2为保护套商品
        this.getProtectingShellList()
        //this.accessoryList = this.protectingShellList
        return
      }
      if (val == 3) {
        //3 为充电器商品
        this.getChargerList()
        //this.accessoryList = this.chargerList
        return
      }
    },
  },
  created() {
    // 获取轮播图数据
    this.$axios
      .post('/api/resources/carousel', {})
      .then((res) => {
        //console.log('--get carousel--');
        this.carousel = res.data.carousel
      })
      .catch((err) => {
        return Promise.reject(err)
      })

    //获取推荐商品列表数据
    //console.log('--getPromotionProduct--');
    this.getPromotionProduct()
    //console.log('--getProtectingShellList--');
    this.getProtectingShellList()
    //console.log('--getChargerList--');
    this.getChargerList()
    //console.log('--getPhoneList--');
    this.getPhoneList()

    this.recommendActive = 1
    this.accessoryActive = 1

    // 获取各类商品数据
    //this.getPromo('手机', 'phoneList')
    //this.getPromo('保护套', 'protectingShellList')
    //this.getPromo('充电器', 'chargerList')
  },
  methods: {
    ...mapActions(['setUser', 'getUser']),
    // 获取家电模块子组件传过来的数据
    getChildMsg(val) {
      this.recommendActive = val
    },
    // 获取配件模块子组件传过来的数据
    getChildMsg2(val) {
      this.accessoryActive = val
    },
    // 获取各类商品数据方法封装
    // getPromo(categoryName, val, api) {
    //   api = api != undefined ? api : '/api/product/getPromoProduct'
    //   this.$axios
    //     .post(api, {
    //       categoryName,
    //     })
    //     .then((res) => {
    //       this[val] = res.data.Product
    //     })
    //     .catch((err) => {
    //       return Promise.reject(err)
    //     })
    // },
    getPhoneList() {
      this.$axios
        .post('/api/product/getPhoneList')
        .then((res) => {
          this.accessoryList = res.data.category
          //console.log('get PhoneList数据',this.accessoryList)
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },
    getProtectingShellList() {
      this.$axios
        .post('/api/product/getProtectingShellList')
        .then((res) => {
          this.accessoryList = res.data.category
          //console.log('get ProtectingShellList数据',this.accessoryList)
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },
    getChargerList() {
      this.$axios
        .post('/api/product/getChargerList')
        .then((res) => {
          this.accessoryList = res.data.category
          //console.log('get ChargerList数据',this.accessoryList)
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },

    //获取促销商品数据
    getPromotionProduct() {
      this.$axios
        .post('/api/product/getPromotionProduct')
        .then((res) => {
          this.promotionList = res.data.category
          //console.log('get 促销商品数据',this.promotionList)
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },
    // 我懂了！！！之所以没数据是因为在后端返回给前端的不是category所以才会显示undefined
    // 直接查看后端返回的是什么字段，把category修改成对应字段就好了

    //获取获取单一用户推荐商品数据
    getOneUserRecommendProduct() {
      this.$axios
        .post('/api/product/getOneUserRecommendProduct')
        .then((res) => {
          this.recommendList = res.data.category
          //console.log('get 用户推荐商品数据')
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },

    //获取大众推荐商品数据
    getAllUserRecommendProduct() {
      this.$axios
        .post('/api/product/getAllUserRecommendProduct')
        .then((res) => {
          this.recommendList = res.data.category
          //console.log('get 大众推荐商品数据')
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },
    
    //轮播图点击事件,跳转到空调搜索界面
    carouselClick(){
      this.$router.push({
        path: '/goods',
        
      })
    }
  },
}
</script>
<style scoped>
/*
 * @Description: 首页css样式
 */
.home{
background-color: #f5f5f5;
}

.main-box {
  background-color: #f5f5f5;
  padding-bottom: 20px;
}

.main {
  margin: 0 auto;
  max-width: 1225px;
}

/* 轮播图CSS */
.block {
  margin: 0 auto;
  max-width: 1225px;
}

.el-carousel__item:nth-child(2n) {
  background-color: #99a9bf;
}

.el-carousel__item:nth-child(2n + 1) {
  background-color: #d3dce6;
}

/* 轮播图CSS END */

.box-hd {
  height: 58px;
  margin: 20px 0 0 0;
}

.box-hd .title {
  float: left;
  font-size: 22px;
  font-weight: 200;
  line-height: 58px;
  color: #333;
}

.box-hd .more {
  float: right;
}

.box-hd .more a {
  font-size: 16px;
  line-height: 58px;
  color: #424242;
}

.box-hd .more a:hover {
  color: #ff6700;
}

.box-bd {
  height: 615px;
}

.box-bd .promo-list {
  float: left;
  height: 615px;
  width: 234px;
}

.box-bd .promo-list li {
  z-index: 1;
  width: 234px;
  height: 300px;
  margin-bottom: 14.5px;
  -webkit-transition: all 0.2s linear;
  transition: all 0.2s linear;
}

.box-bd .promo-list li:hover {
  z-index: 2;
  -webkit-box-shadow: 0 15px 30px rgba(0, 0, 0, .1);
  box-shadow: 0 15px 30px rgba(0, 0, 0, .1);
  -webkit-transform: translate3d(0, -2px, 0);
  transform: translate3d(0, -2px, 0);
}

.box-bd .promo-list li img {
  width: 234px;
  height: 300px;
}

.box-bd .promo-list img {
  width: 234px;
}

.box-bd .list {
  float: left;
  height: 615px;
  width: 991px;
}
</style>
