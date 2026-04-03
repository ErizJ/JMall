<template>
  <el-dialog
    center
    v-bind="dialogInfo"
    v-loading="detailDialogLoading"
    :visible="dialogVisible"
    show-close
    :title="title"
    close-on-press-escape
    close-on-click-modal
    @close="handleClose"
  >
    <el-form
      ref="dialogInfo"
      :rules="rules"
      v-model="dialogInfo"
      :label-position="labelPosition"
      size="mini"
      label-width="80px"
    >
      <el-form-item label="商品ID">{{ dialogInfo.product_id }}</el-form-item>

      <el-form-item label="商品名称" v-if="operateType === 'show'">{{
        dialogInfo.product_name
      }}</el-form-item>

      <el-form-item
        label="商品名称"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.product_name"
          autocomplete="on"
        ></el-input>
      </el-form-item>

      <el-form-item label="类别ID" v-if="operateType === 'show'">{{
        dialogInfo.category_id
      }}</el-form-item>
      <el-form-item label="类别ID" v-if="operateType === 'edit'" prop="notNull">
        <el-input v-model="dialogInfo.category_id" autocomplete="on"></el-input>
      </el-form-item>

      <el-form-item label="商品标题" v-if="operateType === 'show'">{{
        dialogInfo.product_title
      }}</el-form-item>
      <el-form-item
        label="商品标题"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.product_title"
          autocomplete="on"
        ></el-input>
      </el-form-item>

      <el-form-item label="商品图片" v-if="operateType === 'show'">{{
        dialogInfo.product_title
      }}</el-form-item>

      <el-form-item label="商品具体信息" v-if="operateType === 'show'">{{
        dialogInfo.product_intro
      }}</el-form-item>
      <el-form-item
        label="商品具体信息"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input type="textarea" v-model="dialogInfo.product_intro"></el-input>
      </el-form-item>

      <el-form-item label="商品原价" v-if="operateType === 'show'">{{
        dialogInfo.product_price
      }}</el-form-item>
      <el-form-item
        label="商品原价"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.product_price"
          autocomplete="on"
        ></el-input>
      </el-form-item>

      <el-form-item label="商品现价" v-if="operateType === 'show'">{{
        dialogInfo.product_selling_price
      }}</el-form-item>
      <el-form-item
        label="商品现价"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.product_selling_price"
          autocomplete="on"
        ></el-input>
      </el-form-item>

      <el-form-item label="商品数量" v-if="operateType === 'show'">{{
        dialogInfo.product_num
      }}</el-form-item>
      <el-form-item
        label="商品数量"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input v-model="dialogInfo.product_num" autocomplete="on"></el-input>
      </el-form-item>

      <el-form-item label="是否促销" v-if="operateType === 'show'">{{
        dialogInfo.product_isPromotion
      }}</el-form-item>
      <el-form-item label="是否促销" v-if="operateType === 'edit'">
        <el-switch v-model="dialogInfo.product_isPromotion"></el-switch>
      </el-form-item>

      <el-form-item label="商品热度" v-if="operateType === 'show'">{{
        dialogInfo.product_hot
      }}</el-form-item>
      <el-form-item label="商品热度" v-if="operateType === 'edit'">
        <el-input v-model="dialogInfo.product_hot" autocomplete="on"></el-input>
      </el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="handleClose">取消</el-button>
      <el-button
        :loading="ruleFormSubmitting"
        type="primary"
        @click="handleSave"
        v-if="operateType === 'edit'"
        >保存</el-button
      >
    </span>
  </el-dialog>
</template>
<script>
export default {
  name: 'ProductDetailDialog',
  props: ['dialogInfo', 'operateType', 'title', 'dialogVisible'],
  created() {
    console.log(
      this.dialogInfo,
      this.operateType,
      this.title,
      this.dialogVisible
    )
  },
  data() {
    return {
      detailDialogLoading: false,
      ruleFormSubmitting: false,
      labelPosition: 'right',
      rules: {
        notNull: [
          {
            required: true,
            message: '商品详情不能为空！',
            trigger: 'blur',
          },
        ],
      },
    }
  },
  methods: {
    /**
     * 关闭弹框
     */
    handleClose() {
      this.$emit('close')
      if (this.operateType === 'edit') {
        this.$notify({
          showClose: true,
          message: '已取消修改',
        })
      }
    },
    /**
     * 保存数据
     * @argument flag true-保存 false-保存并继续
     */
    handleSave() {
      this.detailDialogLoading = true
      this.ruleFormSubmitting = true
      //发请求
      this.$axios
        .post('/api/product/updateProduct', {
          dialogInfo: this.dialogInfo,
        })
        .then((res) => {
          this.$message({
            message: res.msg,
            type: 'success',
          })
          this.ruleFormSubmitting = false
          this.detailDialogLoading = false
        })
        .catch((error) => {
          // 重置表单
          console.log(error)
          this.ruleFormSubmitting = false
          this.detailDialogLoading = false
        })
      this.dialogVisible = false
    },
  },
}
</script>
<style></style>
