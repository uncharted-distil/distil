<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

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
