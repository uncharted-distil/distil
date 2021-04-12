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
  <ol class="labels" :class="alignment">
    <li
      v-if="!containsUserAnnotation"
      v-for="(label, index) in labels"
      :key="index"
      :title="label.title"
      class="label"
      :class="label.status"
    >
      {{ label.value }}
    </li>
    <label-annotation
      v-if="containsUserAnnotation"
      :item="item"
      :label-feature-name="labelFeatureName"
    />
  </ol>
</template>

<script lang="ts">
import Vue from "vue";
import { Dictionary } from "../util/dict";
import { TableRow, TableColumn, VariableSummary } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as predictGetters } from "../store/predictions/module";
import LabelAnnotation from "./labelingComponents/LabelAnnotation.vue";
import _ from "lodash";
import { minimumRouteKey } from "../util/data";

interface Label {
  status: string;
  value: string;
  title: string;
}

/**
 * Display the prediction result as a label.
 */
export default Vue.extend({
  name: "image-label",

  components: {
    LabelAnnotation,
  },

  data() {
    return {};
  },

  props: {
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: () => {
        return {} as Dictionary<TableColumn>;
      },
    },
    includedActive: {
      type: Boolean as () => boolean,
      default: false,
    },
    item: Object as () => TableRow,
    shortenLabels: {
      type: Boolean as () => boolean,
      default: true,
    },
    alignHorizontal: {
      type: Boolean as () => boolean,
      default: false,
    },
    labelFeatureName: { type: String, default: "" },
  },

  computed: {
    labels(): Label[] {
      const labels: Label[] = [];
      let status: string;
      if (!this.item) {
        return [];
      }
      for (const key in this.dataFields) {
        status = null;

        // If we're showing error, we want to show
        // a) *just* the correct label or
        // b) the incorrect label *and* the ground truth label for comparison.
        if (this.showError) {
          if (key === this.predictedField) {
            status = this.correct() ? "correct" : "incorrect";
          } else if (key === this.targetField && this.correct()) {
            continue;
          }
        }

        // Display the label
        if (key === this.targetField || key === this.predictedField) {
          const fullLabel = <string>this.item[key].value;
          if (this.shortenLabels) {
            labels.push({
              status,
              value: this.shortenLabel(fullLabel),
              title: fullLabel,
            });
          } else {
            labels.push({
              status,
              value: fullLabel,
              title: "",
            });
          }
        }
      }

      return labels;
    },

    predictedField(): string {
      const predictions = requestGetters.getActivePredictions(this.$store);
      if (predictions) {
        return predictions.predictedKey;
      }

      const solution = requestGetters.getActiveSolution(this.$store);
      return solution ? `${solution.predictedKey}` : "";
    },

    showError(): boolean {
      return (
        this.predictedField && !requestGetters.getActivePredictions(this.$store)
      );
    },

    alignment(): string {
      return this.alignHorizontal ? "horizontal" : "vertical";
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    // Creates a dictionary of label lengths keyed by a label's first character.  This supports
    // the generation of a minimum ambiguous label by starting letter.  For the set [Airport, Agriculture, Forest],
    // we will end up with [A=4, F=1], which at runtime will generate display labels of [Ai, Ag, F].  This is a computed
    // property so it should only end up being updated when the underlying data changes.
    labelLengths(): Dictionary<number> {
      let summary: VariableSummary = null;
      // find the target variable and get the prediction labels
      if (this.showError) {
        summary = resultGetters.getTargetSummary(this.$store);
      } else if (this.predictedField !== "") {
        const dict = predictGetters.getPredictionSummaries(this.$store);
        summary = dict.find((pSum) => {
          return pSum.key === this.predictedField;
        });
      } else {
        const minKey = minimumRouteKey();
        const dict = datasetGetters.getVariableSummariesDictionary(this.$store);
        summary = dict[this.targetField][minKey];
      }
      const bucketNames = summary?.baseline?.buckets.map((b) => b.key);
      // If this isn't categorical, don't generate the table.
      if (!bucketNames) {
        return {};
      }

      // initialize label lengths with zeroes
      const imageLabelLengths: Dictionary<number> = {};
      bucketNames.forEach((b) => (imageLabelLengths[b[0]] = 0));

      // Compare each label to the others
      for (let i = 0; i < bucketNames.length; i++) {
        const currLabel = bucketNames[i];
        for (let j = i + 1; j < bucketNames.length; j++) {
          const compareLabel = bucketNames[j];
          // Update the min number of characters required to disambiguate the labels
          // with the same starting character.
          if (currLabel[0] === compareLabel[0]) {
            for (
              let k = imageLabelLengths[currLabel[0]];
              k < Math.min(currLabel.length, compareLabel.length);
              k++
            ) {
              if (currLabel[k] !== compareLabel[k]) {
                break;
              }
              imageLabelLengths[currLabel[0]] += 1;
            }
          }
        }
      }
      return imageLabelLengths;
    },
    containsUserAnnotation(): boolean {
      if (!this.item) {
        return false;
      }
      return this.item[this.labelFeatureName] !== undefined;
    },
  },

  methods: {
    correct(): boolean {
      return (
        this.item[this.targetField]?.value ===
        this.item[this.predictedField]?.value
      );
    },

    // Given a raw label value, returns shortened label that is unique amongst the set of target labels.
    shortenLabel(rawLabel: string): string {
      const labelLength = this.labelLengths[rawLabel[0]];
      return _.isNil(labelLength)
        ? rawLabel
        : rawLabel.substring(0, labelLength + 1);
    },
  },
});
</script>

<style scoped>
.labels {
  color: #ffffff;
  display: inline-block;
  list-style: none;
  list-style-position: outside;
  max-width: 100%;
  padding-inline-start: 0px;
}

.labels.horizontal {
  display: inline-flex;
}

.label {
  background-color: #424242;
  overflow: hidden;
  padding: 0.1em 0.4em;
}

.label.correct {
  background-color: #03c003;
}

.label.incorrect {
  background-color: #be0000;
}
</style>
