<!--
 * @Description: 登录组件 - 主流电商弹窗风格
 -->
<template>
  <transition name="auth-fade">
    <div class="auth-overlay" v-if="isLogin" @click.self="isLogin = false">
      <div class="auth-modal">
        <!-- 左侧品牌区 -->
        <div class="auth-brand">
          <div class="brand-content">
            <span class="brand-logo">JMall</span>
            <h3>欢迎回来</h3>
            <p>登录后享受更多购物权益</p>
            <ul class="brand-features">
              <li><i class="el-icon-check"></i>专属优惠</li>
              <li><i class="el-icon-check"></i>订单追踪</li>
              <li><i class="el-icon-check"></i>收藏同步</li>
            </ul>
          </div>
        </div>
        <!-- 右侧表单区 -->
        <div class="auth-body">
          <span class="modal-close" @click="isLogin = false">
            <i class="el-icon-close"></i>
          </span>
          <div class="form-area">
            <h2>账号登录</h2>
            <el-form
              :model="LoginUser"
              :rules="rules"
              ref="ruleForm"
              class="login-form"
              @submit.native.prevent="Login"
            >
              <el-form-item prop="name">
                <el-input
                  v-model="LoginUser.name"
                  placeholder="用户名"
                  prefix-icon="el-icon-user"
                  @keyup.enter.native="Login"
                ></el-input>
              </el-form-item>
              <el-form-item prop="pass">
                <el-input
                  v-model="LoginUser.pass"
                  type="password"
                  placeholder="密码"
                  prefix-icon="el-icon-lock"
                  show-password
                  @keyup.enter.native="Login"
                ></el-input>
              </el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                class="submit-btn"
                @click="Login"
              >{{ loading ? '登录中' : '登 录' }}</el-button>
            </el-form>
            <div class="form-footer">
              <span>还没有账号？</span>
              <a href="javascript:;" @click="goRegister">立即注册</a>
            </div>
          </div>
          <p class="form-agreement">
            登录即表示同意
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
  name: 'MyLogin',
  data() {
    let validateName = (rule, value, callback) => {
      if (!value) return callback(new Error('请输入用户名'))
      const r = /^[a-zA-Z][a-zA-Z0-9_]{4,15}$/
      if (r.test(value)) {
        this.$refs.ruleForm.validateField('checkPass')
        return callback()
      }
      return callback(new Error('字母开头，长度5-16，允许字母数字下划线'))
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
    return {
      loading: false,
      LoginUser: { name: '', pass: '' },
      rules: {
        name: [{ validator: validateName, trigger: 'blur' }],
        pass: [{ validator: validatePass, trigger: 'blur' }],
      },
    }
  },
  computed: {
    isLogin: {
      get() { return this.$store.getters.getShowLogin },
      set(val) {
        if (this.$refs.ruleForm) this.$refs.ruleForm.resetFields()
        this.setShowLogin(val)
      },
    },
  },
  methods: {
    ...mapActions(['setUser', 'setShowLogin']),
    goRegister() {
      this.isLogin = false
      this.$parent.register = true
    },
    Login() {
      this.$refs.ruleForm.validate((valid) => {
        if (!valid) return
        this.loading = true
        this.$axios
          .post('/api/users/login', {
            userName: this.LoginUser.name,
            password: this.LoginUser.pass,
          })
          .then((res) => {
            this.loading = false
            if (res.data.code === '001') {
              this.isLogin = false
              localStorage.setItem('user', JSON.stringify(res.data.user))
              this.setUser(res.data.user)
              this.notifySucceed(res.data.msg)
            } else {
              this.$refs.ruleForm.resetFields()
              this.notifyError(res.data.msg)
            }
          })
          .catch(() => { this.loading = false })
      })
    },
  },
}
</script>

<style>
/* ===== 登录/注册共享的全局样式 ===== */
.auth-fade-enter-active, .auth-fade-leave-active {
  transition: opacity 0.25s;
}
.auth-fade-enter-active .auth-modal, .auth-fade-leave-active .auth-modal {
  transition: transform 0.25s;
}
.auth-fade-enter, .auth-fade-leave-to {
  opacity: 0;
}
.auth-fade-enter .auth-modal, .auth-fade-leave-to .auth-modal {
  transform: scale(0.95) translateY(10px);
}

.auth-overlay {
  position: fixed;
  inset: 0;
  z-index: 2500;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
}

.auth-modal {
  display: flex;
  width: 720px;
  min-height: 420px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.2);
  background: var(--bg-white, #fff);
}

/* 左侧品牌区 */
.auth-brand {
  width: 260px;
  flex-shrink: 0;
  background: linear-gradient(135deg, #ff6700 0%, #ff8a3d 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 30px;
}
.brand-content { text-align: center; }
.brand-logo {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -1px;
  display: block;
  margin-bottom: 16px;
}
.brand-content h3 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
}
.brand-content p {
  font-size: 13px;
  opacity: 0.85;
  margin-bottom: 24px;
}
.brand-features {
  text-align: left;
  display: inline-block;
}
.brand-features li {
  font-size: 13px;
  line-height: 2;
  opacity: 0.9;
}
.brand-features li i {
  margin-right: 6px;
  font-size: 12px;
}

/* 右侧表单区 */
.auth-body {
  flex: 1;
  padding: 40px 36px 24px;
  display: flex;
  flex-direction: column;
  position: relative;
}
.modal-close {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  color: var(--text-muted, #bbb);
  font-size: 16px;
  transition: all 0.2s;
}
.modal-close:hover {
  background: var(--bg, #f5f5f5);
  color: var(--text, #666);
}

.form-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.form-area h2 {
  font-size: 22px;
  font-weight: 600;
  color: var(--text, #333);
  margin-bottom: 28px;
}

/* 表单输入框 */
.login-form .el-form-item,
.register-form .el-form-item {
  margin-bottom: 22px;
}
.login-form .el-input__inner,
.register-form .el-input__inner {
  height: 46px;
  line-height: 46px;
  border-radius: 6px;
  font-size: 14px;
  border-color: var(--border, #dcdfe6);
}
.login-form .el-input__inner:focus,
.register-form .el-input__inner:focus {
  border-color: #ff6700;
  box-shadow: 0 0 0 2px rgba(255, 103, 0, 0.08);
}
.login-form .el-input__icon,
.register-form .el-input__icon {
  line-height: 46px;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  height: 46px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 6px;
  background: #ff6700;
  border-color: #ff6700;
  letter-spacing: 4px;
  margin-top: 4px;
}
.submit-btn:hover, .submit-btn:focus {
  background: #e55d00;
  border-color: #e55d00;
}

/* 底部链接 */
.form-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-muted, #999);
}
.form-footer a {
  color: #ff6700;
  font-weight: 500;
  margin-left: 4px;
}
.form-footer a:hover {
  text-decoration: underline;
}

/* 协议 */
.form-agreement {
  text-align: center;
  font-size: 12px;
  color: #ccc;
  margin-top: auto;
  padding-top: 16px;
}
.form-agreement a {
  color: var(--text-muted, #999);
}
.form-agreement a:hover {
  color: #ff6700;
}

/* ===== 深色模式 ===== */
html[data-theme="dark"] .auth-modal {
  background: #2c2c2c;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5);
}
html[data-theme="dark"] .auth-brand {
  background: linear-gradient(135deg, #cc5200 0%, #e06b1a 100%);
}
html[data-theme="dark"] .modal-close { color: #666; }
html[data-theme="dark"] .modal-close:hover { background: #3a3a3a; color: #ccc; }
html[data-theme="dark"] .form-area h2 { color: #e0e0e0; }
html[data-theme="dark"] .form-footer { color: #777; }
html[data-theme="dark"] .form-agreement { color: #555; }
html[data-theme="dark"] .form-agreement a { color: #888; }
html[data-theme="dark"] .login-form .el-input__inner,
html[data-theme="dark"] .register-form .el-input__inner {
  background: #363636;
  border-color: #444;
  color: #e0e0e0;
}
html[data-theme="dark"] .strength-track { background: #444; }
</style>
