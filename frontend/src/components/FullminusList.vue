<!-- 
* @Description:满减助手列表组件
-->
<template>
  <el-card class="box-card">
    <div slot="header" class="clearfix">
      <span>
        <span style="color: #409eff; font-size: 24px">Jmall</span>
        为你精心推荐满减组合，省钱又省时
      </span>
    </div>
    <el-card class="box-card" shadow="hover">
      <img :src="$target + item.productImg" alt />
      <h2>{{ item.productName }}</h2>
      <!-- <h3>{{ item.productName }}</h3> -->
      <p>
        <span>{{ item.price }}元</span>
        <!-- <span
          v-show="item.product_price == item.product_selling_price"
          class="del"
        >{{ item.price }}元</span>-->
      </p>
    </el-card>
    <el-card class="box-card" shadow="hover">
      <img :src="$target + this.pairProduct.product_picture" alt />
      <h2>{{ this.pairProduct.product_name }}</h2>
      <h3>{{ this.pairProduct.product_title }}</h3>
      <p>
        <span>{{ this.pairProduct.product_selling_price }}元</span>
        <span
          v-show="
            this.pairProduct.product_price !=
            this.pairProduct.product_selling_price
          "
          class="del"
          >{{ this.pairProduct.product_price }}元</span
        >
      </p>
    </el-card>
    <el-divider></el-divider>
    <div class="priceDiscount">
      <span>
        满减价 ￥
        <span class="discountPrice">{{
          item.price +
          this.pairProduct.product_price -
          combination.priceReductionRange
        }}</span>
        <span class="del"
          >￥{{ item.price + this.pairProduct.product_price }}</span
        >
      </span>
      <el-button style="margin-left: 10px" type="primary" @click="selectThisCombination"
        >收入购物车当中</el-button
      >
    </div>
  </el-card>
</template>

<script>
import { mapActions } from 'vuex'
import { mapGetters } from 'vuex'
export default {
  name: 'FullminusList',
  computed: {
    ...mapGetters(['getShoppingCart']),
    pairProduct() {
      return this.combinationProductList[0]
    },
    shoppingCart() {
      return this.getShoppingCart
    },
  },
  data() {
    return {
      combinationProductList: [],
      combinationProductId: 0,
      combination: {},
    }
  },
  mounted() {
    this.getPairProduct()
  },
  props: ['item', 'index', 'shutDownDrawer'],
  //关于拼单推荐：
  //1、价格不要超过本商品且类别不要和本商品一致
  //2、价格超过本商品，超出用户预算
  //3、类别和用户挑选的一致，冗余商品用户也不会购买
  //以上3点说给系统管理员听/doge,因为这是他们设置的
  methods: {
    ...mapActions([
      'updateShoppingCart',
      'deleteShoppingCart',
      'checkAll',
      'unshiftShoppingCart',
    ]),

    // 加入购物车
    addShoppingCart(user_id, product_id) {
      //若商品不在购物车，则自动加入购物车
      //若商品已在购物车，提示用户该商品已经存在购物车之中
      //查数据库看商品是否存在购物车之中
      this.$axios
        .post('/api/user/shoppingCart/isExistShoppingCart', {
          user_id,
          product_id,
        })
        .then((res) => {
          //不在购物车，加入
          if (res.data.code === '002') {
            //不在加入购物车
            this.$axios
              .post('/api/user/shoppingCart/addShoppingCart', {
                user_id,
                product_id,
              })
              .then((res) => {
                // 新加入购物车成功
                this.unshiftShoppingCart(res.data.shoppingCartData[0])
                this.notifySucceed(res.data.msg)
              })
          } else {
            //在则提示
            this.$notify({
              title: '该组合已在购物车中',
              message: '真是英雄所见略同，喜欢的话快去购物车买买买吧！',
              type: 'success',
            })
          }
        })
    },

    selectThisCombination() {
      this.$emit('shutDownDrawer')

      //let id_1 = this.item.productID
      //let id_2 = this.pairProduct.product_id

      //遍历购物车，找到对应的商品把它的check设置为true
      // for (let i = 0; i < this.shoppingCart.length; i++) {
      //   if (this.shoppingCart[i].productID === this.item.productID) {
      //     this.checkChange(true, i)
      //   }
      // }
      //这里表示的是用户选择了该拼单推荐，要把对应的右边那个商品加入数据库当中
      //将其加入数据库shoppingCart表即可

      this.addShoppingCart(
        this.$store.getters.getUser.user_id,
        this.pairProduct.product_id
      )

      this.$router.push('/shoppingCart')
      // this.checkChange(true, this.shoppingCart.length)
      // console.log(this.shoppingCart)
      //遍历购物车，找到对应的商品把它的check设置为true
      // for (let i = 0; i < this.shoppingCart.length + 1; i++) {
      //   if (this.shoppingCart[i].productID == this.pairProduct.product_id) {
      //     this.checkChange(true, i)
      //   }
      //   console.log('购物车一行信息', this.shoppingCart[i])
      //   console.log(i)
      // }
      //this.$router.push('/confirmOrder')
    },

    checkChange(val, key) {
      // 更新vuex中购物车商品是否勾选的状态
      this.updateShoppingCart({ key: key, prop: 'check', val: val })
    },

    getPairProduct() {
      //console.log(this.item)
      //获取与当前商品对应的组合商品的ID
      this.$axios
        .post('/api/management/getProductCombination', {
          product_id: this.item.productID,
        })
        .then((res) => {
          if (res.data.code === '001') {
            // 001 为成功
            let category = res.data.category
            //console.log('category', category)
            //这里可能会出现bug
            this.combination = category[0]
            //console.log('combination', this.combination)

            this.combinationProductId = this.combination.vice_product_id
            //console.log('combinationProductId', this.combinationProductId)
            this.$notify({
              title: '成功',
              message: 'success',
              type: 'success',
            })
            //获取和该商品对应的组合商品的详细信息
            this.$axios
              .post('/api/management/getCombinationProductList', {
                product_id: this.combinationProductId,
              })
              .then((res) => {
                if (res.data.code === '001') {
                  // 001 为成功
                  this.combinationProductList = res.data.category
                  //console.log('combinationProductList',this.combinationProductList)
                  this.$notify({
                    title: '成功',
                    message: 'success',
                    type: 'success',
                  })
                  //console.log('pairProduct', this.pairProduct)
                } else {
                  // 提示失败信息
                  this.notifyError(res.data.msg)
                }
              })
          } else {
            // 提示失败信息
            this.notifyError(res.data.msg)
          }
        })
    },
  },
}
</script>

<style scoped>
/* el-card {
  width: 300px;
  height: 180px;
}
el-card el-card {
  width: 160px;
}

el-card el-card img {
  width: 140px;
  height: 160px;
  margin: 0 auto;
} */
.box-card {
  width: auto;
  margin: 10px;
}
.box-card .box-card {
  display: inline-block;
}
.box-card img {
  display: block;
  margin: 0 auto;
}
.box-card span,
p,
h2,
h3 {
  text-align: center;
}

.discountPrice {
  color: red;
  font-size: 24px;
}

.del {
  margin-left: 0.5em;
  color: #b0b0b0;
  font-size: 14px;
  text-decoration: line-through;
}

.priceDiscount {
  text-align: right;
}

.priceDiscount span {
  text-align: right;
}
</style>
