<template>
  <div class="manage-page">
    <div class="manage-header">
      <h2>促销管理</h2>
      <el-button type="primary" size="small" icon="el-icon-plus" @click="showAdd=true">新增组合</el-button>
    </div>

    <!-- 组合卡片列表 -->
    <div class="combo-grid" v-if="totalTableData.length > 0">
      <div class="combo-card" v-for="item in tableData" :key="item.id">
        <div class="combo-badge">满减</div>
        <div class="combo-body">
          <div class="combo-products">
            <div class="combo-item">
              <i class="el-icon-s-goods"></i>
              <span>主商品 #{{ item.main_product_id }}</span>
            </div>
            <div class="combo-plus">+</div>
            <div class="combo-item">
              <i class="el-icon-s-goods"></i>
              <span>搭配 #{{ item.vice_product_id }}</span>
            </div>
          </div>
          <div class="combo-rule">
            <div class="rule-tag threshold">满 ¥{{ item.amountThreshold }}</div>
            <div class="rule-tag reduction">减 ¥{{ item.priceReductionRange }}</div>
          </div>
        </div>
        <div class="combo-footer">
          <span class="combo-id">ID: {{ item.id }}</span>
          <el-button type="text" size="mini" style="color:#f56c6c" @click="handleDelete(item)">
            <i class="el-icon-delete"></i> 删除
          </el-button>
        </div>
      </div>
    </div>
    <el-card v-else shadow="never" class="empty-card">
      <div class="empty-state">
        <i class="el-icon-s-ticket"></i>
        <p>暂无促销组合</p>
        <el-button type="primary" size="small" @click="showAdd=true">创建第一个</el-button>
      </div>
    </el-card>

    <div class="pagi" v-if="totalTableData.length > pageSize">
      <el-pagination background layout="total,prev,pager,next" :total="totalTableData.length" :page-size="pageSize" :current-page.sync="currentPage"></el-pagination>
    </div>

    <!-- 新增弹框 -->
    <el-dialog title="新增满减组合" :visible.sync="showAdd" width="480px" @close="resetForm">
      <el-form :model="form" label-width="100px" size="small">
        <el-form-item label="主商品ID">
          <el-input v-model="form.main_product_id" placeholder="输入商品ID"></el-input>
        </el-form-item>
        <el-form-item label="搭配商品ID">
          <el-input v-model="form.vice_product_id" placeholder="输入商品ID"></el-input>
        </el-form-item>
        <el-form-item label="满减门槛(元)">
          <el-input v-model="form.amountThreshold" placeholder="如 200"></el-input>
        </el-form-item>
        <el-form-item label="减免金额(元)">
          <el-input v-model="form.priceReductionRange" placeholder="如 20"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button size="small" @click="showAdd=false">取消</el-button>
        <el-button type="primary" size="small" icon="el-icon-check" @click="onSubmit">确认添加</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  created() { this.getTableList() },
  data() {
    return {
      totalTableData:[], currentPage:1, pageSize:12,
      showAdd:false,
      form:{ main_product_id:'', vice_product_id:'', amountThreshold:'', priceReductionRange:'' },
    }
  },
  computed: {
    tableData() { const s=(this.currentPage-1)*this.pageSize; return this.totalTableData.slice(s,s+this.pageSize) },
  },
  methods: {
    resetForm() { this.form={ main_product_id:'', vice_product_id:'', amountThreshold:'', priceReductionRange:'' } },
    handleDelete(row) {
      this.$confirm('确定删除该组合？','提示',{type:'warning'}).then(()=>{
        this.$axios.post('/api/management/deleteProductCombinationById',{id:row.id}).then(r=>{
          if(r.data.code==='001') this.$message.success('已删除')
          this.getTableList()
        })
      }).catch(()=>{})
    },
    onSubmit() {
      this.$axios.post('/api/management/addProductCombination',{productCombination:this.form}).then(r=>{
        if(r.data.code==='001'){this.$message.success('添加成功');this.showAdd=false}
        this.getTableList()
      })
    },
    getTableList() {
      this.$axios.post('/api/management/getAllDiscounts').then(r=>{this.totalTableData=r.data.category||[]})
    },
  },
}
</script>

<style scoped>
.manage-page {}
.manage-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
.manage-header h2 { font-size:18px; font-weight:600; color:#333; }

/* 卡片网格 */
.combo-grid {
  display:grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap:16px;
}
.combo-card {
  background:#fff;
  border-radius:10px;
  border:1px solid #f0f0f0;
  overflow:hidden;
  transition: box-shadow 0.2s;
  position:relative;
}
.combo-card:hover { box-shadow:0 4px 16px rgba(0,0,0,0.08); }
.combo-badge {
  position:absolute;
  top:12px;
  right:-24px;
  background:#ff6700;
  color:#fff;
  font-size:11px;
  padding:2px 28px;
  transform:rotate(45deg);
  font-weight:600;
}
.combo-body { padding:20px 20px 12px; }
.combo-products {
  display:flex;
  align-items:center;
  gap:12px;
  margin-bottom:16px;
}
.combo-item {
  flex:1;
  background:#f9f9f9;
  border-radius:8px;
  padding:12px;
  text-align:center;
  font-size:13px;
  color:#555;
}
.combo-item i { display:block; font-size:24px; color:#999; margin-bottom:4px; }
.combo-plus { font-size:20px; color:#ccc; font-weight:300; flex-shrink:0; }
.combo-rule {
  display:flex;
  gap:8px;
  justify-content:center;
}
.rule-tag {
  padding:4px 14px;
  border-radius:14px;
  font-size:13px;
  font-weight:500;
}
.threshold { background:#fff7e6; color:#fa8c16; }
.reduction { background:#fff1f0; color:#f5222d; }
.combo-footer {
  display:flex;
  align-items:center;
  justify-content:space-between;
  padding:10px 20px;
  border-top:1px solid #f5f5f5;
}
.combo-id { font-size:11px; color:#ccc; }

/* 空状态 */
.empty-card { border-radius:8px; }
.empty-state { text-align:center; padding:40px 0; }
.empty-state i { font-size:48px; color:#ddd; }
.empty-state p { color:#999; margin:8px 0 16px; }

.pagi { display:flex; justify-content:flex-end; padding:16px 0; }
</style>
