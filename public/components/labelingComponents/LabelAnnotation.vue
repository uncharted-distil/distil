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
  <span class="stacked-icons">
    <i v-if="shouldRender" class="fa fa-circle fa-stack-1x p-1" />
    <i :class="getLabel" />
  </span>
</template>

<script lang="ts">
import Vue from "vue";
import { LowShotLabels, LOW_SHOT_LABEL_COLUMN_NAME } from "../../util/data";
export default Vue.extend({
  name: "label-annotation",
  props: {
    item: Object as () => any,
  },
  computed: {
    getLabel(): string {
      switch (this.item[LOW_SHOT_LABEL_COLUMN_NAME].value) {
        case LowShotLabels.positive:
          return "fa fa-plus-circle text-success p-1 fa-stack-1x";
          break;
        case LowShotLabels.negative:
          return "fa fa-minus-circle red p-1 fa-stack-1x";
          break;
        default:
          return "d-none";
          break;
      }
    },
    shouldRender(): boolean {
      return this.getLabel !== "d-none";
    },
  },
});
</script>

<style scoped>
.red {
  color: var(--red);
}
.stacked-icons {
  position: relative;
  display: inline-block;
  width: 1em;
  height: 1em;
  line-height: 1em;
  vertical-align: middle;
}
</style>
