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
  <div class="form-container">
    <div>
      <b-button
        :disabled="
          isLoading || !minimumRequirementsMet || !annotationHasChanged
        "
        size="lg"
        @click="onEvent(applyEvent)"
        title="must have 1 positive and negative label"
      >
        <template v-if="isLoading">
          <div v-html="spinnerHTML" />
        </template>
        <template v-else> Search Similar </template>
      </b-button>
    </div>
    <div>
      <b-button size="lg" @click="onEvent(exportEvent)">Export</b-button>
      <b-button size="lg" variant="primary" @click="onEvent(saveEvent)">
        Save
      </b-button>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { circleSpinnerHTML } from "../../util/spinner";
import { getters as datasetGetters } from "../../store/dataset/module";
import { VariableSummary } from "../../store/dataset";
import { Dictionary } from "../../util/dict";
import { LowShotLabels } from "../../util/data";
import { getters as routeGetters } from "../../store/route/module";
const enum COMPONENT_EVENT {
  EXPORT = "export",
  SAVE = "save",
  APPLY = "apply",
}
export default Vue.extend({
  name: "create-labeling-form",
  props: {
    isLoading: { type: Boolean as () => boolean, default: false },
    lowShotSummary: Object as () => VariableSummary,
    labelFeatureName: { type: String, default: "" },
  },
  computed: {
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    saveEvent(): string {
      return COMPONENT_EVENT.SAVE;
    },
    applyEvent(): string {
      return COMPONENT_EVENT.APPLY;
    },
    exportEvent(): string {
      return COMPONENT_EVENT.EXPORT;
    },
    annotationHasChanged(): boolean {
      return routeGetters.getAnnotationHasChanged(this.$store);
    },
    minimumRequirementsMet(): boolean {
      const keys = new Map(
        this.lowShotSummary?.baseline?.buckets.map((b) => [b.key, true])
      );
      if (!keys) {
        return false;
      }
      return (
        keys.has(LowShotLabels.positive) && keys.has(LowShotLabels.negative)
      );
    },
    lowShotLabel(): Dictionary<VariableSummary> {
      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );
      return summaryDictionary
        ? summaryDictionary[this.labelFeatureName]
        : null;
    },
  },
  methods: {
    onEvent(event: COMPONENT_EVENT) {
      this.$emit(event);
    },
  },
});
</script>

<style scoped>
.form-container {
  display: flex;
  justify-content: space-between;
  height: 15%;
}

.check-box {
  margin-left: 16px;
}
</style>
