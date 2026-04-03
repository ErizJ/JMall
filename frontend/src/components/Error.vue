<template>
  <div class="error">
    
    <div class="main" ></div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      loading: true,
    }
  },
  methods: {
    // changeLoading() {
    //   setTimeout(() => {
    //     this.loading != this.loading
    //   }, 3000)
    // },
  },
  // mounted() {
  //   this.changeLoading()
  // },
}
</script>

<style>
.error {
  height: 700px;
  width: 1225px;
  margin: 0 auto;
}
.error .main {
  margin: 0 362.5px;
  height: 700px;
  background: url(../assets/imgs/error.png) no-repeat;
}
</style>