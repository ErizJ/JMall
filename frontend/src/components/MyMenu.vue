<!--
 * @Description: 标签切换菜单组件
 -->
<template>
  <div class="tab-menu">
    <span
      v-for="item in val"
      :key="item"
      :class="['tab-item', activeClass == item ? 'active' : '']"
      @click="select(item)"
    >
      <slot :name="item"></slot>
    </span>
  </div>
</template>
<script>
export default {
  props: ['val'],
  name: 'MyMenu',
  data() {
    return { activeClass: 1 }
  },
  methods: {
    select(val) {
      this.activeClass = val
    },
  },
  watch: {
    activeClass(val) {
      this.$emit('fromChild', val)
    },
  },
}
</script>
<style scoped>
.tab-menu { display: flex; gap: 4px; }
.tab-item {
  padding: 4px 14px; font-size: 13px; color: var(--text-secondary, #666);
  border-radius: 16px; cursor: pointer; transition: all 0.2s; user-select: none;
}
.tab-item:hover { color: var(--primary, #ff6700); background: rgba(255, 103, 0, 0.06); }
.tab-item.active { color: #fff; background: var(--primary, #ff6700); }
</style>
