<!--
 * @Description: 用户注册组件 - 主流电商弹窗风格
 -->
<template>
  <transition name="auth-fade">
    <div class="auth-overlay" v-if="isRegister" @click.self="isRegister = false">
      <div class="auth-modal">
        <!-- 左侧品牌区 -->
        <div class="auth-brand">
          <div class="brand-content">
            <span class="brand-logo">JMall</span>
            <h3>加入我们</h3>
            <p>注册账号，开启购物之旅</p>
            <ul class="brand-features">
              <li><i class="el-icon-check"></i>新人专享优惠</li>
              <li><i class="el-icon-check"></i>海量商品</li>
              <li><i class="el-icon-check"></i>极速配送</li>
            </ul>
          </div>
        </div>
        <!-- 右侧表单区 -->
        <div class="auth-body">
          <span class="modal-close" @click="isRegister = false">
            <i class="el-icon-close"></i>
          </span>
          <div class="form-area">
            <h2>注册新账号</h2>
            <el-form
              :model="RegisterUser"
              :rules="rules"
              ref="ruleForm"
              class="register-form"
              @submit.native.prevent="Register"
            >
              <el-form-item prop="name">
                <el-input
                  v-model="RegisterUser.name"
                  placeholder="用户名（字母开头，5-16位）"
                  prefix-icon="el-icon-user"
                ></el-input>
              </el-form-item>
              <el-form-item prop="pass">
                <el-input
                  v-model="RegisterUser.pass"
                  type="password"
                  placeholder="密码（字母开头，6-18位）"
                  prefix-icon="el-icon-lock"
                  show-password
                ></el-input>
              </el-form-item>
              <!-- 密码强度 -->
              <div class="strength-bar" v-if="RegisterUser.pass">
                <div class="strength-track">
                  <div
                    :class="['strength-fill', strengthClass]"
                    :style="{ width: strengthPercent + '%' }"
                  ></div>
                </div>
                <span :class="['strength-label', strengthClass]">
                  {{ strengthText }}
                </span>
              </div>
              <el-form-item prop="confirmPass">
                <el-input
                  v-model="RegisterUser.confirmPass"
                  type="password"
                  placeholder="确认密码"
                  prefix-icon="el-icon-lock"
                  show-password
                  @keyup.enter.native="Register"
                ></el-input>
              </el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                class="submit-btn"
                @click="Register"
              >{{ loading ? '注册中' : '注 册' }}</el-button>
            </el-form>
            <div class="form-footer">
              <span>已有账号？</span>
              <a href="javascript:;" @click="goLogin">去登录</a>
            </div>
          </div>
          <p class="form-agreement">
            注册即表示同意
            <a href="javascript:;">用户协议</a> 和
            <a href="javascript:;">隐私政策</a>
          </p>
        </div>
      </div>
    </div>
  </transition>
</template>

<script>
import { mapActions } from 'vuex'

export default {
  name: 'MyRegister',
  props: ['register'],
  data() {
    let validateName = (rule, value, callback) => {
      if (!value) return callback(new Error('请输入用户名'))
      const r = /^[a-zA-Z][a-zA-Z0-9_]{4,15}$/
      if (r.test(value)) {
        this.$axios
          .post('/api/users/findUserName', { userName: this.RegisterUser.name })
          .then((res) => {
            if (res.data.code == '001') {
              this.$refs.ruleForm.validateField('checkPass')
              return callback()
            }
            return callback(new Error(res.data.msg))
          })
          .catch((err) => Promise.reject(err))
      } else {
        return callback(new Error('字母开头，长度5-16，允许字母数字下划线'))
      }
    }
    let validatePass = (rule, value, callback) => {
      if (value === '') return callback(new Error('请输入密码'))
      const r = /^[a-zA-Z]\w{5,17}$/
      if (r.test(value)) {
        this.$refs.ruleForm.validateField('checkPass')
        return callback()
      }
      return callback(new Error('字母开头，长度6-18，允许字母数字下划线'))
    }
    let validateConfirmPass = (rule, value, callback) => {
      if (value === '') return callback(new Error('请输入确认密码'))
      if (this.RegisterUser.pass !== '' && value === this.RegisterUser.pass) {
        this.$refs.ruleForm.validateField('checkPass')
        return callback()
      }
      return callback(new Error('两次输入的密码不一致'))
    }
    return {
      loading: false,
      isRegister: false,
      RegisterUser: { name: '', pass: '', confirmPass: '' },
      rules: {
        name: [{ validator: validateName, trigger: 'blur' }],
        pass: [{ validator: validatePass, trigger: 'blur' }],
        confirmPass: [{ validator: validateConfirmPass, trigger: 'blur' }],
      },
    }
  },
  computed: {
    passwordStrength() {
      const p = this.RegisterUser.pass
      if (!p) return 0
      let s = 0
      if (p.length >= 6) s++
      if (p.length >= 10) s++
      if (/[A-Z]/.test(p)) s++
      if (/[0-9]/.test(p)) s++
      if (/[_]/.test(p)) s++
      return Math.min(s, 4)
    },
    strengthPercent() { return this.passwordStrength * 25 },
    strengthClass() {
      return ['', 'weak', 'fair', 'good', 'strong'][this.passwordStrength] || ''
    },
    strengthText() {
      return ['', '弱', '一般', '较强', '强'][this.passwordStrength] || ''
    },
  },
  watch: {
    register(val) { if (val) this.isRegister = val },
    isRegister(val) {
      if (!val) {
        if (this.$refs.ruleForm) this.$refs.ruleForm.resetFields()
        this.$emit('fromChild', val)
      }
    },
  },
  methods: {
    ...mapActions(['setShowLogin']),
    goLogin() {
      this.isRegister = false
      this.$nextTick(() => { this.setShowLogin(true) })
    },
    Register() {
      this.$refs.ruleForm.validate((valid) => {
        if (!valid) return
        this.loading = true
        this.$axios
          .post('/api/users/register', {
            userName: this.RegisterUser.name,
            password: this.RegisterUser.pass,
          })
          .then((res) => {
            this.loading = false
            if (res.data.code === '001') {
              this.isRegister = false
              this.notifySucceed(res.data.msg)
            } else {
              this.notifyError(res.data.msg)
            }
          })
          .catch(() => { this.loading = false })
      })
    },
  },
}
</script>

<style scoped>
/* 密码强度条 */
.strength-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: -12px 0 18px;
}
.strength-track {
  flex: 1;
  height: 4px;
  background: var(--border, #eee);
  border-radius: 2px;
  overflow: hidden;
}
.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s, background 0.3s;
}
.strength-fill.weak { background: #f56c6c; }
.strength-fill.fair { background: #e6a23c; }
.strength-fill.good { background: #409eff; }
.strength-fill.strong { background: #67c23a; }
.strength-label {
  font-size: 12px;
  flex-shrink: 0;
  width: 28px;
}
.strength-label.weak { color: #f56c6c; }
.strength-label.fair { color: #e6a23c; }
.strength-label.good { color: #409eff; }
.strength-label.strong { color: #67c23a; }
</style>
