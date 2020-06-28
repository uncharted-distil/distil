<template>
  <div class="image-mosaic">
    <template v-for="imageField in imageFields">
      <template v-for="item in items">
        <div class="image-tile">
          <template v-for="(fieldInfo, fieldKey) in fields">
            <image-preview
              v-if="fieldKey === imageField.key"
              class="image-preview"
              :row="item"
              :image-url="item[fieldKey].value"
              :width="imageWidth"
              :height="imageHeight"
              :on-click="onImageClick"
              :type="imageField.type"
            ></image-preview>
          </template>
          <div v-if="showError" class="image-label-container">
            <template v-for="(fieldInfo, fieldKey) in fields">
              <div
                v-if="fieldKey == predictedField && correct(item)"
                class="image-label-correct"
              >
                {{ shortenLabel(item[fieldKey].value) }}
              </div>
              <div
                v-if="fieldKey == targetField && !correct(item)"
                class="image-label"
              >
                {{ shortenLabel(item[fieldKey].value) }}
              </div>
              <div
                v-if="fieldKey == predictedField && !correct(item)"
                class="image-label-incorrect"
              >
                {{ shortenLabel(item[fieldKey].value) }}
              </div>
            </template>
          </div>
          <div v-if="!showError" class="image-label-container">
            <template v-for="(fieldInfo, fieldKey) in fields">
              <div
                v-if="fieldKey == targetField || fieldKey == predictedField"
                class="image-label"
              >
                {{ shortenLabel(item[fieldKey].value) }}
              </div>
            </template>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ImagePreview from "./ImagePreview";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import {
  RowSelection,
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Row,
  Variable,
  VariableSummary
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as resultGetters } from "../store/results/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection
} from "../util/row";
import { getImageFields } from "../util/data";
import { Solution } from "../store/requests/index";
import { keys } from "d3";
import { min } from "moment";

export default Vue.extend({
  name: "image-mosaic",

  components: {
    ImagePreview
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>
  },

  data() {
    return {
      imageWidth: 128,
      imageHeight: 128
    };
  },

  computed: {
    items(): TableRow[] {
      if (this.dataItems) {
        return this.dataItems;
      }
      const items = this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },

    fields(): Dictionary<TableColumn> {
      const currentFields = this.dataFields
        ? this.dataFields
        : this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
      return currentFields;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
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

    // Creates a dictionary of label lengths keyed by a label's first character.  This supports
    // the generation of a minimum ambiguous label by starting letter.  For the set [Airport, Agriculture, Forest],
    // we will end up with [A=4, F=1], which at runtime will generate display lables of [Ai, Ag, F].  This is a computed
    // property so it should only end up being updated when the underlying data changes.
    labelLengths(): Dictionary<number> {
      let summary: VariableSummary = null;
      // find the target variable and get the prediction labels
      if (this.showError) {
        summary = resultGetters.getTargetSummary(this.$store);
      } else {
        summary = datasetGetters
          .getVariableSummaries(this.$store)
          .find(v => v.key === this.targetField);
      }
      const bucketNames = summary.baseline.buckets.map(b => b.key);
      // If this isn't categorical, don't generate the table.
      if (!bucketNames) {
        return {};
      }

      // initailize label lengths with zeroes
      const imageLabelLengths: Dictionary<number> = {};
      bucketNames.forEach(b => (imageLabelLengths[b[0]] = 0));

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
    }
  },

  methods: {
    onImageClick(event: any) {
      if (!isRowSelected(this.rowSelection, event.row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      }
    },

    correct(item: any): boolean {
      return item[this.targetField].value === item[this.predictedField].value;
    },

    // Given a raw label value, returns shortened label that is unique amongst the set of target labels.
    shortenLabel(rawLabel: string): string {
      const labelLength = this.labelLengths[rawLabel[0]];
      return _.isNil(labelLength)
        ? rawLabel
        : rawLabel.substring(0, labelLength + 1);
    }
  }
});
</script>

<style>
.image-mosaic {
  display: block;
  overflow: auto;
  padding-bottom: 0.5rem; /* To add some spacing on overflow. */
  height: 100%;
  width: 100%;
}

.image-tile {
  display: inline-block;
  position: relative;
  vertical-align: bottom;
  margin: 2px;
}

.image-preview {
  position: relative;
}

.image-label-container {
  position: absolute;
  top: 2px;
  left: 2px;
  z-index: 1;
}

.image-label {
  float: right;
  background-color: #424242;
  color: #fff;
  padding: 0 4px;
  margin: 0, 2px;
}

.image-label-correct {
  float: right;
  background-color: #03c003;
  color: #fff;
  padding: 0 4px;
  margin: 0, 2px;
}

.image-label-incorrect {
  float: right;
  background-color: #be0000;
  color: #fff;
  padding: 0 4px;
  margin: 0, 2px;
}
</style>
