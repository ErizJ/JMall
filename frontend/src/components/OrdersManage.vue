<template>
  <el-container>
    <el-main>
      <div class="el-main-up">
        <el-input placeholder="输入用户名搜索对应订单" v-model="search">
          <el-button @click="searchit" slot="append">搜索</el-button>
        </el-input>
      </div>
      <div class="el-main-down">
        <el-table
          :data="tableData"
          show-header
          border
          fit
          highlight-current-row
          height="580px"
          :row-style="{ height: '70px' }"
          :cell-style="{ padding: '1' }"
          element-loading-text="Loading"
        >
          <el-table-column label="订单号" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.order_id }}</span>
            </template>
          </el-table-column>
          <el-table-column label="用户ID" width="180">
            <template slot-scope="scope">
              <el-popover trigger="hover" placement="top">
                <p>ID: {{ scope.row.user_id }}</p>

                <div slot="reference" class="name-wrapper">
                  <el-tag size="medium">{{ scope.row.user_id }}</el-tag>
                </div>
              </el-popover>
            </template>
          </el-table-column>
          <el-table-column label="商品ID" width="180">
            <template slot-scope="scope">
              <el-popover trigger="hover" placement="top">
                <p>ID: {{ scope.row.product_id }}</p>

                <div slot="reference" class="name-wrapper">
                  <el-tag size="medium">{{ scope.row.product_id }}</el-tag>
                </div>
              </el-popover>
            </template>
          </el-table-column>
          <el-table-column label="购买数量" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.product_num }}</span>
            </template>
          </el-table-column>
          <el-table-column label="实付款" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.product_price
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="下单时间" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.order_time }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template slot-scope="scope">
              <el-button
                size="mini"
                type="primary"
                @click="handleClick('show', scope.row)"
                >查看</el-button
              >
              <el-button
                size="mini"
                type="danger"
                @click="handleClick('delete', scope.row)"
                >删除</el-button
              >
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
    <OrderDetailDialog
      :dialogVisible="dialogVisible"
      :title="detailTitle"
      :dialogInfo="dialogInfo"
      @close="closeDetailDialog"
    ></OrderDetailDialog>
  </el-container>
</template>

<script>
import OrderDetailDialog from './OrderDetailDialog.vue'
export default {
  components: { OrderDetailDialog },
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
      pageSize: 7,

      //对话框相关
      dialogVisible: false, //弹窗显隐
      detailTitle: '', //弹窗标题
      dialogInfo: {
        order_id: '',
        user_id: '',
        user_name: '',
        product_id: '',
        product_num: '',
        product_price: '',
        order_time: '',
      }, //数据参数

      //用于根据用户ID确定对应的用户名
      user_name: '',
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
    tableData: function () {
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

    getTableList() {
      this.loading = true
      if (this.search == '') {
        this.$axios
          .post('/api/management/getAllOrders')
          .then((res) => {
            this.totalTableData = res.data.category
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      } else {
        //搜索商品类名
        this.$axios
          .post('/api/management/getOrdersByUserName', {
            user_name: this.search,
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

    handleCurrentChange(val) {
      this.getTableList()
      this.currentPage = val
      this.tableData = this.totalTableData.slice(
        (this.currentPage - 1) * this.pageSize,
        this.currentPage * this.pageSize
      )
    },

    handleClick(type, row) {
      console.log(row, type)
      //获取当前行信息
      let id = row.order_id
      this.$axios
        .post('/api/order/getDetails', {
          order_id: id,
        })
        .then((res) => {
          console.log(res.data)
          this.dialogInfo = res.data.category[0]
        })
        .catch((err) => {
          return Promise.reject(err)
        })

      switch (type) {
        case 'show':
          //把带表单的对话框的显示visible设置为true
          this.dialogVisible = true
          this.detailTitle = `查看订单信息`

          break
        case 'delete':
          //弹出一个警告框，让用户确认是否要删除
          this.$confirm('是否删除该订单?', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          })
            .then(() => {
              //调数据库删除商品信息
              this.deleteOrder(id)
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
    deleteOrder(id) {
      this.$axios
        .post('/api/order/deleteOrderById', {
          order_id: id,
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
