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
      v-model="dialogInfo"
      :label-position="labelPosition"
      size="mini"
      label-width="80px"
    >
      <el-form-item label="订单号">{{ dialogInfo.order_id }}</el-form-item>

      <el-form-item label="用户ID">{{ dialogInfo.user_id }}</el-form-item>

      <el-form-item label="商品ID">{{ dialogInfo.product_id }}</el-form-item>

      <el-form-item label="购买数量">{{ dialogInfo.product_num }}</el-form-item>

      <el-form-item label="实付款">{{ dialogInfo.product_price }}</el-form-item>

      <el-form-item label="下单时间">{{ dialogInfo.order_time }}</el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="handleClose" type="primary">确定</el-button>
    </span>
  </el-dialog>
</template>
<script>
export default {
  name: 'OrderDetailDialog',
  props: ['dialogInfo', 'title', 'dialogVisible'],
  created() {
    console.log(this.dialogInfo, this.title, this.dialogVisible)
  },
  data() {
    return {
      detailDialogLoading: false,
      ruleFormSubmitting: false,
      labelPosition: 'right',
    }
  },
  methods: {
    /**
     * 关闭弹框
     */
    handleClose() {
      this.$emit('close')
    },
  },
}
</script>
<style></style>
