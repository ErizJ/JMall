<template>
  <div class="manage-page">
    <div class="manage-header">
      <h2>商品管理</h2>
      <div class="header-actions">
        <el-input placeholder="按分类名搜索" v-model="search" size="small" style="width:220px" clearable @clear="searchit" @keyup.enter.native="searchit">
          <el-button slot="append" icon="el-icon-search" @click="searchit"></el-button>
        </el-input>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="openUpload">上架商品</el-button>
      </div>
    </div>

    <el-card shadow="never" class="manage-card">
      <el-table :data="tableData" border fit size="small" v-loading="loading" empty-text="暂无商品"
        :header-cell-style="{ background:'#fafafa', color:'#333', fontWeight:600 }">
        <el-table-column prop="product_id" label="ID" width="60" align="center"></el-table-column>
        <el-table-column label="商品" min-width="260">
          <template slot-scope="{ row }">
            <div class="cell-product">
              <img :src="$target + row.product_picture" class="thumb" />
              <div><p class="name">{{ row.product_name }}</p><p class="sub">{{ row.product_title }}</p></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="130" align="center">
          <template slot-scope="{ row }">
            <span class="c-price">¥{{ row.product_selling_price }}</span>
            <span class="c-old" v-if="row.product_price !== row.product_selling_price">¥{{ row.product_price }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_num" label="库存" width="70" align="center" sortable></el-table-column>
        <el-table-column label="促销" width="65" align="center">
          <template slot-scope="{ row }">
            <el-tag size="mini" :type="row.product_isPromotion ? 'danger' : 'info'">{{ row.product_isPromotion ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="product_hot" label="热度" width="65" align="center" sortable></el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="mini" @click="handleClick('show', row)">查看</el-button>
            <el-button type="text" size="mini" @click="handleClick('edit', row)">编辑</el-button>
            <el-button type="text" size="mini" style="color:#f56c6c" @click="handleClick('delete', row)">下架</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagi"><el-pagination background layout="total,prev,pager,next" :total="totalTableData.length" :page-size="pageSize" :current-page.sync="currentPage"></el-pagination></div>
    </el-card>

    <!-- 查看/编辑弹框 -->
    <GoodDetailDialog :dialogVisible="dialogVisible" :title="detailTitle" :operateType="operateType" :dialogInfo="dialogInfo" @close="dialogVisible=false"></GoodDetailDialog>

    <!-- 上架商品弹框 -->
    <el-dialog title="上架新商品" :visible.sync="uploadVisible" width="560px" @close="resetUploadForm">
      <el-form :model="uploadForm" label-width="90px" size="small">
        <el-form-item label="商品名称"><el-input v-model="uploadForm.product_name" placeholder="请输入"></el-input></el-form-item>
        <el-form-item label="分类ID"><el-input v-model="uploadForm.category_id" placeholder="1=手机 2=电视 ..."></el-input></el-form-item>
        <el-form-item label="商品标题"><el-input v-model="uploadForm.product_title" placeholder="简短描述"></el-input></el-form-item>
        <el-form-item label="详细介绍"><el-input type="textarea" :rows="2" v-model="uploadForm.product_intro"></el-input></el-form-item>
        <el-form-item label="图片地址"><el-input v-model="uploadForm.product_picture" placeholder="URL或路径"></el-input></el-form-item>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="原价"><el-input-number v-model="uploadForm.product_price" :min="0" :precision="2" style="width:100%"></el-input-number></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="售价"><el-input-number v-model="uploadForm.product_selling_price" :min="0" :precision="2" style="width:100%"></el-input-number></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="库存"><el-input-number v-model="uploadForm.product_num" :min="1" style="width:100%"></el-input-number></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="促销"><el-switch v-model="uploadForm.product_isPromotion"></el-switch></el-form-item></el-col>
        </el-row>
      </el-form>
      <span slot="footer">
        <el-button size="small" @click="uploadVisible=false">取消</el-button>
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="submitUpload">确认上架</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import GoodDetailDialog from './GoodDetailDialog.vue'
export default {
  components: { GoodDetailDialog },
  mounted() { this.getTableList() },
  data() {
    return {
      search: '', loading: false,
      totalTableData: [], currentPage: 1, pageSize: 8,
      dialogVisible: false, detailTitle: '', operateType: 'show', dialogInfo: {},
      uploadVisible: false,
      uploadForm: { product_name:'', category_id:'', product_title:'', product_intro:'', product_picture:'', product_price:0, product_selling_price:0, product_num:1, product_isPromotion:false },
    }
  },
  watch: { dialogVisible(v) { if(!v) this.getTableList() } },
  computed: {
    tableData() { const s=(this.currentPage-1)*this.pageSize; return this.totalTableData.slice(s,s+this.pageSize) },
  },
  methods: {
    searchit() { this.currentPage=1; this.getTableList() },
    getTableList() {
      this.loading=true
      const api=this.search?'/api/management/getProductsByCategoryName':'/api/management/getAllProducts'
      const data=this.search?{category_name:this.search}:{}
      this.$axios.post(api,data).then(r=>{this.totalTableData=r.data.category||[]}).catch(()=>{}).finally(()=>{this.loading=false})
    },
    handleClick(type,row) {
      if(type==='delete'){
        this.$confirm('确定下架该商品？','提示',{type:'warning'}).then(()=>{
          this.$axios.post('/api/product/deleteProductById',{productID:row.product_id}).then(()=>{this.$message.success('已下架');this.getTableList()})
        }).catch(()=>{})
        return
      }
      this.$axios.post('/api/product/getDetails',{productID:row.product_id}).then(r=>{
        this.dialogInfo=r.data.Product[0]; this.operateType=type; this.detailTitle=type==='show'?'商品详情':'编辑商品'; this.dialogVisible=true
      })
    },
    openUpload() { this.uploadVisible=true },
    resetUploadForm() {
      this.uploadForm={product_name:'',category_id:'',product_title:'',product_intro:'',product_picture:'',product_price:0,product_selling_price:0,product_num:1,product_isPromotion:false}
    },
    submitUpload() {
      this.$axios.post('/api/management/addProduct',{productInfo:this.uploadForm}).then(r=>{
        if(r.data.code==='001'){this.$message.success('上架成功');this.uploadVisible=false;this.getTableList()}
        else this.$message.error(r.data.msg||'上架失败')
      })
    },
  },
}
</script>

<style scoped>
.manage-page {}
.manage-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
.manage-header h2 { font-size:18px; font-weight:600; color:var(--text, #333); }
.header-actions { display:flex; gap:12px; align-items:center; }
.manage-card { border-radius:8px; }
.cell-product { display:flex; align-items:center; gap:10px; padding:4px 0; }
.thumb { width:44px; height:44px; border-radius:6px; object-fit:contain; background:var(--bg, #f9f9f9); border:1px solid var(--border, #f0f0f0); flex-shrink:0; }
.name { font-size:13px; color:var(--text, #333); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.sub { font-size:12px; color:var(--text-muted, #999); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; margin-top:2px; }
.c-price { color:#ff6700; font-weight:600; font-size:13px; }
.c-old { color:#bbb; text-decoration:line-through; font-size:12px; margin-left:4px; }
.pagi { display:flex; justify-content:flex-end; padding:12px 16px; }
</style>
