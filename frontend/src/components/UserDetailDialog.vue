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
      <el-form-item label="用户ID">{{ dialogInfo.user_id }}</el-form-item>

      <el-form-item label="用户名称" v-if="operateType === 'show'">{{
        dialogInfo.userName
      }}</el-form-item>

      <el-form-item
        label="用户名称"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.userName"
          autocomplete="on"
        ></el-input>
      </el-form-item>

      <el-form-item label="用户密码" v-if="operateType === 'show'">{{
        dialogInfo.password
      }}</el-form-item>
      <el-form-item
        label="用户密码"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input v-model="dialogInfo.password" autocomplete="on"></el-input>
      </el-form-item>

      <el-form-item label="联系方式" v-if="operateType === 'show'">{{
        dialogInfo.userPhoneNumber
      }}</el-form-item>
      <el-form-item
        label="联系方式"
        v-if="operateType === 'edit'"
        prop="notNull"
      >
        <el-input
          v-model="dialogInfo.userPhoneNumber"
          autocomplete="on"
        ></el-input>
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
            message: '用户信息详情不能为空！',
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
        .post('/api/users/updateUser', {
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
