<template>
  <div class="confirm-page">
    <el-page-header @back="$router.back()" content="确认订单" />

    <el-empty v-if="!items.length" description="没有待结算的商品">
      <el-button type="primary" @click="$router.push('/shoppingCart')">返回购物车</el-button>
    </el-empty>

    <template v-else>
      <el-card class="order-card">
        <template #header><span>商品清单</span></template>
        <div v-for="item in items" :key="item.product_id" class="order-item">
          <img :src="item.productImg" class="item-img" :alt="item.productName" />
          <div class="item-info">
            <div class="item-name">{{ item.productName }}</div>
            <div class="item-meta">
              <span>单价 {{ item.price.toFixed(2) }}</span>
              <span>x {{ item.num }}</span>
            </div>
          </div>
          <div class="item-subtotal">{{ (item.price * item.num).toFixed(2) }}</div>
        </div>
      </el-card>

      <div class="order-footer">
        <div class="total-line">
          共 {{ items.length }} 件商品，合计：
          <span class="price">{{ totalPrice.toFixed(2) }} 元</span>
        </div>
        <el-button type="danger" size="large" :loading="submitting" @click="handleSubmit">
          提交订单
        </el-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { orderApi } from '@/api/order'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()
const submitting = ref(false)

const items = computed(() => cartStore.checkedItems)
const totalPrice = computed(() => cartStore.totalPrice)

async function handleSubmit() {
  if (!userStore.user || !items.value.length) return
  submitting.value = true
  try {
    const orderItems = items.value.map(i => ({
      product_id: i.product_id,
      product_num: i.num,
      product_price: i.price,
    }))
    const { data } = await orderApi.addOrder({
      user_id: userStore.user.userId,
      items: orderItems,
    })
    if (data.code === '200' && data.order_id) {
      cartStore.clearCart()
      ElMessage.success('下单成功，正在跳转支付...')
      router.push({ path: '/payment', query: { order_id: String(data.order_id) } })
    } else if (data.code === '012') {
      ElMessage.warning('请勿重复提交')
    } else if (data.code === '014') {
      ElMessage.error('库存不足，请调整数量')
    } else {
      ElMessage.error('下单失败：' + data.code)
    }
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.confirm-page { max-width: 800px; margin: 20px auto; padding: 0 20px; }
.order-card { margin-top: 20px; }
.order-item { display: flex; align-items: center; gap: 16px; padding: 12px 0; border-bottom: 1px solid #f0f0f0; }
.order-item:last-child { border-bottom: none; }
.item-img { width: 80px; height: 80px; object-fit: contain; }
.item-info { flex: 1; }
.item-name { font-size: 14px; margin-bottom: 6px; }
.item-meta { color: #999; font-size: 13px; display: flex; gap: 12px; }
.item-subtotal { font-size: 16px; color: #e4393c; font-weight: bold; min-width: 80px; text-align: right; }
.order-footer { display: flex; justify-content: flex-end; align-items: center; gap: 20px; margin-top: 24px; padding: 16px 0; border-top: 2px solid #e4393c; }
.total-line { font-size: 14px; }
.price { color: #e4393c; font-size: 20px; font-weight: bold; }
</style>
