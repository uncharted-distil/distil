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
  <div>
    <b-button title="Select all items on page" @click="onSelectAll"
      >Select All</b-button
    >
    <b-button
      title="Annotate selected items to positive"
      @click="onButtonClick(positive)"
    >
      <span class="stacked-icons">
        <i class="fa fa-circle fa-stack-1x" />
        <i class="fa fa-plus-circle text-success fa-stack-1x" />
      </span>
      Positive
    </b-button>
    <b-button
      title="Annotate selected items to negative"
      @click="onButtonClick(negative)"
    >
      <span class="stacked-icons">
        <i class="fa fa-circle fa-stack-1x" />
        <i class="fa fa-minus-circle red fa-stack-1x" />
      </span>
      Negative</b-button
    >
    <b-button
      title="Annotate select items to negative"
      @click="onButtonClick(unlabeled)"
      >Unlabeled</b-button
    >
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { LowShotLabels } from "../../util/data";
import { EventList } from "../../util/events";
export default Vue.extend({
  name: "label-header-buttons",
  computed: {
    negative(): string {
      return LowShotLabels.negative;
    },
    positive(): string {
      return LowShotLabels.positive;
    },
    unlabeled(): string {
      return LowShotLabels.unlabeled;
    },
  },
  methods: {
    onButtonClick(event: string) {
      this.$eventBus.$emit(EventList.LABEL.ANNOTATION_EVENT, event);
    },
    onSelectAll() {
      this.$eventBus.$emit(EventList.LABEL.SELECT_ALL_EVENT);
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
  width: 2em;
  height: 1em;
  line-height: 1em;
  vertical-align: middle;
}
</style>
