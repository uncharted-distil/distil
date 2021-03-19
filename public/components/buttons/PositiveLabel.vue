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

<!-- Dropdown to select a positive label for Binary Classification task -->
<template>
  <b-form-group
    label="Positive Label:"
    label-class="font-weight-bold"
    label-cols="auto"
    label-size="sm"
  >
    <b-form-select
      id="positive-label"
      v-model="positiveLabel"
      :options="options"
      size="sm"
    />
  </b-form-group>
</template>

<script lang="ts">
import Vue from "vue";
import { findBestMatch } from "string-similarity";
import { VariableSummary } from "../../store/dataset/index";
import { getters as routeGetters } from "../../store/route/module";
import { overlayRouteEntry, RouteArgs } from "../../util/routes";

export default Vue.extend({
  name: "PositiveLabel",

  props: {
    targetSummary: {
      type: Object as () => VariableSummary,
      default: null,
    },
  },

  data() {
    return {
      positiveLabel: null as string,
    };
  },

  computed: {
    // Define the posible options for the positive label <select>.
    options(): string[] {
      // retreive the target variable buckets
      const buckets = this.targetSummary?.baseline?.buckets;
      if (!buckets) return;

      // Use the buckets key as <options>
      const options = buckets.map((bucket) => bucket.key);

      // Pre-select the label that's most likely to be a positive label
      this.findPositiveLabel(options);

      return options;
    },

    routePositiveLabel(): string {
      return routeGetters.getPositiveLabel(this.$store);
    },
  },

  watch: {
    positiveLabel(label: string, oldLabel: string): void {
      if (label === oldLabel) return;
      if (label === this.routePositiveLabel) return;
      this.updateRoute({ positiveLabel: label });
    },
  },

  beforeMount() {
    // If the positive label is already set in the route, pre-select it.
    if (!!this.routePositiveLabel && !this.positiveLabel) {
      this.positiveLabel = this.routePositiveLabel;
    }
  },

  methods: {
    // Find which options is most suited to be the positive label
    findPositiveLabel(options: string[]): void {
      // Do not find a new label if the positiveLabel is already set
      if (!!this.positiveLabel) return;

      // findBestMatch();

      const label = options[0];
      this.positiveLabel = label;
    },

    updateRoute(args: RouteArgs): void {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>
