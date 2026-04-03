<template>
  <el-row>
    <el-col :span="6"><div class="grid-content"></div></el-col>

    <el-col :span="12">
      <el-link class="title" type="primary" disabled>填写上架商品信息</el-link>

      <el-main class="myfrom">
        <el-form
          center
          :model="productInfo"
          ref="productInfo"
          class="demo-ruleForm"
          label-width="80px"
        >
          <el-form-item label="商品名称">
            <el-input v-model="productInfo.product_name"></el-input>
          </el-form-item>
          <el-form-item label="类别ID">
            <el-input v-model="productInfo.category_id"></el-input>
          </el-form-item>

          <el-form-item label="商品标题">
            <el-input v-model="productInfo.product_title"></el-input>
          </el-form-item>

          <el-form-item label="具体信息">
            <el-input
              type="textarea"
              v-model="productInfo.product_intro"
            ></el-input>
          </el-form-item>

          <el-form-item label="商品图片">
            <el-upload
  action="#"
  list-type="picture-card"
  :auto-upload="false">
    <i slot="default" class="el-icon-plus"></i>
    <div slot="file" slot-scope="{file}">
      <img
        class="el-upload-list__item-thumbnail"
        :src="file.url" alt=""
      >
      <span class="el-upload-list__item-actions">
        <span
          class="el-upload-list__item-preview"
          @click="handlePictureCardPreview(file)"
        >
          <i class="el-icon-zoom-in"></i>
        </span>
        <span
          v-if="!disabled"
          class="el-upload-list__item-delete"
          @click="handleDownload(file)"
        >
          <i class="el-icon-download"></i>
        </span>
        <span
          v-if="!disabled"
          class="el-upload-list__item-delete"
          @click="handleRemove(file)"
        >
          <i class="el-icon-delete"></i>
        </span>
      </span>
    </div>
</el-upload>

          </el-form-item>

          <el-form-item label="商品原价" prop="notNull">
            <el-input-number
              v-model="productInfo.product_price"
              :min="1"
            ></el-input-number>
          </el-form-item>

          <el-form-item label="商品现价">
            <el-input-number
              v-model="productInfo.product_selling_price"
              :min="1"
            ></el-input-number>
          </el-form-item>

          <el-form-item label="商品数量">
            <el-input-number
              v-model="productInfo.product_num"
              :min="1"
            ></el-input-number>
          </el-form-item>

          <el-form-item label="是否促销">
            <el-switch v-model="productInfo.product_isPromotion"></el-switch>
          </el-form-item>

          <el-form-item>
            <el-button @click="resetForm">重置</el-button>
            <el-button type="primary" @click="submitForm">上架</el-button>
          </el-form-item>
        </el-form>
      </el-main>
      <el-footer> </el-footer>
    </el-col>
    <el-col :span="6"><div class="grid-content"></div></el-col>
  </el-row>
</template>
<script>
export default {
  data() {
    return {
      productInfo: {
        product_name: '',
        category_id: '',
        product_title: '',
        product_intro: '',
        product_picture: '',
        product_price: '',
        product_selling_price: '',
        product_num: 1,
        product_isPromotion: 0,
      },
    }
  },
  methods: {
    submitForm() {
      console.log(this.productInfo)
      this.$axios
        .post('/api/management/addProduct', {
          productInfo: this.productInfo,
        })
        .then((res) => {
          if (res.data.code === '001') {
            this.$message({
              message: res.data.msg,
              type: 'success',
            })
            this.$router.push('/goodsmanage')
          } else {
            this.$router.push('/error')
          }
        })
        .catch((err) => {
          return Promise.reject(err)
        })
    },
    resetForm() {
      this.productInfo = {
        product_name: '',
        category_id: '',
        product_title: '',
        product_intro: '',
        product_picture: '',
        product_price: '',
        product_selling_price: '',
        product_num: 1,
        product_isPromotion: false,
      }
    },
  },
}
</script>
<style scoped>
.grid-content {
  width: auto;
  height: 100%;
  min-height: 60px;
  background-color: transparent;
}
.title {
  display: block;
  text-align: center;
  font-size: 24px;
  align-content: center;
  padding-top: 20px;
}
.myfrom{
  height: 100%;
}
</style>
