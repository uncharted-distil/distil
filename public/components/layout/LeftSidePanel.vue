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
  <aside class="d-flex flex-column p-2 h-100 left-side-panel-container">
    <header class="mb-1">
      <h5>{{ panelTitle }}</h5>
      <!-- <close-button @close="onClose" /> -->
    </header>
    <div class="panel-body">
      <slot />
    </div>
  </aside>
</template>

<script lang="ts">
import Vue from "vue";
// import CloseButton from "../layout/CloseButton.vue";

export default Vue.extend({
  name: "LeftSidePanel",

  // components: { CloseButton },

  props: {
    // isOpen: { type: Boolean, default: true },
    panelTitle: { type: String, default: "" },
  },

  // methods: {
  //   onClose(): void {
  //     this.$emit("close-panel");
  //   },
  // },
});
</script>

<style scoped>
.left-side-panel-container {
  background-color: var(--bg-color-side-panel);
  border-color: var(--border-color);
  border-style: solid;
  border-width: 0 1px;
  width: var(--width-left-side-panel);
  z-index: var(--z-index-left-side-panel);
}
</style>
