<template>
  <div class="container-fluid d-flex flex-column h-100">
    <div class="row flex-0-nav"></div>
    <div class="col-12 col-md-12">
      <h4 class="export-header-label">Models exported successfully</h4>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "export-success",
});
</script>

<style>
.export-header-label {
  padding: 28px 0 28px 0;
  text-align: center;
}
</style>
