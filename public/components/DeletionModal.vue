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
  <b-modal :id="modalId" @ok="okHandler">
    <template #modal-title> Deleting {{ target }}</template>
    This action can NOT be undone. Do you want to delete {{ target }}?
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { EventList } from "../util/events";

export default Vue.extend({
  name: "deletion-modal",
  props: {
    target: { type: String as () => string, default: "" },
  },
  data() {
    return { modalId: "deletion-modal" };
  },
  watch: {
    target() {
      if (!this.target.length) {
        return;
      }
      // show deletion prompt if target changed
      this.$bvModal.show(this.modalId);
    },
  },
  methods: {
    okHandler() {
      this.$emit(EventList.MODEL.DELETE_EVENT);
    },
  },
});
</script>
