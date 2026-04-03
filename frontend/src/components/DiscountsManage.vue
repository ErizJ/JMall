<template>
  <el-container>
   
    <el-main>
      <el-form :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item label="选择主商品">
          <el-input
            v-model="formInline.main_product_id"
            placeholder="输入主商品号"
          ></el-input>
        </el-form-item>
        <el-form-item label="选择组合商品">
          <el-input
            v-model="formInline.vice_product_id"
            placeholder="输入组合商品号"
          ></el-input>
        </el-form-item>
        <el-form-item label="达到满减门槛">
          <el-input
            v-model="formInline.amountThreshold"
            placeholder="达到满减门槛"
          ></el-input>
        </el-form-item>
        <el-form-item label="满减金额">
          <el-input
            v-model="formInline.priceReductionRange"
            placeholder="达到满减门槛"
          ></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">设置</el-button>
        </el-form-item>
      </el-form>
      <div class="el-main-down">
        <el-table
          :data="tableData"
          show-header
          border
          fit
          height="560px"
          :row-style="{ height: '70px' }"
          :cell-style="{ padding: '0' }"
          v-loading="loading"
          element-loading-spinner="el-icon-loading"
          empty-text="当前暂无数据噢~"
        >
          <el-table-column label="ID" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column label="已选商品" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.main_product_id
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="拼单推荐商品" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.vice_product_id
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="已达满减金额标准" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.amountThreshold
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="满减力度" width="180">
            <template slot-scope="scope">
              <span style="margin-left: 10px">{{
                scope.row.priceReductionRange
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template slot-scope="scope">
              <el-button
                size="mini"
                type="danger"
                @click="handleDelete(scope.row)"
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
  </el-container>
</template>

<script>
export default {
  created() {
    this.getTableList()
  },
  data() {
    return {
      //搜索框关键字
      search: '',
      formInline: {
        main_product_id: '',
        vice_product_id: '',
        amountThreshold: '',
        priceReductionRange: '',
      },

      //分页相关
      totalTableData: [],
      currentPage: 1,
      pageSize: 7,

      //用于根据用户ID确定对应的用户名
      user_name: '',
    }
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
    handleDelete(row) {
      let id = row.id
      this.$axios
        .post('/api/management/deleteProductCombinationById', {
          id,
        })
        .then((res) => {
          if (res.data.code === '001') {
            console.log(res)
            this.$message({
              message: res.data.msg,
              type: 'success',
            })
          }
          this.getTableList()
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },

    onSubmit() {
      this.$axios
        .post('/api/management/addProductCombination', {
          productCombination: this.formInline,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.$message({
              message: '新增组合成功！',
              type: 'success',
            })
          }
          this.getTableList()
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },

    getTableList() {
      this.$axios
        .post('/api/management/getAllDiscounts')
        .then((res) => {
          this.totalTableData = res.data.category
          console.log('get totalTableData')
          console.log(this.totalTableData)
        })
        .catch((err) => {
          return Promise.reject(err)
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
