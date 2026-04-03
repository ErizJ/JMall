<template>
  <el-row class="tac" style="background-color:rgb(245,245,245)">
    <el-col :span="3" style="height:100%;background-color:white">
      <el-menu
        default-active="active"
        class="el-menu-vertical-demo"
        @open="handleOpen"
        text-color="#909399"
        router
      >
      <el-menu-item index="/" disabled>
          <span slot="title">———系统菜单———</span>
      </el-menu-item>

      
      <el-menu-item index="goodsmanage">
          <i class="el-icon-s-goods"></i>
          <span slot="title">商品详情</span>
        </el-menu-item>
        <el-menu-item index="uploadgoods">
          <i class="el-icon-s-goods"></i>
          <span slot="title">商品上架</span>
        </el-menu-item>
        <el-menu-item index="ordersmanage">
          <i class="el-icon-s-order"></i>
          <span slot="title">订单管理</span>
        </el-menu-item>
        <el-menu-item index="discountsmanage">
          <i class="el-icon-message-solid"></i>
          <span slot="title">促销管理</span>
        </el-menu-item>
        <el-menu-item index="usersmanage">
          <i class="el-icon-s-custom"></i>
          <span slot="title">用户管理</span>
        </el-menu-item>
        <el-menu-item index="about">
          <i class="el-icon-cpu"></i>
          <span slot="title">关于系统</span>
        </el-menu-item>
      </el-menu>
    </el-col>
    <el-col :span="21">
      <router-view></router-view>
    </el-col>
  </el-row>
</template>

<script>
export default {
  data() {
    return {
      active: '2',
    }
  },
  methods: {
    handleOpen(key, keyPath) {
      console.log(key, keyPath)
    },
  },
}
</script>

<style>
.tac {
  padding: 2px;
  height: 900px;
  min-width: 1424px;
}
</style>