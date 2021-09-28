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
    <div class="p-1">
      <b-button
        :disabled="isLoading || !minimumRequirementsMet"
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
        <template v-if="isSaving">
          <div v-html="spinnerHTML" />
        </template>
        <template v-else> Save</template>
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
import { EventList } from "../../util/events";

export default Vue.extend({
  name: "create-labeling-form",
  props: {
    isLoading: { type: Boolean as () => boolean, default: false },
    lowShotSummary: Object as () => VariableSummary,
    labelFeatureName: { type: String, default: "" },
    isSaving: { type: Boolean as () => boolean, default: false },
  },
  computed: {
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    saveEvent(): string {
      return EventList.LABEL.OPEN_SAVE_MODAL_EVENT;
    },
    applyEvent(): string {
      return EventList.LABEL.APPLY_EVENT;
    },
    exportEvent(): string {
      return EventList.LABEL.EXPORT_EVENT;
    },
    annotationHasChanged(): boolean {
      return routeGetters.getAnnotationHasChanged(this.$store);
    },
    minimumRequirementsMet(): boolean {
      const keys = new Map(
        this.lowShotSummary?.baseline?.buckets.map((b) => [b.key, b.count])
      );
      if (!keys) {
        return false;
      }
      return (
        keys.has(LowShotLabels.positive) && keys.get(LowShotLabels.positive) > 0
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
    onEvent(event: string) {
      this.$eventBus.$emit(event);
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
