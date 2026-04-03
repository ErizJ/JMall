<template>
  <div class="cart-page">
    <el-page-header @back="$router.back()" content="购物车" />

    <el-empty v-if="!cartStore.items.length" description="购物车是空的">
      <el-button type="primary" @click="$router.push('/home')">去逛逛</el-button>
    </el-empty>

    <template v-else>
      <el-table :data="cartStore.items" style="margin-top: 20px">
        <el-table-column width="55">
          <template #header>
            <el-checkbox
              :model-value="cartStore.isAllChecked"
              @change="(val: boolean) => cartStore.checkAll(val)"
            />
          </template>
          <template #default="{ row }">
            <el-checkbox
              :model-value="row.check"
              @change="cartStore.toggleCheck(row.product_id)"
            />
          </template>
        </el-table-column>
        <el-table-column label="商品" min-width="200">
          <template #default="{ row }">
            <div class="product-cell">
              <img :src="row.productImg" class="product-img" :alt="row.productName" />
              <span>{{ row.productName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="120">
          <template #default="{ row }">¥{{ row.price.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="数量" width="180">
          <template #default="{ row }">
            <el-input-number
              :model-value="row.num"
              :min="1"
              :max="row.maxNum"
              size="small"
              @change="(val: number) => handleUpdateNum(row.product_id, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="小计" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ (row.price * row.num).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleDelete(row.product_id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="cart-footer">
        <div class="total">
          已选 {{ cartStore.checkedItems.length }} 件，合计：
          <span class="price">¥{{ cartStore.totalPrice.toFixed(2) }}</span>
        </div>
        <el-button
          type="danger"
          size="large"
          :disabled="!cartStore.checkedItems.length"
          @click="handleCheckout"
        >
          去结算
        </el-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { cartApi } from '@/api/cart'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

onMounted(async () => {
  if (!userStore.user) return
  const { data } = await cartApi.getCart(userStore.user.userId)
  if (data.code === '200') {
    cartStore.setCart((data.items || []).map((i: any) => ({ ...i, check: false })))
  }
})

async function handleUpdateNum(productId: number, num: number) {
  if (!userStore.user) return
  const { data } = await cartApi.updateCart({ user_id: userStore.user.userId, product_id: productId, num })
  if (data.code === '200') {
    cartStore.updateItem(productId, num)
  } else {
    ElMessage.warning('更新失败')
  }
}

async function handleDelete(productId: number) {
  if (!userStore.user) return
  const { data } = await cartApi.deleteCart({ user_id: userStore.user.userId, product_id: productId })
  if (data.code === '200') {
    cartStore.removeItem(productId)
    ElMessage.success('已删除')
  }
}

function handleCheckout() {
  if (!cartStore.checkedItems.length) return
  router.push('/confirmOrder')
}
</script>

<style scoped>
.cart-page { max-width: 1000px; margin: 20px auto; padding: 0 20px; }
.product-cell { display: flex; align-items: center; gap: 10px; }
.product-img { width: 60px; height: 60px; object-fit: contain; }
.price { color: #e4393c; font-weight: bold; }
.cart-footer { display: flex; justify-content: flex-end; align-items: center; gap: 20px; margin-top: 20px; padding: 16px 0; border-top: 1px solid #eee; }
.total { font-size: 14px; }
</style>
