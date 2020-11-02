<template>
  <div class="filter-badge" :class="{ active: activeFilter }">
    {{ name }} {{ content }}
    <b-button class="remove-button" size="sm" @click="onClick">
      <i class="fa fa-times" />
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from "moment";
import {
  removeFilterFromRoute,
  Filter,
  NUMERICAL_FILTER,
  DATETIME_FILTER,
  GEOBOUNDS_FILTER,
  CATEGORICAL_FILTER,
  CLUSTER_FILTER,
  TEXT_FILTER,
} from "../util/filters";
import { clearHighlight } from "../util/highlights";
import { removeClusterPrefix } from "../util/types";

export default Vue.extend({
  name: "filter-badge",

  props: {
    filter: Object as () => Filter,
    activeFilter: Boolean as () => boolean,
  },

  computed: {
    name(): string {
      return this.filter.displayName;
    },

    content(): string {
      if (this.filter.type === NUMERICAL_FILTER) {
        const min = this.filter.min.toFixed(2);
        const max = this.filter.max.toFixed(2);
        return `${min} : ${max}`;
      } else if (this.filter.type === DATETIME_FILTER) {
        const min = this.formatDate(this.filter.min * 1000);
        const max = this.formatDate(this.filter.max * 1000);
        return `${min} : ${max}`;
      } else if (this.filter.type === GEOBOUNDS_FILTER) {
        const minX = this.filter.minX.toFixed(2);
        const minY = this.filter.minY.toFixed(2);
        const maxX = this.filter.maxX.toFixed(2);
        const maxY = this.filter.maxY.toFixed(2);
        return `[${minX}, ${minY}] to [${maxX}, ${maxY}]`;
      } else if (
        [CATEGORICAL_FILTER, CLUSTER_FILTER, TEXT_FILTER].includes(
          this.filter.type
        )
      ) {
        return this.filter.categories.join(",");
      }
    },
  },

  methods: {
    onClick() {
      if (!this.activeFilter) {
        removeFilterFromRoute(this.$router, this.filter);
      } else {
        clearHighlight(this.$router);
      }
    },

    formatDate(epochTime: number): string {
      return moment(epochTime).format("YYYY/MM/DD");
    },
  },
});
</script>

<style scoped>
.filter-badge {
  position: relative;
  height: 28px;
  display: inline-block;
  color: #fff;
  padding-left: 8px;
  border-radius: 4px;
  background-color: #333;
}

.filter-badge.active {
  background-color: #255dcc;
}

.remove-button {
  color: #fff;
  margin-left: 8px;
  background: none;
  border-radius: 0px;
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
  border: none;
  border-left: 1px solid #fff;
}

.remove-button:hover {
  color: #fff;
  background-color: #3d70d3;
  border: none;
  border-left: 1px solid #fff;
}

.active .remove-button:hover {
  background-color: #3d70d3;
}
</style>
