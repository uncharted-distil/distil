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
  <div class="w-100 h-100 flex justify-content-center align-items-center">
    <div class="d-flex justify-content-center align-items-center flex-column">
      <p>
        <strong>{{ state }}</strong>
      </p>
      <b-spinner />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "LoadingSpinner",
  props: {
    state: { type: String as () => string, default: "" },
  },
});
</script>

<style scoped>
.flex {
  display: flex;
}
</style>
