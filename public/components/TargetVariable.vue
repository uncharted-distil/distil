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
    <variable-facets
      class="target-summary"
      enable-highlighting
      enable-type-change
      :summaries="targetSummaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    />

    <!-- Dropdown to select a positive label for Binary Classification task -->
    <b-form-group
      v-if="options"
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
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as routeGetters } from "../store/route/module";
import { TARGET_VAR_INSTANCE } from "../store/route/index";
import { VariableSummary } from "../store/dataset/index";
import { Activity } from "../util/userEvents";
import { overlayRouteEntry, RouteArgs } from "../util/routes";

export default Vue.extend({
  name: "TargetVariable",

  components: {
    VariableFacets,
  },

  data() {
    return {
      instanceName: TARGET_VAR_INSTANCE,
      logActivity: Activity.DATA_PREPARATION,
      positiveLabel: null as string,
    };
  },

  computed: {
    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store);
    },

    // Define the posible options for the positive label <select>.
    options(): string[] {
      // Check that we are on a binary classification task
      if (!routeGetters.isBinaryClassification(this.$store)) return;

      // retreive the target variable buckets
      const buckets = this.targetSummaries?.[0]?.baseline?.buckets;
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
