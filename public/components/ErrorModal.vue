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
    <b-modal v-model="show" cancel-disabled hide-header hide-footer size="lg">
      <div class="row justify-content-center">
        <i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
        <div>
          <b>{{ title }}:</b> {{ error }}
        </div>
      </div>
      <div class="row justify-content-center">
        <b-btn
          class="mt-3 close-modal-button"
          variant="outline-secondary"
          block
          @click="onClose"
          >OK</b-btn
        >
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import { EventList } from "../util/events";

export default Vue.extend({
  name: "error-model",

  props: {
    title: {
      type: String as () => string,
      default: "Error",
    },
    error: {
      type: String as () => string,
      default: "Internal server error",
    },
    show: Boolean as () => boolean,
  },

  methods: {
    onClose() {
      this.$emit(EventList.BASIC.CLOSE_EVENT);
    },
  },
});
</script>

<style>
.close-modal-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
}
</style>
