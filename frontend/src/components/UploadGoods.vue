<template>
  <div class="manage-page">
    <div class="manage-header">
      <h2>商品上架</h2>
    </div>

    <el-card shadow="never" class="upload-card">
      <el-form
        :model="productInfo"
        ref="productInfo"
        label-width="100px"
        size="small"
        style="max-width: 600px"
      >
        <el-form-item label="商品名称">
          <el-input v-model="productInfo.product_name" placeholder="请输入商品名称"></el-input>
        </el-form-item>
        <el-form-item label="分类ID">
          <el-input v-model="productInfo.category_id" placeholder="如 1=手机, 2=电视"></el-input>
        </el-form-item>
        <el-form-item label="商品标题">
          <el-input v-model="productInfo.product_title" placeholder="简短描述"></el-input>
        </el-form-item>
        <el-form-item label="详细介绍">
          <el-input type="textarea" :rows="3" v-model="productInfo.product_intro" placeholder="商品详细信息"></el-input>
        </el-form-item>
        <el-form-item label="图片地址">
          <el-input v-model="productInfo.product_picture" placeholder="图片URL或路径"></el-input>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="原价(元)">
              <el-input-number v-model="productInfo.product_price" :min="0" :precision="2" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="售价(元)">
              <el-input-number v-model="productInfo.product_selling_price" :min="0" :precision="2" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="库存数量">
              <el-input-number v-model="productInfo.product_num" :min="1" style="width:100%"></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="参与促销">
              <el-switch v-model="productInfo.product_isPromotion" active-text="是" inactive-text="否"></el-switch>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button @click="resetForm">重置</el-button>
          <el-button type="primary" icon="el-icon-upload2" @click="submitForm">确认上架</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
export default {
  data() {
    return {
      productInfo: {
        product_name: '', category_id: '', product_title: '', product_intro: '',
        product_picture: '', product_price: 0, product_selling_price: 0,
        product_num: 1, product_isPromotion: false,
      },
    }
  },
  methods: {
    submitForm() {
      this.$axios.post('/api/management/addProduct', { productInfo: this.productInfo }).then((res) => {
        if (res.data.code === '001') {
          this.$message.success('上架成功')
          this.$router.push('/goodsmanage')
        } else {
          this.$message.error(res.data.msg || '上架失败')
        }
      })
    },
    resetForm() {
      this.productInfo = {
        product_name: '', category_id: '', product_title: '', product_intro: '',
        product_picture: '', product_price: 0, product_selling_price: 0,
        product_num: 1, product_isPromotion: false,
      }
    },
  },
}
</script>

<style scoped>
.manage-page { }
.manage-header { margin-bottom: 16px; }
.manage-header h2 { font-size: 18px; font-weight: 600; color: #333; }
.upload-card { border-radius: 8px; }
</style>
