<template>
  <div class="payment-page">
    <el-page-header @back="$router.push('/order')" content="订单支付" />

    <!-- 加载中 -->
    <div v-if="loading" class="loading-box">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <p>正在加载订单信息...</p>
    </div>

    <!-- 支付表单 -->
    <template v-else-if="step === 'pay'">
      <el-card class="pay-card">
        <template #header><span>订单信息</span></template>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="订单号">{{ orderId }}</el-descriptions-item>
          <el-descriptions-item label="商品数量">{{ orderItems.length }} 件</el-descriptions-item>
          <el-descriptions-item label="订单金额">
            <span class="price">{{ totalPrice.toFixed(2) }} 元</span>
          </el-descriptions-item>
        </el-descriptions>

        <div class="items-preview">
          <div v-for="item in orderItems" :key="item.id" class="preview-item">
            <img :src="item.productImg" class="preview-img" :alt="item.productName" />
            <span class="preview-name">{{ item.productName }}</span>
            <span class="preview-qty">x{{ item.product_num }}</span>
          </div>
        </div>
      </el-card>

      <el-card class="pay-card">
        <template #header><span>选择支付方式</span></template>
        <el-radio-group v-model="channel" class="channel-group">
          <el-radio value="mock" size="large">模拟支付（测试）</el-radio>
        </el-radio-group>
      </el-card>

      <div class="pay-footer">
        <div class="pay-amount">
          应付：<span class="price">{{ totalPrice.toFixed(2) }} 元</span>
        </div>
        <el-button type="danger" size="large" :loading="paying" @click="handlePay">
          立即支付
        </el-button>
      </div>
    </template>

    <!-- Mock 支付确认 -->
    <template v-else-if="step === 'mock-confirm'">
      <el-card class="pay-card mock-card">
        <el-result icon="info" title="模拟支付">
          <template #sub-title>
            <p>支付金额：<span class="price">{{ totalPrice.toFixed(2) }} 元</span></p>
            <p style="color: #999; font-size: 13px;">这是模拟支付环境，点击下方按钮确认支付</p>
          </template>
          <template #extra>
            <el-button type="primary" size="large" :loading="confirming" @click="handleMockConfirm">
              确认支付
            </el-button>
            <el-button size="large" @click="$router.push('/order')">取消</el-button>
          </template>
        </el-result>
      </el-card>
    </template>

    <!-- 支付结果 -->
    <template v-else-if="step === 'result'">
      <el-card class="pay-card">
        <el-result
          :icon="paySuccess ? 'success' : 'error'"
          :title="paySuccess ? '支付成功' : '支付失败'"
          :sub-title="paySuccess ? '感谢您的购买' : '请稍后重试'"
        >
          <template #extra>
            <el-button type="primary" @click="$router.push('/order')">查看订单</el-button>
            <el-button @click="$router.push('/home')">继续购物</el-button>
          </template>
        </el-result>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { orderApi } from '@/api/order'
import { paymentApi } from '@/api/payment'
import type { OrderItem } from '@/types'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const paying = ref(false)
const confirming = ref(false)
const step = ref<'pay' | 'mock-confirm' | 'result'>('pay')
const paySuccess = ref(false)

const orderId = ref(0)
const orderItems = ref<OrderItem[]>([])
const channel = ref('mock')
const paymentNo = ref('')

const totalPrice = computed(() =>
  orderItems.value.reduce((sum, i) => sum + i.product_price * i.product_num, 0)
)

onMounted(async () => {
  const oid = Number(route.query.order_id)
  if (!oid) {
    ElMessage.error('缺少订单号')
    router.push('/order')
    return
  }
  orderId.value = oid

  try {
    const { data } = await orderApi.getDetails(oid)
    if (data.code === '200' && data.orders?.length) {
      orderItems.value = data.orders
      // 如果订单已支付，直接跳到结果页
      if (data.orders[0].status === 1) {
        paySuccess.value = true
        step.value = 'result'
      }
    } else {
      ElMessage.error('订单不存在')
      router.push('/order')
    }
  } finally {
    loading.value = false
  }
})

async function handlePay() {
  paying.value = true
  try {
    const { data } = await paymentApi.create(orderId.value, channel.value)
    if (data.code === '200' && data.payment_no) {
      paymentNo.value = data.payment_no
      if (channel.value === 'mock') {
        step.value = 'mock-confirm'
      }
      // 真实渠道时可以跳转 data.pay_url
    } else if (data.code === '005') {
      ElMessage.warning('已有进行中的支付，请勿重复操作')
    } else if (data.code === '011') {
      ElMessage.warning('订单状态不允许支付')
      paySuccess.value = true
      step.value = 'result'
    } else {
      ElMessage.error('创建支付失败：' + data.code)
    }
  } finally {
    paying.value = false
  }
}

async function handleMockConfirm() {
  if (!paymentNo.value) return
  confirming.value = true
  try {
    const { data } = await paymentApi.mockPay(paymentNo.value)
    if (data.code === '200') {
      paySuccess.value = true
      step.value = 'result'
      ElMessage.success('支付成功')
    } else if (data.code === '007') {
      ElMessage.error('支付已过期，请重新下单')
      step.value = 'result'
      paySuccess.value = false
    } else {
      ElMessage.error('支付失败')
      step.value = 'result'
      paySuccess.value = false
    }
  } finally {
    confirming.value = false
  }
}
</script>

<style scoped>
.payment-page { max-width: 700px; margin: 20px auto; padding: 0 20px; }
.loading-box { text-align: center; padding: 60px 0; color: #999; }
.pay-card { margin-top: 20px; }
.mock-card { text-align: center; }
.price { color: #e4393c; font-weight: bold; font-size: 18px; }
.items-preview { margin-top: 16px; display: flex; flex-wrap: wrap; gap: 12px; }
.preview-item { display: flex; align-items: center; gap: 8px; padding: 6px 10px; background: #f9f9f9; border-radius: 6px; }
.preview-img { width: 40px; height: 40px; object-fit: contain; }
.preview-name { font-size: 13px; max-width: 150px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.preview-qty { color: #999; font-size: 12px; }
.channel-group { display: flex; flex-direction: column; gap: 12px; }
.pay-footer { display: flex; justify-content: flex-end; align-items: center; gap: 20px; margin-top: 24px; padding: 16px 0; border-top: 2px solid #e4393c; }
.pay-amount { font-size: 14px; }
</style>
