<template>
  <div>
    <b-button @click="onTimeseriesClick" variant="dark">
      <i class="fa fa-area-chart" /> Timeseries
    </b-button>
    <b-button @click="onMapClick" variant="dark">
      <i class="fa fa-globe" /> Map
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

import { GROUPING_ROUTE } from "../../store/route/index";
import { getters as routeGetters } from "../../store/route/module";
import { createRouteEntry } from "../../util/routes";
import { GEOCOORDINATE_TYPE, TIMESERIES_TYPE } from "../../util/types";

export default Vue.extend({
  name: "AddVariablePane",

  methods: {
    groupingClick(type) {
      const entry = createRouteEntry(GROUPING_ROUTE, {
        dataset: routeGetters.getRouteDataset(this.$store),
        groupingType: type,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    onMapClick() {
      this.groupingClick(GEOCOORDINATE_TYPE);
    },

    onTimeseriesClick() {
      this.groupingClick(TIMESERIES_TYPE);
    },
  },
});
</script>
