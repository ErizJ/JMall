import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { CartItem } from '@/types'

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])

  const totalCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.num, 0),
  )

  const checkedItems = computed(() => items.value.filter((i) => i.check))

  const totalPrice = computed(() =>
    checkedItems.value.reduce((sum, item) => sum + item.price * item.num, 0),
  )

  const isAllChecked = computed(
    () => items.value.length > 0 && items.value.every((i) => i.check),
  )

  function setCart(cartItems: CartItem[]) {
    items.value = cartItems
  }

  function addItem(item: CartItem) {
    items.value.unshift(item)
  }

  function updateItem(productId: number, num: number) {
    const idx = items.value.findIndex((i) => i.product_id === productId)
    if (idx !== -1) items.value[idx].num = num
  }

  function removeItem(productId: number) {
    items.value = items.value.filter((i) => i.product_id !== productId)
  }

  function toggleCheck(productId: number) {
    const item = items.value.find((i) => i.product_id === productId)
    if (item) item.check = !item.check
  }

  function checkAll(val: boolean) {
    items.value.forEach((i) => (i.check = val))
  }

  function clearCart() {
    items.value = []
  }

  return {
    items,
    totalCount,
    checkedItems,
    totalPrice,
    isAllChecked,
    setCart,
    addItem,
    updateItem,
    removeItem,
    toggleCheck,
    checkAll,
    clearCart,
  }
})
