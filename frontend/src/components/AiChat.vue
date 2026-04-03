<template>
  <div class="ai-chat">
    <!-- 悬浮按钮 -->
    <div class="ai-chat-fab" @click="toggleChat" :class="{ active: visible }">
      <i :class="visible ? 'el-icon-close' : 'el-icon-chat-dot-round'"></i>
    </div>

    <!-- 聊天窗口 -->
    <transition name="chat-slide">
      <div v-if="visible" class="ai-chat-window">
        <div class="chat-header">
          <div class="chat-header-info">
            <i class="el-icon-s-opportunity"></i>
            <span>JMall 智能助手</span>
          </div>
          <i class="el-icon-minus chat-minimize" @click="visible = false"></i>
        </div>

        <div class="chat-body" ref="chatBody">
          <!-- 欢迎消息 -->
          <div class="chat-msg bot" v-if="messages.length === 0">
            <div class="msg-avatar"><i class="el-icon-s-opportunity"></i></div>
            <div class="msg-bubble">
              你好！我是 JMall 智能购物助手 🛒<br/>
              我可以帮你查询商品、了解价格和促销信息。试试问我：<br/>
              <span class="suggestion" @click="sendSuggestion('有什么热门商品推荐？')">🔥 热门商品推荐</span>
              <span class="suggestion" @click="sendSuggestion('现在有什么促销活动？')">🏷️ 促销活动</span>
              <span class="suggestion" @click="sendSuggestion('帮我找一款手机')">📱 搜索商品</span>
            </div>
          </div>

          <div v-for="(msg, idx) in messages" :key="idx" class="chat-msg" :class="msg.role">
            <div class="msg-avatar">
              <i :class="msg.role === 'user' ? 'el-icon-user-solid' : 'el-icon-s-opportunity'"></i>
            </div>
            <div class="msg-bubble" v-html="formatMsg(msg.content)"></div>
          </div>

          <!-- 加载指示器 -->
          <div v-if="loading" class="chat-msg bot">
            <div class="msg-avatar"><i class="el-icon-s-opportunity"></i></div>
            <div class="msg-bubble typing">
              <span v-if="thinkingText">{{ thinkingText }}</span>
              <span v-else class="dot-typing">
                <span></span><span></span><span></span>
              </span>
            </div>
          </div>
        </div>

        <div class="chat-footer">
          <el-input
            v-model="inputMsg"
            placeholder="输入你的问题..."
            size="small"
            @keyup.enter.native="sendMessage"
            :disabled="loading"
          >
            <el-button
              slot="append"
              icon="el-icon-s-promotion"
              @click="sendMessage"
              :disabled="loading || !inputMsg.trim()"
            ></el-button>
          </el-input>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'AiChat',
  data() {
    return {
      visible: false,
      inputMsg: '',
      messages: [],
      loading: false,
      thinkingText: '',
      isMock: process.env.VUE_APP_USE_MOCK === 'true',
    }
  },
  methods: {
    toggleChat() {
      this.visible = !this.visible
    },
    sendSuggestion(text) {
      this.inputMsg = text
      this.sendMessage()
    },
    async sendMessage() {
      const msg = this.inputMsg.trim()
      if (!msg || this.loading) return

      this.messages.push({ role: 'user', content: msg })
      this.inputMsg = ''
      this.loading = true
      this.thinkingText = ''
      this.$nextTick(() => this.scrollToBottom())

      if (this.isMock) {
        await this.sendMockMessage(msg)
      } else {
        await this.sendStreamMessage(msg)
      }
    },

    // Mock 模式：通过 Axios 调用非流式接口（会被 mock 拦截器捕获）
    async sendMockMessage(msg) {
      try {
        const res = await this.$axios.post('/api/aichat/chat', { message: msg })
        const reply = res.data.reply || '抱歉，我暂时无法回答。'
        // 模拟逐字打字效果
        let botMsg = { role: 'bot', content: '' }
        this.messages.push(botMsg)
        const chars = reply.split('')
        for (let i = 0; i < chars.length; i++) {
          botMsg.content += chars[i]
          this.$set(this.messages, this.messages.length - 1, { ...botMsg })
          if (i % 3 === 0) {
            await new Promise(r => setTimeout(r, 30))
            this.scrollToBottom()
          }
        }
      } catch (err) {
        this.messages.push({ role: 'bot', content: '网络异常，请稍后再试。' })
      } finally {
        this.loading = false
        this.thinkingText = ''
        this.scrollToBottom()
      }
    },

    // 正常模式：通过 fetch SSE 调用流式接口
    async sendStreamMessage(msg) {
      try {
        // 从 localStorage 读取 token
        let authHeader = {}
        try {
          const user = JSON.parse(localStorage.getItem('user'))
          if (user && user.token) {
            authHeader['Authorization'] = 'Bearer ' + user.token
          }
        } catch (e) { /* ignore */ }

        const response = await fetch('/api/aichat/stream', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', ...authHeader },
          body: JSON.stringify({ message: msg }),
        })

        if (!response.ok) throw new Error('请求失败')

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let botMsg = { role: 'bot', content: '' }
        this.messages.push(botMsg)
        let buffer = ''

        let reading = true
        while (reading) {
          const { done, value } = await reader.read()
          if (done) { reading = false; break }

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''

          for (const line of lines) {
            if (!line.startsWith('data: ')) continue
            const data = line.slice(6)
            if (data === '[DONE]') break

            try {
              const parsed = JSON.parse(data)
              if (parsed.thinking) {
                this.thinkingText = parsed.thinking
              } else if (parsed.content) {
                this.thinkingText = ''
                botMsg.content += parsed.content
                this.$set(this.messages, this.messages.length - 1, { ...botMsg })
              } else if (parsed.error) {
                botMsg.content = '抱歉，出了点问题：' + parsed.error
                this.$set(this.messages, this.messages.length - 1, { ...botMsg })
              }
            } catch (e) {
              // ignore parse errors
            }
          }
          this.scrollToBottom()
        }
      } catch (err) {
        this.messages.push({ role: 'bot', content: '网络异常，请稍后再试。' })
      } finally {
        this.loading = false
        this.thinkingText = ''
        this.scrollToBottom()
      }
    },

    formatMsg(text) {
      if (!text) return ''
      return text
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\n/g, '<br/>')
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const body = this.$refs.chatBody
        if (body) body.scrollTop = body.scrollHeight
      })
    },
  },
}
</script>

<style scoped>
.ai-chat-fab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff6700, #ff8533);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(255, 103, 0, 0.4);
  z-index: 9999;
  transition: transform 0.2s, box-shadow 0.2s;
}
.ai-chat-fab:hover {
  transform: scale(1.08);
  box-shadow: 0 6px 20px rgba(255, 103, 0, 0.5);
}
.ai-chat-fab.active {
  background: #666;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

.ai-chat-window {
  position: fixed;
  bottom: 88px;
  right: 24px;
  width: 380px;
  height: 520px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.15);
  display: flex;
  flex-direction: column;
  z-index: 9998;
  overflow: hidden;
}

.chat-header {
  background: linear-gradient(135deg, #ff6700, #ff8533);
  color: #fff;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 500;
}
.chat-header-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.chat-header-info i {
  font-size: 20px;
}
.chat-minimize {
  cursor: pointer;
  font-size: 18px;
  opacity: 0.8;
}
.chat-minimize:hover {
  opacity: 1;
}

.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: #f5f5f5;
}

.chat-msg {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  align-items: flex-start;
}
.chat-msg.user {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #999;
  flex-shrink: 0;
}
.chat-msg.bot .msg-avatar {
  background: linear-gradient(135deg, #ff6700, #ff8533);
  color: #fff;
}
.chat-msg.user .msg-avatar {
  background: #409eff;
  color: #fff;
}

.msg-bubble {
  max-width: 75%;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.6;
  word-break: break-word;
}
.chat-msg.bot .msg-bubble {
  background: #fff;
  color: #333;
  border-top-left-radius: 4px;
}
.chat-msg.user .msg-bubble {
  background: #ff6700;
  color: #fff;
  border-top-right-radius: 4px;
}

.suggestion {
  display: inline-block;
  margin: 4px 4px 0 0;
  padding: 4px 10px;
  background: #fff3e8;
  border: 1px solid #ffd6b3;
  border-radius: 14px;
  font-size: 12px;
  cursor: pointer;
  color: #ff6700;
  transition: background 0.2s;
}
.suggestion:hover {
  background: #ffe8d4;
}

.typing .dot-typing {
  display: inline-flex;
  gap: 4px;
  align-items: center;
}
.dot-typing span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #999;
  animation: dotBounce 1.2s infinite;
}
.dot-typing span:nth-child(2) { animation-delay: 0.2s; }
.dot-typing span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dotBounce {
  0%, 80%, 100% { opacity: 0.3; transform: scale(0.8); }
  40% { opacity: 1; transform: scale(1); }
}

.chat-footer {
  padding: 10px 12px;
  border-top: 1px solid #eee;
  background: #fff;
}

/* Transition */
.chat-slide-enter-active, .chat-slide-leave-active {
  transition: all 0.3s ease;
}
.chat-slide-enter, .chat-slide-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}

/* Dark mode support */
[data-theme="dark"] .ai-chat-window {
  background: #1e1e1e;
}
[data-theme="dark"] .chat-body {
  background: #2a2a2a;
}
[data-theme="dark"] .chat-msg.bot .msg-bubble {
  background: #333;
  color: #e0e0e0;
}
[data-theme="dark"] .chat-footer {
  background: #1e1e1e;
  border-top-color: #333;
}
[data-theme="dark"] .suggestion {
  background: #3a2a1a;
  border-color: #5a3a1a;
  color: #ff8533;
}
</style>
