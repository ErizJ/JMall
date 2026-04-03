<template>
  <div class="manage-page">
    <div class="manage-header">
      <h2>用户管理 <span class="count-badge">{{ totalTableData.length }} 人</span></h2>
      <el-input placeholder="按用户名搜索" v-model="search" size="small" style="width:260px" clearable @clear="searchit" @keyup.enter.native="searchit">
        <el-button slot="append" icon="el-icon-search" @click="searchit"></el-button>
      </el-input>
    </div>

    <el-card shadow="never" class="manage-card">
      <el-table :data="tableData" fit size="small" v-loading="loading" empty-text="暂无用户"
        :header-cell-style="{ background:'#fafafa', color:'#333', fontWeight:600 }">
        <el-table-column prop="user_id" label="ID" width="70" align="center"></el-table-column>
        <el-table-column label="用户" min-width="220">
          <template slot-scope="{ row }">
            <div class="cell-user">
              <el-avatar size="small" :src="avatarUrl(row.user_id)"></el-avatar>
              <div class="user-meta">
                <p class="user-name-text">{{ row.userName || row.user_name }}</p>
                <p class="user-id-text">UID: {{ row.user_id }}</p>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="手机号" width="150" align="center">
          <template slot-scope="{ row }">
            <span>{{ row.userPhoneNumber || row.user_phone_number || '未绑定' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="90" align="center">
          <template slot-scope="{ row }">
            <el-tag size="mini" :type="row.user_id === 1 ? 'danger' : ''">{{ row.user_id === 1 ? '管理员' : '普通用户' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="mini" @click="handleClick('show', row)">查看</el-button>
            <el-button type="text" size="mini" @click="handleClick('edit', row)">编辑</el-button>
            <el-button type="text" size="mini" style="color:#f56c6c" @click="handleClick('delete', row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagi"><el-pagination background layout="total,prev,pager,next" :total="totalTableData.length" :page-size="pageSize" :current-page.sync="currentPage"></el-pagination></div>
    </el-card>

    <!-- 用户详情/编辑弹框 -->
    <el-dialog :title="detailTitle" :visible.sync="dialogVisible" width="480px">
      <div class="user-profile" v-if="operateType==='show'">
        <div class="profile-header">
          <el-avatar :size="64" :src="avatarUrl(dialogInfo.user_id)"></el-avatar>
          <div>
            <h3>{{ dialogInfo.userName || dialogInfo.user_name }}</h3>
            <p class="uid">UID: {{ dialogInfo.user_id }}</p>
          </div>
        </div>
        <el-divider></el-divider>
        <div class="profile-info">
          <div class="info-row"><span class="label">手机号</span><span>{{ dialogInfo.userPhoneNumber || dialogInfo.user_phone_number || '未绑定' }}</span></div>
          <div class="info-row"><span class="label">角色</span><el-tag size="mini" :type="dialogInfo.user_id===1?'danger':''">{{ dialogInfo.user_id===1?'管理员':'普通用户' }}</el-tag></div>
        </div>
      </div>
      <el-form v-else :model="dialogInfo" label-width="80px" size="small">
        <el-form-item label="用户名"><el-input v-model="dialogInfo.userName"></el-input></el-form-item>
        <el-form-item label="手机号"><el-input v-model="dialogInfo.userPhoneNumber"></el-input></el-form-item>
      </el-form>
      <span slot="footer">
        <el-button size="small" @click="dialogVisible=false">{{ operateType==='show'?'关闭':'取消' }}</el-button>
        <el-button v-if="operateType==='edit'" type="primary" size="small" @click="saveUser">保存</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  mounted() { this.getTableList() },
  data() {
    return {
      search:'', loading:false,
      totalTableData:[], currentPage:1, pageSize:10,
      dialogVisible:false, detailTitle:'', operateType:'show', dialogInfo:{},
    }
  },
  watch: { dialogVisible(v) { if(!v) this.getTableList() } },
  computed: {
    tableData() { const s=(this.currentPage-1)*this.pageSize; return this.totalTableData.slice(s,s+this.pageSize) },
  },
  methods: {
    avatarUrl(id) { return 'https://api.dicebear.com/7.x/initials/svg?seed=user' + id },
    searchit() { this.currentPage=1; this.getTableList() },
    getTableList() {
      this.loading=true
      const api=this.search?'/api/users/getUserByName':'/api/management/getAllUsers'
      const data=this.search?{userName:this.search}:{}
      this.$axios.post(api,data).then(r=>{this.totalTableData=r.data.category||[]}).catch(()=>{}).finally(()=>{this.loading=false})
    },
    handleClick(type,row) {
      if(type==='delete'){
        this.$confirm('确定删除该用户？','提示',{type:'warning'}).then(()=>{
          this.$axios.post('/api/users/deleteUserById',{user_id:row.user_id}).then(()=>{this.$message.success('已删除');this.getTableList()})
        }).catch(()=>{})
        return
      }
      this.$axios.post('/api/users/getDetails',{user_id:row.user_id}).then(r=>{
        this.dialogInfo=(r.data.category&&r.data.category[0])||r.data||row
        this.operateType=type; this.detailTitle=type==='show'?'用户详情':'编辑用户'; this.dialogVisible=true
      })
    },
    saveUser() {
      this.$axios.post('/api/users/updateUser',{dialogInfo:this.dialogInfo}).then(()=>{
        this.$message.success('保存成功'); this.dialogVisible=false
      })
    },
  },
}
</script>

<style scoped>
.manage-page {}
.manage-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
.manage-header h2 { font-size:18px; font-weight:600; color:#333; }
.count-badge { font-size:12px; font-weight:400; color:#999; background:#f5f5f5; padding:2px 8px; border-radius:10px; margin-left:8px; }
.manage-card { border-radius:8px; }
.cell-user { display:flex; align-items:center; gap:10px; }
.user-meta {}
.user-name-text { font-size:13px; color:#333; font-weight:500; }
.user-id-text { font-size:11px; color:#bbb; margin-top:1px; }
.pagi { display:flex; justify-content:flex-end; padding:12px 16px; }

/* 用户详情弹框 */
.user-profile {}
.profile-header { display:flex; align-items:center; gap:16px; }
.profile-header h3 { font-size:18px; color:#333; }
.profile-header .uid { font-size:12px; color:#999; margin-top:2px; }
.profile-info {}
.info-row { display:flex; justify-content:space-between; padding:8px 0; font-size:14px; color:#333; border-bottom:1px solid #fafafa; }
.info-row .label { color:#999; }
</style>
