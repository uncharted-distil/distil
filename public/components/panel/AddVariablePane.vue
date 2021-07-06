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
  <div>
    <b-button variant="dark" @click="onTimeseriesClick">
      <i class="fa fa-area-chart" /> Timeseries
    </b-button>
    <b-button variant="dark" @click="onMapClick">
      <i class="fa fa-globe" /> Map
    </b-button>
    <b-button v-if="enableLabel" variant="dark" @click="onLabelClick">
      <i class="fa fa-tag" /> Label
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

import { GROUPING_ROUTE } from "../../store/route/index";
import { getters as routeGetters } from "../../store/route/module";
import { EventList } from "../../util/events";
import { createRouteEntry } from "../../util/routes";
import { GEOCOORDINATE_TYPE, TIMESERIES_TYPE } from "../../util/types";

export default Vue.extend({
  name: "AddVariablePane",
  props: {
    enableLabel: { type: Boolean as () => boolean, default: false },
  },
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
    onLabelClick() {
      this.$emit(EventList.EXPLORER.SWITCH_TO_LABELING_EVENT);
    },
  },
});
</script>
