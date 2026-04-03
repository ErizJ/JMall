<template>
  <el-container>
    <el-main>
      <div class="el-main-up">
        <el-input placeholder="输入商品类别搜索" v-model="search">
          <el-button @click="searchit" slot="append">搜索</el-button>
        </el-input>
      </div>
      <div class="el-main-down">
        <el-table
          :data="tableData"
          show-header
          border
          fit
          height="569px"
          :row-style="{ height: '160px' }"
          :cell-style="{ padding: '0' }"
          v-loading="loading"
          element-loading-spinner="el-icon-loading"
          empty-text="当前暂无数据噢~"
        >
          <el-table-column
            fixed
            prop="product_name"
            label="商品名称"
            width="120"
          ></el-table-column>
          <el-table-column
            prop="product_picture"
            label="商品图片"
            fixed
            width="120"
          >
            <template slot-scope="scope">
              <el-popover placement="top-start" title="" trigger="hover">
                <img
                  :src="$target + scope.row.product_picture"
                  alt=""
                  style="width: 180px; height: 180px"
                />
                <img
                  slot="reference"
                  :src="$target + scope.row.product_picture"
                  style="width: 80px; height: 80px"
                />
              </el-popover>
            </template>
          </el-table-column>
          <el-table-column
            prop="product_id"
            label="商品ID"
            width="150"
          ></el-table-column>
          <el-table-column
            prop="category_id"
            label="商品类别ID"
            width="120"
          ></el-table-column>
          <el-table-column
            prop="product_title"
            label="商品标题"
            width="120"
          ></el-table-column>
          <el-table-column
            prop="product_intro"
            label="商品具体信息"
            width="300"
          ></el-table-column>
          <el-table-column
            prop="product_price"
            label="商品原价"
            width="120"
            sortable
          ></el-table-column>
          <el-table-column
            prop="product_selling_price"
            label="商品现价"
            width="120"
            sortable
          ></el-table-column>
          <el-table-column
            prop="product_num"
            label="商品数量"
            width="120"
            sortable
          ></el-table-column>
          <el-table-column
            prop="product_isPromotion"
            label="是否促销商品"
            width="120"
          ></el-table-column>
          <el-table-column
            prop="product_hot"
            label="商品热度"
            width="120"
            sortable
          ></el-table-column>
          <el-table-column fixed="right" label="操作" width="220">
            <template slot-scope="scope">
              <el-row>
                <el-button
                  type="primary"
                  size="small"
                  @click="handleClick('show', scope.row)"
                  >查看</el-button
                >
                <el-button
                  type="success"
                  size="small"
                  @click="handleClick('edit', scope.row)"
                  >编辑</el-button
                >

                <el-button
                  type="danger"
                  size="small"
                  @click="handleClick('delete', scope.row)"
                  >下架</el-button
                >
              </el-row>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          layout="total,prev, pager, next"
          :total="totalTableData.length"
          :page-size="pageSize"
          @current-change="handleCurrentChange"
          :current-page="currentPage"
          style="padding-top: 10px; float: right"
        ></el-pagination>
      </div>
    </el-main>

    <GoodDetailDialog
      :dialogVisible="dialogVisible"
      :title="detailTitle"
      :operateType="operateType"
      :dialogInfo="dialogInfo"
      @close="closeDetailDialog"
    ></GoodDetailDialog>
  </el-container>
</template>

<script>
import GoodDetailDialog from './GoodDetailDialog.vue'
export default {
  components: { GoodDetailDialog },
  mounted() {
    this.getTableList()
    this.currentPage = 1
    this.tableData = this.totalTableData.slice(
      (this.currentPage - 1) * this.pageSize,
      this.currentPage * this.pageSize
    )
  },
  data() {
    return {
      //搜索框关键字
      search: '',
      //查询列表数据等待
      loading: false,
      //分页相关
      totalTableData: [],
      currentPage: 1,
      pageSize: 3,
      //对话框相关
      dialogVisible: false, //弹窗显隐
      detailTitle: '', //弹窗标题
      operateType: 'show', //弹窗操作类型 edit-编辑 show-预览
      dialogInfo: {
        product_id: '',
        product_name: '',
        category_id: '',
        product_title: '',
        product_picture: '',
        product_intro: '',
        product_price: '',
        product_selling_price: '',
        product_num: '',
        product_isPromotion: '',
        product_hot: '',
      }, //数据参数
    }
  },
  watch: {
    deep: true,
    dialogVisible(newVal) {
      if (newVal === false) {
        this.getTableList()
      }
    },
  },
  computed: {
    tableData() {
      return this.totalTableData.slice(
        (this.currentPage - 1) * this.pageSize,
        this.currentPage * this.pageSize
      )
    },
  },
  methods: {
    //搜索框点击事件
    searchit() {
      this.getTableList()
      this.currentPage = 1
    },

    //获取所有商品信息数据，用于展示在表格中
    getTableList() {
      this.loading = true
      if (this.search == '') {
        this.$axios
          .post('/api/management/getAllProducts')
          .then((res) => {
            this.totalTableData = res.data.category
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      } else {
        //搜索商品类名
        this.$axios
          .post('/api/management/getProductsByCategoryName', {
            category_name: this.search,
          })
          .then((res) => {
            this.totalTableData = res.data.category
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      }

      this.loading = false
      this.$message({
        message: '查询成功！',
        type: 'success',
      })
    },

    //分页-页面切换
    handleCurrentChange(val) {
      this.getTableList()
      this.currentPage = val
      this.tableData = this.totalTableData.slice(
        (this.currentPage - 1) * this.pageSize,
        this.currentPage * this.pageSize
      )
    },

    //表格每行按钮的点击事件处理
    handleClick(type, row) {
      console.log(row)
      //获取当前行的信息
      let id = row.product_id
      this.$axios
        .post('/api/product/getDetails', {
          productID: id,
        })
        .then((res) => {
          this.dialogInfo = res.data.Product[0]
          this.$message({
            message: '查询成功',
            type: 'success',
          })
        })
        .catch((err) => {
          return Promise.reject(err)
        })

      switch (type) {
        case 'show':
          this.operateType = type
          //把带表单的对话框的显示visible设置为true
          this.dialogVisible = true
          this.detailTitle = `商品详情信息`

          break
        case 'edit':
          this.operateType = type
          //把带表单的对话框的显示visible设置为true
          this.dialogVisible = true
          this.detailTitle = `修改商品信息`

          break
        case 'delete':
          //弹出一个警告框，让用户确认是否要删除
          this.$confirm('是否删除该商品?', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          })
            .then(() => {
              //调数据库删除商品信息
              this.deleteProduct(id)
              this.getTableList()
            })
            .catch(() => {
              this.$message({
                type: 'info',
                message: '已取消删除',
              })
            })
          break
      }
    },

    //删除某个商品信息
    deleteProduct(id) {
      this.$axios
        .post('/api/product/deleteProductById', {
          productID: id,
        })
        .then((res) => {
          if (res.code === '001') {
            this.$message({
              type: 'success',
              message: res.msg,
            })
          }
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },

    //关闭对话框
    closeDetailDialog() {
      this.dialogVisible = false
    },
  },
}
</script>

<style>
.el-main-up {
  width: 400px;
  padding-top: 0px;
  padding-bottom: 10px;
}
/* 面包屑CSS */
.el-tabs--card .el-tabs__header {
  border-bottom: none;
}
.breadcrumb {
  height: 40px;
  background-color: white;
}
.breadcrumb .el-breadcrumb {
  width: 100%;
  line-height: 40px;
  font-size: 18px;
  margin: 0 auto;
}
/* 面包屑CSS END */
</style>
