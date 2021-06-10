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
  <div class="font-weight-bold" v-bind:class="{ 'nav-link': !hasTabs }">
    <b-nav :tabs="hasTabs">
      <slot></slot>
      <template>
        <b-form-group class="view-button">
          <b-form-radio-group
            buttons
            v-model="content"
            button-variant="outline-secondary"
          >
            <b-form-radio :value="TABLE_VIEW" class="view-button">
              <i class="fa fa-columns" />
            </b-form-radio>
            <b-form-radio
              :value="IMAGE_VIEW"
              v-if="hasImageVariables"
              :disabled="!hasImageVariables"
              class="view-button"
            >
              <i class="fa fa-image"></i>
            </b-form-radio>
            <b-form-radio
              :value="GRAPH_VIEW"
              v-if="hasGraphVariables"
              class="view-button"
            >
              <i class="fa fa-share-alt"></i>
            </b-form-radio>
            <b-form-radio
              :value="GEO_VIEW"
              v-if="hasGeoVariables"
              :disabled="!hasGeoVariables"
              class="view-button"
            >
              <i class="fa fa-globe"></i>
            </b-form-radio>
            <b-form-radio
              :value="TIMESERIES_VIEW"
              v-if="hasTimeseriesVariables"
              class="view-button"
            >
              <i class="fa fa-line-chart"></i>
            </b-form-radio>
          </b-form-radio-group>
        </b-form-group>
      </template>
    </b-nav>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Variable } from "../store/dataset/index";
import {
  IMAGE_TYPE,
  LONGITUDE_TYPE,
  LATITUDE_TYPE,
  GEOCOORDINATE_TYPE,
  MULTIBAND_IMAGE_TYPE,
  GEOBOUNDS_TYPE,
} from "../util/types";
import { EventList } from "../util/events";

const TABLE_VIEW = "table";
const IMAGE_VIEW = "image";
const GRAPH_VIEW = "graph";
const GEO_VIEW = "geo";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "view-type-toggle",

  props: {
    value: String as () => string,
    variables: Array as () => Variable[],
    hasTabs: {
      type: Boolean as () => boolean,
      default: false,
    },
    availableVariables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
    isSelectView: { type: Boolean as () => boolean, default: false },
  },

  data() {
    return {
      internalVal: this.value,
      TABLE_VIEW: TABLE_VIEW,
      IMAGE_VIEW: IMAGE_VIEW,
      GRAPH_VIEW: GRAPH_VIEW,
      GEO_VIEW: GEO_VIEW,
      TIMESERIES_VIEW: TIMESERIES_VIEW,
    };
  },

  computed: {
    content: {
      get(): string {
        return this.internalVal;
      },
      set(value: string) {
        this.internalVal = value;
        this.$emit(EventList.BASIC.INPUT_EVENT, this.internalVal);
      },
    },
    hasImageVariables(): boolean {
      if (this.isSelectView) {
        return this.hasAvailableImageVariables;
      }
      return (
        this.variables.filter(
          (v) => v.colType === IMAGE_TYPE || v.colType === MULTIBAND_IMAGE_TYPE
        ).length > 0
      );
    },
    hasAvailableImageVariables(): boolean {
      return (
        this.availableVariables.filter(
          (v) => v.colType === IMAGE_TYPE || v.colType === MULTIBAND_IMAGE_TYPE
        ).length > 0
      );
    },
    hasGraphVariables(): boolean {
      // TODO: add this in
      return false;
    },
    hasGeoVariables(): boolean {
      if (this.isSelectView) {
        return this.hasAvailableGeoVariables;
      }
      const hasGeocoord = this.variables.some(
        (v) =>
          v.grouping &&
          [GEOCOORDINATE_TYPE, MULTIBAND_IMAGE_TYPE].includes(v.grouping.type)
      );
      const hasLat = this.variables.some((v) => v.colType === LONGITUDE_TYPE);
      const hasLon = this.variables.some((v) => v.colType === LATITUDE_TYPE);

      return (hasLat && hasLon) || hasGeocoord;
    },
    hasAvailableGeoVariables(): boolean {
      const hasGeocoord = this.availableVariables.some(
        (v) =>
          (v.grouping &&
            [GEOCOORDINATE_TYPE, GEOBOUNDS_TYPE].includes(v.grouping.type)) ||
          v.colType === GEOBOUNDS_TYPE
      );
      const hasLat = this.availableVariables.some(
        (v) => v.colType === LONGITUDE_TYPE
      );
      const hasLon = this.availableVariables.some(
        (v) => v.colType === LATITUDE_TYPE
      );
      return (hasLat && hasLon) || hasGeocoord;
    },

    /*
      TODO - Reimplement test once the Timeseries view works again.
      See https://github.com/uncharted-distil/distil/issues/1690
    */
    hasTimeseriesVariables(): boolean {
      return false;
      /*
      return (
        this.variables.filter(
          v => v.grouping && v.grouping.type === TIMESERIES_TYPE
        ).length > 0
      );
      */
    },
  },
});
</script>

<style scoped>
.view-button {
  cursor: pointer;
}
.view-button input[type="radio"] {
  display: none;
}
</style>
