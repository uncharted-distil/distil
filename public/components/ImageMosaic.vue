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
              :key="fieldKey"
            ></image-preview>
          </template>
          <image-label
            class="image-label"
            :dataFields="dataFields"
            includedActive
            shortenLabels
            alignHorizontal
            :item="item"
          />
        </div>
      </template>
    </template>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ImageLabel from "./ImageLabel";
import ImagePreview from "./ImagePreview";
import {
  RowSelection,
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Row,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as requestGetters } from "../store/requests/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
} from "../util/row";
import { getImageFields } from "../util/data";
import { Solution } from "../store/requests/index";
import { keys } from "d3";
import { min } from "moment";

export default Vue.extend({
  name: "image-mosaic",

  components: {
    ImageLabel,
    ImagePreview,
  },

  props: {
    instanceName: String as () => string,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
  },

  data() {
    return {
      imageWidth: 128,
      imageHeight: 128,
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
        this.instanceName,
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

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },
  },

  methods: {
    onImageClick(event: any) {
      if (!isRowSelected(this.rowSelection, event.row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD],
        );
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD],
        );
      }
    },
  },
});
</script>

<style scoped>
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

.image-label {
  position: absolute;
  left: 2px;
  top: 2px;
  z-index: 1;
}
</style>
