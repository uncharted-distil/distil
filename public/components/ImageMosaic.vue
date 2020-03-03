<template>
  <div class="image-mosaic">
    <template v-for="imageField in imageFields">
      <template v-for="item in items">
        <div class="image-tile">
          <template v-for="(fieldInfo, fieldKey) in fields">
            <image-preview
              v-if="fieldKey === imageField"
              class="image-preview"
              :row="item"
              :image-url="item[fieldKey].value"
              :width="imageWidth"
              :height="imageHeight"
              :on-click="onImageClick"
            ></image-preview>
          </template>
          <div v-if="showError" class="image-label-container">
            <template v-for="(fieldInfo, fieldKey) in fields">
              <div v-if="fieldKey == targetField" class="image-label">
                {{ item[fieldKey].value }}
              </div>
              <div
                v-if="fieldKey == predictedField && correct(item)"
                class="image-label-correct"
              >
                {{ item[fieldKey].value }}
              </div>
              <div
                v-if="fieldKey == predictedField && !correct(item)"
                class="image-label-incorrect"
              >
                {{ item[fieldKey].value }}
              </div>
            </template>
          </div>
          <div v-if="!showError" class="image-label-container">
            <template v-for="(fieldInfo, fieldKey) in fields">
              <div v-if="fieldKey == targetField" class="image-label">
                {{ item[fieldKey].value }}
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
import { getters as solutionGetters } from "../store/solutions/module";
import {
  RowSelection,
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Row
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection
} from "../util/row";
import { IMAGE_TYPE } from "../util/types";
import { Solution } from "../store/solutions";

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

    imageFields(): string[] {
      return _.map(this.fields, (field, key) => {
        return {
          key: key,
          type: field.type
        };
      })
        .filter(field => field.type === IMAGE_TYPE)
        .map(field => field.key);
    },

    solution(): Solution {
      return solutionGetters.getActiveSolution(this.$store);
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    predictedField(): string {
      return this.solution ? `${this.solution.predictedKey}` : "";
    },

    showError(): boolean {
      return this.predictedField !== "";
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
    }
  }
});
</script>

<style>
.image-mosaic {
  display: block;
  overflow: visible;
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
