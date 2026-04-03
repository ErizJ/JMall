<template>
  <el-dialog
    :visible="dialogVisible"
    :title="title"
    width="640px"
    top="5vh"
    @close="handleClose"
  >
    <!-- 查看模式 -->
    <div v-if="operateType === 'show'" class="detail-view">
      <!-- 商品头部 -->
      <div class="detail-header">
        <img
          v-if="dialogInfo.product_picture"
          :src="$target + dialogInfo.product_picture"
          class="detail-cover"
        />
        <div class="detail-head-info">
          <h3>{{ dialogInfo.product_name }}</h3>
          <p class="detail-title">{{ dialogInfo.product_title }}</p>
          <div class="detail-price-row">
            <span class="detail-selling">¥{{ dialogInfo.product_selling_price }}</span>
            <span
              v-if="dialogInfo.product_price !== dialogInfo.product_selling_price"
              class="detail-original"
            >¥{{ dialogInfo.product_price }}</span>
          </div>
        </div>
      </div>

      <el-divider></el-divider>

      <!-- 商品信息 -->
      <div class="detail-rows">
        <div class="detail-row">
          <span class="label">商品ID</span>
          <span class="value mono">{{ dialogInfo.product_id }}</span>
        </div>
        <div class="detail-row">
          <span class="label">类别ID</span>
          <span class="value">{{ dialogInfo.category_id }}</span>
        </div>
        <div class="detail-row">
          <span class="label">库存数量</span>
          <span class="value">{{ dialogInfo.product_num }}</span>
        </div>
        <div class="detail-row">
          <span class="label">是否促销</span>
          <span class="value">
            <el-tag size="mini" :type="dialogInfo.product_isPromotion ? 'danger' : 'info'">
              {{ dialogInfo.product_isPromotion ? '促销中' : '否' }}
            </el-tag>
          </span>
        </div>
        <div class="detail-row">
          <span class="label">商品热度</span>
          <span class="value">{{ dialogInfo.product_hot }}</span>
        </div>
        <div class="detail-row" v-if="dialogInfo.product_intro">
          <span class="label">商品介绍</span>
          <span class="value intro">{{ dialogInfo.product_intro }}</span>
        </div>
      </div>
    </div>

    <!-- 编辑模式 -->
    <el-form
      v-if="operateType === 'edit'"
      ref="editForm"
      :model="dialogInfo"
      :rules="rules"
      label-width="90px"
      size="small"
    >
      <el-form-item label="商品ID">
        <span class="mono">{{ dialogInfo.product_id }}</span>
      </el-form-item>
      <el-form-item label="商品名称" prop="notNull">
        <el-input v-model="dialogInfo.product_name" placeholder="请输入商品名称"></el-input>
      </el-form-item>
      <el-form-item label="类别ID" prop="notNull">
        <el-input v-model="dialogInfo.category_id" placeholder="分类ID"></el-input>
      </el-form-item>
      <el-form-item label="商品标题" prop="notNull">
        <el-input v-model="dialogInfo.product_title" placeholder="简短描述"></el-input>
      </el-form-item>
      <el-form-item label="商品介绍" prop="notNull">
        <el-input type="textarea" :rows="3" v-model="dialogInfo.product_intro" placeholder="详细介绍"></el-input>
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="原价" prop="notNull">
            <el-input-number v-model="dialogInfo.product_price" :min="0" :precision="2" style="width:100%"></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="售价" prop="notNull">
            <el-input-number v-model="dialogInfo.product_selling_price" :min="0" :precision="2" style="width:100%"></el-input-number>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="库存" prop="notNull">
            <el-input-number v-model="dialogInfo.product_num" :min="0" style="width:100%"></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="热度">
            <el-input-number v-model="dialogInfo.product_hot" :min="0" style="width:100%"></el-input-number>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="促销">
        <el-switch v-model="dialogInfo.product_isPromotion" active-text="是" inactive-text="否"></el-switch>
      </el-form-item>
    </el-form>

    <span slot="footer">
      <el-button size="small" @click="handleClose">{{ operateType === 'show' ? '关闭' : '取消' }}</el-button>
      <el-button
        v-if="operateType === 'edit'"
        type="primary"
        size="small"
        :loading="ruleFormSubmitting"
        @click="handleSave"
      >保存</el-button>
    </span>
  </el-dialog>
</template>

<script>
export default {
  name: 'GoodDetailDialog',
  props: ['dialogInfo', 'operateType', 'title', 'dialogVisible'],
  data() {
    return {
      ruleFormSubmitting: false,
      rules: {
        notNull: [{ required: true, message: '此项不能为空', trigger: 'blur' }],
      },
    }
  },
  methods: {
    handleClose() {
      this.$emit('close')
    },
    handleSave() {
      this.ruleFormSubmitting = true
      this.$axios
        .post('/api/product/updateProduct', { dialogInfo: this.dialogInfo })
        .then((res) => {
          this.$message.success(res.msg || '保存成功')
          this.ruleFormSubmitting = false
          this.$emit('close')
        })
        .catch((error) => {
          console.log(error)
          this.ruleFormSubmitting = false
        })
    },
  },
}
</script>

<style scoped>
/* 查看模式 */
.detail-view {}
.detail-header {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}
.detail-cover {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  object-fit: contain;
  background: var(--bg, #f9f9f9);
  border: 1px solid var(--border, #f0f0f0);
  flex-shrink: 0;
}
.detail-head-info { flex: 1; min-width: 0; }
.detail-head-info h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text, #333);
  margin-bottom: 4px;
}
.detail-title {
  font-size: 13px;
  color: var(--text-muted, #999);
  margin-bottom: 12px;
}
.detail-price-row { display: flex; align-items: baseline; gap: 10px; }
.detail-selling {
  font-size: 24px;
  font-weight: 700;
  color: var(--primary, #ff6700);
}
.detail-original {
  font-size: 14px;
  color: #bbb;
  text-decoration: line-through;
}

.detail-rows {}
.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  font-size: 14px;
  border-bottom: 1px solid #fafafa;
}
.detail-row .label { color: var(--text-muted, #999); flex-shrink: 0; }
.detail-row .value { color: var(--text, #333); text-align: right; }
.detail-row .value.intro { text-align: right; max-width: 400px; line-height: 1.6; }
.mono { font-family: monospace; }
</style>
