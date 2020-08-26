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
              <i class="fa fa-columns"></i>
            </b-form-radio>
            <b-form-radio
              :value="IMAGE_VIEW"
              v-if="hasImageVariables"
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
import { getters as routeGetters } from "../store/route/module";
import {
  TIMESERIES_TYPE,
  IMAGE_TYPE,
  LONGITUDE_TYPE,
  LATITUDE_TYPE,
  GEOCOORDINATE_TYPE,
  REMOTE_SENSING_TYPE,
} from "../util/types";

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
        this.$emit("input", this.internalVal);
      },
    },
    hasImageVariables(): boolean {
      return (
        this.variables.filter(
          (v) => v.colType === IMAGE_TYPE || v.colType === REMOTE_SENSING_TYPE,
        ).length > 0
      );
    },
    hasGraphVariables(): boolean {
      // TODO: add this in
      return false;
    },
    hasGeoVariables(): boolean {
      const hasGeocoord =
        this.variables.filter(
          (v) =>
            v.grouping &&
            [GEOCOORDINATE_TYPE, REMOTE_SENSING_TYPE].includes(v.grouping.type),
        ).length > 0;
      const hasLat =
        this.variables.filter((v) => v.colType === LONGITUDE_TYPE).length > 0;
      const hasLon =
        this.variables.filter((v) => v.colType === LATITUDE_TYPE).length > 0;

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

<style>
.view-button {
  cursor: pointer;
}
.view-button input[type="radio"] {
  display: none;
}
</style>
