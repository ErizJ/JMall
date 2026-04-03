<!--
 * @Description: 关于我们页面组件
 -->

<template>
  <el-container>
    <el-header></el-header>
    <el-main>
      <el-row :gutter="20">
  <el-col :span="3"><div class="grid-content bg-purple"></div></el-col>
  <el-col :span="18"><div>
            <img src="../assets/imgs/logo.png" />
          </div>
          <div class="img">

              <img src="../assets/imgs/about.png" />

          </div></el-col>
  <el-col :span="3"><div class="grid-content bg-purple"></div></el-col>
</el-row>
    </el-main>
    <el-footer></el-footer>
  </el-container>
</template>

<script>
export default {}
</script>

<style scoped>
.el-row {
    margin-bottom: 20px;
  }
  .el-col {
    border-radius: 4px;
  }
  .bg-purple-dark {
    background: transparent;
  }
  .bg-purple {
    background: transparent;
  }
  .bg-purple-light {
    background: transparent;
  }
  .grid-content {
    border-radius: 4px;
    min-height: 36px;
  }
  .row-bg {
    padding: 10px 0;
    background-color: #f9fafc;
  }
  

</style>
