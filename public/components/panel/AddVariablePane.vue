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
    <b-button variant="dark" @click="onLabelClick">
      <i class="fa fa-tag" /> Label
    </b-button>
    <b-modal
      :id="modalId"
      @hide="onLabelSubmit"
      no-close-on-backdrop
      ok-only
      no-close-on-esc
    >
      <template #modal-header>
        {{ labelModalTitle }}
      </template>
      <b-form-group
        v-if="!isClone"
        id="input-group-1"
        label="Label name:"
        label-for="label-input-field"
        description="Enter the name of label."
      >
        <b-form-input
          id="label-input-field"
          v-model="labelName"
          type="text"
          required
          :placeholder="labelName"
        />
      </b-form-group>
      <b-form-group
        v-else
        label="Label name:"
        label-for="label-select-field"
        description="Select the label field."
      >
        <b-form-select
          id="label-select-field"
          v-model="labelName"
          :options="options"
        />
      </b-form-group>
    </b-modal>
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
    onLabelClick() {
      return;
    },
  },
});
</script>
