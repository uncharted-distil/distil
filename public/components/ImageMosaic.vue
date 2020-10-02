<template>
  <div class="mosaic-container">
    <div class="image-mosaic">
      <template v-for="imageField in imageFields">
        <template v-for="item in paginatedItems">
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
    <b-pagination
      v-if="items && items.length > perPage"
      align="center"
      first-number
      last-number
      size="sm"
      v-model="currentPage"
      :per-page="perPage"
      :total-rows="itemCount"
    ></b-pagination>
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
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
} from "../util/row";
import { getImageFields } from "../util/data";

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
      currentPage: 1,
      perPage: 100,
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
    paginatedItems(): TableRow[] {
      const page = this.currentPage - 1; // currentPage starts at 1
      const start = page * this.perPage;
      const end = start + this.perPage;

      return this.items.slice(start, end);
    },
    itemCount(): number {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataLength(this.$store)
        : datasetGetters.getExcludedTableDataLength(this.$store);
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
  margin-bottom: 0.25rem;
}
.mosaic-container {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-box-direction: normal;
  flex-direction: column;
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
