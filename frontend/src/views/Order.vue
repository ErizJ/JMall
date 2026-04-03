<template>
  <div class="order-page">
    <el-page-header @back="$router.back()" content="我的订单" />

    <el-empty v-if="!loading && !groupedOrders.length" description="暂无订单">
      <el-button type="primary" @click="$router.push('/home')">去逛逛</el-button>
    </el-empty>

    <div v-loading="loading" class="order-list">
      <el-card v-for="group in groupedOrders" :key="group.orderId" class="order-card">
        <template #header>
          <div class="order-header">
            <span>订单号：{{ group.orderId }}</span>
            <span>{{ group.items[0].order_time }}</span>
            <el-tag :type="statusTagType(group.status)" size="small">
              {{ statusText(group.status) }}
            </el-tag>
          </div>
        </template>

        <div v-for="item in group.items" :key="item.id" class="order-item">
          <img :src="item.productImg" class="item-img" :alt="item.productName" />
          <div class="item-info">
            <div>{{ item.productName }}</div>
            <div class="item-meta">x{{ item.product_num }}</div>
          </div>
          <div class="item-price">{{ item.product_price.toFixed(2) }} 元</div>
        </div>

        <div class="order-actions">
          <span class="order-total">
            合计：<span class="price">{{ group.total.toFixed(2) }} 元</span>
          </span>
          <div class="action-btns">
            <el-button
              v-if="group.status === 0"
              type="danger"
              size="small"
              @click="goPayment(group.orderId)"
            >
              去支付
            </el-button>
            <el-button
              v-if="group.status === 0"
              size="small"
              @click="handleDelete(group.orderId)"
            >
              取消订单
            </el-button>
            <el-button
              v-if="group.status === 1"
              size="small"
              @click="handleRefund(group)"
            >
              申请退款
            </el-button>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { orderApi } from '@/api/order'
import { paymentApi } from '@/api/payment'
import { ORDER_STATUS_MAP } from '@/types'
import type { OrderItem } from '@/types'

interface OrderGroup {
  orderId: number
  status: number
  total: number
  items: OrderItem[]
}

const router = useRouter()
const userStore = useUserStore()
const loading = ref(true)
const orders = ref<OrderItem[]>([])

const groupedOrders = computed<OrderGroup[]>(() => {
  const map = new Map<number, OrderItem[]>()
  for (const item of orders.value) {
    if (!map.has(item.order_id)) map.set(item.order_id, [])
    map.get(item.order_id)!.push(item)
  }
  const groups: OrderGroup[] = []
  for (const [orderId, items] of map) {
    groups.push({
      orderId,
      status: items[0].status,
      total: items.reduce((s, i) => s + i.product_price * i.product_num, 0),
      items,
    })
  }
  return groups
})

function statusText(status: number) {
  return ORDER_STATUS_MAP[status] || '未知'
}

function statusTagType(status: number) {
  const map: Record<number, string> = { 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }
  return (map[status] || 'info') as any
}

onMounted(async () => {
  if (!userStore.user) return
  try {
    const { data } = await orderApi.getOrder(userStore.user.userId)
    if (data.code === '200') {
      orders.value = data.orders || []
    }
  } finally {
    loading.value = false
  }
})

function goPayment(orderId: number) {
  router.push({ path: '/payment', query: { order_id: String(orderId) } })
}

async function handleDelete(orderId: number) {
  try {
    await ElMessageBox.confirm('取消订单后库存将释放，确定取消？', '取消订单')
  } catch { return }

  const { data } = await orderApi.deleteOrder(orderId)
  if (data.code === '200') {
    orders.value = orders.value.filter(o => o.order_id !== orderId)
    ElMessage.success('订单已取消')
  } else if (data.code === '005') {
    ElMessage.warning('已支付订单不可取消，请先退款')
  } else {
    ElMessage.error('操作失败')
  }
}

async function handleRefund(group: OrderGroup) {
  try {
    await ElMessageBox.confirm(
      `确定退款 ${group.total.toFixed(2)} 元？`,
      '申请退款'
    )
  } catch { return }

  // 先查支付单
  const { data: listData } = await paymentApi.list(userStore.user!.userId)
  if (listData.code !== '200' || !listData.payments?.length) {
    ElMessage.error('未找到支付记录')
    return
  }
  const payment = listData.payments.find(
    (p: any) => p.order_id === group.orderId && p.status === 2
  )
  if (!payment) {
    ElMessage.error('未找到对应的成功支付记录')
    return
  }

  const amountFen = Math.round(group.total * 100)
  const { data } = await paymentApi.refund(payment.payment_no, amountFen, '用户申请退款')
  if (data.code === '200') {
    ElMessage.success('退款成功')
    // 刷新订单列表
    const { data: refreshData } = await orderApi.getOrder(userStore.user!.userId)
    if (refreshData.code === '200') orders.value = refreshData.orders || []
  } else if (data.code === '016') {
    ElMessage.warning('退款处理中，请勿重复操作')
  } else {
    ElMessage.error('退款失败：' + data.code)
  }
}
</script>

<style scoped>
.order-page { max-width: 800px; margin: 20px auto; padding: 0 20px; }
.order-list { margin-top: 20px; }
.order-card { margin-bottom: 16px; }
.order-header { display: flex; align-items: center; gap: 16px; font-size: 13px; color: #666; }
.order-item { display: flex; align-items: center; gap: 12px; padding: 10px 0; border-bottom: 1px solid #f5f5f5; }
.order-item:last-child { border-bottom: none; }
.item-img { width: 60px; height: 60px; object-fit: contain; }
.item-info { flex: 1; font-size: 14px; }
.item-meta { color: #999; font-size: 12px; margin-top: 4px; }
.item-price { color: #333; font-weight: 500; }
.order-actions { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; padding-top: 12px; border-top: 1px solid #eee; }
.order-total { font-size: 14px; }
.price { color: #e4393c; font-weight: bold; }
.action-btns { display: flex; gap: 8px; }
</style>
