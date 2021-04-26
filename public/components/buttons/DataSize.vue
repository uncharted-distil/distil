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
  <b-dropdown
    :text="currentSize.toString()"
    ref="dropdown"
    variant="light"
    size="sm"
  >
    <b-dropdown-form form-class="data-size-form">
      <b-form-group label="Number of data displayed">
        <b-input-group size="sm" :append="numDisplay">
          <b-form-input
            v-model="dataSize"
            number
            min="1"
            :max="total"
            step="1"
            type="range"
          />
        </b-input-group>
      </b-form-group>
      <b-form-group>
        <b-overlay
          :show="isUpdating"
          class="d-inline-block"
          opacity="0.6"
          rounded
          spinner-small
          spinner-variant="primary"
        >
          <b-button
            :disabled="isDisabled"
            size="sm"
            variant="primary"
            @click="onUpdate"
          >
            Update
          </b-button>
        </b-overlay>
      </b-form-group>
    </b-dropdown-form>
  </b-dropdown>
</template>

<script lang="ts">
import Vue from "vue";
import { BDropdown } from "bootstrap-vue";
import { getters as routeGetters } from "../../store/route/module";
/**
 * Button to change the size of a current data.
 * @param {Number} currentSize - the current number of data.
 * @param {Number} total - the total number of data.
 * @emits updated - a boolean to signal that the size has been updated.
 */
export default Vue.extend({
  name: "data-size",

  props: {
    currentSize: { type: Number, default: 1 },
    total: { type: Number, default: 1 },
    excluded: Boolean,
  },

  data() {
    return {
      dataSize: this.currentSize,
      isUpdating: false,
    };
  },

  computed: {
    /* Display the selected number of data displayed. */
    numDisplay(): string {
      return this.dataSize.toString();
    },

    /* Disable the Update button */
    isDisabled(): boolean {
      return this.isUpdating || this.dataSize === this.currentSize;
    },
    routeDataSize(): number {
      return routeGetters.getRouteDataSize(this.$store);
    },
  },

  methods: {
    /* Set the dataSize in the URI, and reload the page */
    onUpdate() {
      this.isUpdating = true;
      this.$emit("submit", this.dataSize);
    },
  },

  watch: {
    currentSize(oldValue, newValue) {
      if (oldValue === newValue) return;
      this.dataSize = this.currentSize; // Set the input range to the appropriate value.
      this.isUpdating = false;
      const dropdown = this.$refs.dropdown as typeof BDropdown;
      dropdown.hide();
    },
    routeDataSize() {
      this.dataSize = this.routeDataSize;
      this.onUpdate();
    },
  },
});
</script>

<style scoped>
.data-size-form {
  width: 300px;
}
</style>
