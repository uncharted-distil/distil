<template>
  <div class="filter-badge" :class="{ active: isActive }">
    <span :title="title">{{ name }} {{ content }}</span>
    <b-button
      v-if="isHighlight"
      size="sm"
      @click="onAdd"
      title="Add highlight as a filter"
    >
      <i class="fa fa-plus" />
    </b-button>
    <b-button size="sm" @click="onRemove" title="Remove filter">
      <i class="fa fa-times" />
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from "moment";
import {
  addFilterToRoute,
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

export default Vue.extend({
  name: "filter-badge",

  props: {
    filter: Object as () => Filter,
    activeFilter: Boolean,
    isHighlight: Boolean,
  },

  computed: {
    isActive(): boolean {
      return this.activeFilter || this.isHighlight;
    },

    name(): string {
      return this.filter.displayName;
    },

    title(): string {
      return this.isActive
        ? `Highlighted data filtered by ${this.name}`
        : `Data filtered by ${this.name}`;
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
    onAdd(): void {
      addFilterToRoute(this.$router, this.filter);
      clearHighlight(this.$router);
    },

    onRemove(): void {
      if (this.isActive) {
        clearHighlight(this.$router);
      } else {
        removeFilterFromRoute(this.$router, this.filter);
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
  --height: 28px;
  background-color: var(--gray-900);
  border-radius: 2px;
  color: var(--white);
  display: inline-flex;
  height: var(--height);
  margin: 0.2rem;
  overflow: hidden; /* hide issue with button:hover styling */
  position: relative;
}

.active {
  background-color: var(--blue);
}

span {
  line-height: var(--height);
  padding: 0 0.5em;
}

.btn {
  color: #fff;
  background: none;
  border: none;
  border-left: 1px solid var(--white);
  border-radius: 0;
}

.btn:hover {
  color: var(--white);
  background-color: var(--blue);
  border-left: 1px solid var(--white);
}

.active .btn:hover {
  background-color: #3d70d3;
}
</style>
