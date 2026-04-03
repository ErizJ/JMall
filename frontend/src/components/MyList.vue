<!--
 * @Description: 商品卡片列表组件
 -->
<template>
  <div class="product-grid">
    <div class="product-card" v-for="item in list" :key="item.product_id">
      <el-popover placement="top" v-if="isDelete">
        <p>确定删除吗？</p>
        <div style="text-align: right; margin: 10px 0 0">
          <el-button type="primary" size="mini" @click="deleteCollect(item.product_id)">确定</el-button>
        </div>
        <i class="el-icon-close card-delete" slot="reference"></i>
      </el-popover>
      <router-link :to="{ path: '/goods/details', query: { productID: item.product_id } }" class="card-link">
        <div class="card-img">
          <img :src="$target + item.product_picture" alt="" />
        </div>
        <div class="card-info">
          <h3 class="card-name">{{ item.product_name }}</h3>
          <p class="card-desc">{{ item.product_title }}</p>
          <div class="card-price">
            <span class="price-now">¥{{ item.product_selling_price }}</span>
            <span class="price-old" v-show="item.product_price != item.product_selling_price">
              ¥{{ item.product_price }}
            </span>
          </div>
        </div>
      </router-link>
    </div>
    <div class="product-card more-card" v-show="isMore && list.length >= 1">
      <router-link :to="{ path: '/goods', query: { categoryID: categoryID } }" class="more-link">
        <i class="el-icon-right"></i>
        <span>浏览更多</span>
      </router-link>
    </div>
  </div>
</template>
<script>
export default {
  name: 'MyList',
  props: ['list', 'isMore', 'isDelete'],
  computed: {
    categoryID() {
      let ids = []
      if (this.list && this.list.length) {
        for (let i = 0; i < this.list.length; i++) {
          const id = this.list[i].category_id
          if (!ids.includes(id)) ids.push(id)
        }
      }
      return ids
    },
  },
  methods: {
    deleteCollect(product_id) {
      this.$axios
        .post('/api/user/collect/deleteCollect', {
          user_id: this.$store.getters.getUser.user_id,
          product_id: product_id,
        })
        .then((res) => {
          if (res.data.code === '001') {
            for (let i = 0; i < this.list.length; i++) {
              if (this.list[i].product_id == product_id) {
                this.list.splice(i, 1)
                break
              }
            }
            this.notifySucceed(res.data.msg)
          } else {
            this.notifyError(res.data.msg)
          }
        })
        .catch((err) => Promise.reject(err))
    },
  },
}
</script>
<style scoped>
.product-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.product-card {
  width: 228px;
  background: var(--bg-white, #fff);
  border-radius: var(--radius, 8px);
  overflow: hidden;
  transition: all 0.25s ease;
  position: relative;
  border: 1px solid transparent;
}
.product-card:hover {
  box-shadow: var(--shadow-md, 0 8px 24px rgba(0,0,0,0.1));
  transform: translateY(-4px);
  border-color: var(--border, #eee);
}

.card-link { display: block; color: inherit; }

.card-img {
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg, #fafafa);
  overflow: hidden;
  padding: 16px;
}
.card-img img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transition: transform 0.3s;
}
.product-card:hover .card-img img { transform: scale(1.05); }

.card-info { padding: 12px 16px 16px; }
.card-name {
  font-size: 14px; font-weight: 500; color: var(--text, #333);
  line-height: 1.4; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 4px;
}
.card-desc {
  font-size: 12px; color: var(--text-muted, #999);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 10px; line-height: 1.4;
}
.card-price { display: flex; align-items: baseline; gap: 6px; }
.price-now { font-size: 18px; font-weight: 600; color: var(--primary, #ff6700); }
.price-old { font-size: 12px; color: var(--text-muted, #bbb); text-decoration: line-through; }

.card-delete {
  position: absolute; top: 8px; right: 8px; width: 24px; height: 24px; line-height: 24px;
  text-align: center; border-radius: 50%; background: rgba(0,0,0,0.4); color: #fff;
  font-size: 12px; cursor: pointer; opacity: 0; transition: opacity 0.2s; z-index: 2;
}
.product-card:hover .card-delete { opacity: 1; }
.card-delete:hover { background: #f56c6c; }

.more-card {
  display: flex; align-items: center; justify-content: center;
  background: var(--bg, #fafafa); border: 1px dashed var(--border, #ddd);
}
.more-card:hover { border-color: var(--primary, #ff6700); background: var(--bg-white, #fff); }
.more-link {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  color: var(--text-muted, #999); font-size: 14px;
}
.more-link i { font-size: 28px; color: var(--border, #ccc); transition: color 0.2s; }
.more-card:hover .more-link, .more-card:hover .more-link i { color: var(--primary, #ff6700); }
</style>
