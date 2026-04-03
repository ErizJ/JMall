<template>
  <el-container>
    <el-main>
      <div class="el-main-up">
        <el-input placeholder="请输入搜索内容" v-model="search">
          <el-button @click="searchit" slot="append">搜索</el-button>
        </el-input>
      </div>
      <div class="el-main-down">
        <el-table
          center
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
          <el-table-column label="用户ID" width="181">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.user_id }}</span>
            </template>
          </el-table-column>
          <el-table-column label="用户名称" width="300">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.userName }}</span>
            </template>
          </el-table-column>
          <el-table-column label="用户密码" width="300">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.password }}</span>
            </template>
          </el-table-column>
          <el-table-column label="联系方式" width="300">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.userPhoneNumber
              }}</span>
            </template>
          </el-table-column>
          <el-table-column fixed="right" label="操作" width="420">
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
                  >删除</el-button
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

    <UserDetailDialog
      :dialogVisible="dialogVisible"
      :title="detailTitle"
      :operateType="operateType"
      :dialogInfo="dialogInfo"
      @close="closeDetailDialog"
    ></UserDetailDialog>
  </el-container>
</template>

<script>
import UserDetailDialog from './UserDetailDialog.vue'
export default {
  components: { UserDetailDialog },
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
      pageSize: 6,

      //对话框相关
      dialogVisible: false, //弹窗显隐
      detailTitle: '', //弹窗标题
      operateType: 'show', //弹窗操作类型 edit-编辑 show-预览 delete-删除
      dialogInfo: {
        user_id: '',
        userName: '',
        password: '',
        userPhoneNumber: '',
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
    tableData() {
      return this.totalTableData.slice(
        (this.currentPage - 1) * this.pageSize,
        this.currentPage * this.pageSize
      )
    },
  },
  methods: {
    searchit() {
      this.getTableList()
      this.currentPage = 1
    },

    handleClick(type, row) {
      console.log(row)
      //获取当前行的信息
      let id = row.user_id
      this.$axios
        .post('/api/users/getDetails', {
          user_id: id,
        })
        .then((res) => {
          console.log('11', res.data)
          this.dialogInfo = res.data.category[0]
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
          this.detailTitle = `用户详情信息`

          break
        case 'edit':
          this.operateType = type
          //把带表单的对话框的显示visible设置为true
          this.dialogVisible = true
          this.detailTitle = `修改用户信息`
          break
        case 'delete':
          //弹出一个警告框，让用户确认是否要删除
          this.$confirm('是否删除该用户?', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          })
            .then(() => {
              //调数据库删除商品信息
              this.deleteUser(id)
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

    getTableList() {
      this.loading = true

      if (this.search == '') {
        this.$axios
          .post('/api/management/getAllUsers')
          .then((res) => {
            this.totalTableData = res.data.category
            this.$message({
              message: '查询成功！',
              type: 'success',
            })
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      } else {
        //搜索商品类名
        this.$axios
          .post('/api/users/getUserByName', {
            userName: this.search,
          })
          .then((res) => {
            this.totalTableData = res.data.category
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      }
      this.loading = false
    },

    handleCurrentChange(val) {
      this.getTableList()
      this.currentPage = val
      this.tableData = this.totalTableData.slice(
        (this.currentPage - 1) * this.pageSize,
        this.currentPage * this.pageSize
      )
    },

    //删除某个商品信息
    deleteUser(id) {
      this.$axios
        .post('/api/users/deleteUserById', {
          user_id: id,
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
