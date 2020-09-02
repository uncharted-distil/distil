<template>
  <div class="container-fluid d-flex flex-column h-100">
    <b-modal
      v-model="showGeoModal"
      title="Missing Geocoordinate Features"
      ok-only
      hide-header-close
      @ok="onClose"
    >
      <p>Not enough columns to create a Geocoordinate feature.</p>
      <p>
        Please check the dataset to see if additional columns can be set to a
        Latitude, Longitude or Decimal Type.
      </p>
    </b-modal>
    <b-modal
      v-model="showTimeModal"
      title="Missing Time Series Features"
      ok-only
      hide-header-close
      @ok="onClose"
    >
      <p>Not enough columns to create a Time feature.</p>
      <p>
        Please check the dataset to see if additional columns can be set to a
        Integer, Date/Time or Decimal Type.
      </p>
    </b-modal>
    <b-row class="flex-0-nav"></b-row>

    <b-row class="flex-shrink-0 align-items-center bg-white">
      <b-col v-if="isTimeseries" cols="4" class="offset-md-1">
        <h5 class="header-label">Configure Time Series</h5>
      </b-col>
      <b-col v-if="isGeocoordinate" cols="4" class="offset-md-1">
        <h5 class="header-label">Configure Geocoordinate</h5>
      </b-col>
    </b-row>

    <b-container class="mt-3 h-100">
      <b-row>
        <b-col v-if="isTimeseries" cols="12">
          <p>
            To predict a value over time <strong>(forecasting)</strong> your
            target should be a <strong>timeseries.</strong><br />
            Select a <strong>time</strong> column, a
            <strong>value</strong> column and if available, optionally add one
            or more <strong>series id</strong> column(s) to create multiple
            timeseries.
          </p>
        </b-col>
        <b-col v-if="isGeocoordinate" cols="12">
          <p>
            If your data contains geocoordinate data (<strong
              >latitude, longitude</strong
            >) in separate columns, select those to display the location data on
            a map.
          </p>
        </b-col>
      </b-row>
      <b-row>
        <b-col cols="6" v-if="isTimeseries">
          <template v-if="idCols.length > 0">
            <b-row
              class="mt-1 mb-1"
              v-for="(idCol, index) in idCols"
              :key="idCol.value"
            >
              <b-col cols="5">
                <template
                  v-if="index === 0 && idOptions(idCol.value).length !== 0"
                >
                  <b>Series ID Column(s):</b>
                </template>
              </b-col>

              <b-col
                cols="7"
                class="d-flex align-content-center"
                v-if="idOptions(idCol.value).length !== 0"
              >
                <b-form-select
                  class="mr-auto"
                  v-model="idCol.value"
                  :options="idOptions(idCol.value)"
                  @input="onIdChange"
                />
                <b-button
                  class="ml-1"
                  variant="outline-danger"
                  v-if="idCol.value"
                  title="Clear Selection"
                  @click="removeIdCol(idCol.value)"
                >
                  <i class="fa fa-times-circle"></i>
                </b-button>
              </b-col>
            </b-row>
          </template>

          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b>Time Column:</b>
            </b-col>

            <b-col cols="7">
              <b-form-select
                v-model="xCol"
                :options="xColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>

          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b>Value Column:</b>
            </b-col>

            <b-col cols="7">
              <b-form-select
                v-model="yCol"
                :options="yColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>
        </b-col>
        <b-col cols="6" v-if="isGeocoordinate">
          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b>Longitude Column:</b>
            </b-col>

            <b-col cols="7">
              <b-form-select
                v-model="xCol"
                :options="xColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>

          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b>Latitude Column:</b>
            </b-col>

            <b-col cols="7">
              <b-form-select
                v-model="yCol"
                :options="yColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>
        </b-col>

        <b-col cols="6">
          <div v-if="isReady && previewSummary && previewSummary.baseline">
            <component
              :summary="previewSummary"
              :is="getFacetByType(groupingType)"
            >
            </component>
          </div>
          <div v-else>
            <facet-loading :summary="{ label: 'pending' }"> </facet-loading>
          </div>
        </b-col>
      </b-row>

      <b-row align-h="center">
        <b-btn
          class="mt-3 var-grouping-button"
          variant="outline-secondary"
          :disabled="isPending"
          @click="onClose"
        >
          <div class="row justify-content-center">
            <i class="fa fa-times-circle fa-2x mr-2"></i>
            <b>Cancel</b>
          </div>
        </b-btn>
        <b-btn
          class="mt-3 var-grouping-button"
          variant="primary"
          :disabled="isPending || !isReady"
          @click="onGroup"
        >
          <div class="row justify-content-center">
            <i class="fa fa-check-circle fa-2x mr-2"></i>
            <b>Submit</b>
          </div>
        </b-btn>
      </b-row>

      <b-row class="grouping-progress">
        <b-progress
          v-if="isPending"
          :value="percentComplete"
          variant="secondary"
          striped
          :animated="true"
        ></b-progress>
      </b-row>
    </b-container>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import {
  VariableSummary,
  Variable,
  Grouping,
  TimeseriesGrouping,
  GeoCoordinateGrouping,
} from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";
import {
  INTEGER_TYPE,
  TEXT_TYPE,
  ORDINAL_TYPE,
  TIMESTAMP_TYPE,
  CATEGORICAL_TYPE,
  DATE_TIME_TYPE,
  REAL_TYPE,
  GEOCOORDINATE_TYPE,
  TIMESERIES_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  isLongitudeGroupType,
  isTimeGroupType,
  isLatitudeGroupType,
  isValueGroupType,
} from "../util/types";
import {
  filterSummariesByDataset,
  getComposedVariableKey,
  hasTimeseriesFeatures,
  hasGeoordinateFeatures,
  minimumRouteKey,
} from "../util/data";
import { getFacetByType } from "../util/facets";
import { SELECT_TARGET_ROUTE } from "../store/route/index";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";
import FacetLoading from "../components/facets/FacetLoading.vue";
import FacetTimeseries from "../components/facets/FacetTimeseries.vue";
import GeocoordinateFacet from "../components/facets/GeocoordinateFacet.vue";

export default Vue.extend({
  name: "variable-grouping",

  components: {
    FacetLoading,
    FacetTimeseries,
    GeocoordinateFacet,
  },

  data() {
    return {
      idCols: [{ value: null }],
      prevIdCols: 0,
      xCol: null,
      yCol: null,
      hideIdCol: [false],
      hideXCol: true,
      hideYCol: true,
      hideClusterCol: true,
      other: [],
      isPending: false,
      percentComplete: 100,
      isUpdating: false,
    };
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    groupingType(): string {
      return routeGetters.getGroupingType(this.$store);
    },
    isGeocoordinate(): boolean {
      return this.groupingType === GEOCOORDINATE_TYPE;
    },
    isTimeseries(): boolean {
      return this.groupingType === TIMESERIES_TYPE;
    },
    xColOptions(): Object[] {
      if (!this.isGeocoordinate && !this.isTimeseries) {
        return [];
      }

      const def = {
        value: null,
        text: "",
        disabled: true,
      };
      let xFilterFunction = null;

      if (this.isGeocoordinate) {
        def.text = `Choose ${LONGITUDE_TYPE} column`;
        xFilterFunction = isLongitudeGroupType;
      } else if (this.isTimeseries) {
        def.text = "Choose value column";
        xFilterFunction = isTimeGroupType;
      }

      const suggestions = this.variables
        .filter((v) => xFilterFunction(v.colType))
        .filter((v) => !this.isIDCol(v.colName))
        .filter((v) => !this.isYCol(v.colName))
        .map((v) => {
          return { value: v.colName, text: v.colDisplayName };
        });
      return [].concat([def], suggestions);
    },

    yColOptions(): Object[] {
      if (!this.isGeocoordinate && !this.isTimeseries) {
        return [];
      }

      const def = {
        value: null,
        text: "",
        disabled: true,
      };
      let yFilterFunction = null;

      if (this.isGeocoordinate) {
        def.text = `Choose ${LATITUDE_TYPE} column`;
        yFilterFunction = isLatitudeGroupType;
      } else if (this.isTimeseries) {
        def.text = "Choose value column";
        yFilterFunction = isValueGroupType;
      }

      const suggestions = this.variables
        .filter((v) => yFilterFunction(v.colType))
        .filter((v) => !this.isIDCol(v.colName))
        .filter((v) => !this.isXCol(v.colName))
        .map((v) => {
          return { value: v.colName, text: v.colDisplayName };
        });
      return [].concat(def, suggestions);
    },
    isReady(): boolean {
      const hasBasicFields =
        this.xCol !== null && this.yCol !== null && this.groupingType !== null;
      return hasBasicFields;
    },
    previewSummary(): VariableSummary {
      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );
      const summaryKeys = Object.keys(summaryDictionary);
      const previewKey = summaryKeys.filter(
        (v) => v.indexOf(this.xCol) > -1 && v.indexOf(this.yCol) > -1
      )[0];
      const minKey = minimumRouteKey();
      const pv = summaryDictionary?.[previewKey]?.[minKey];
      return pv;
    },
    showGeoModal: {
      get(): boolean {
        return (
          this.variables &&
          this.isGeocoordinate &&
          !this.isUpdating &&
          !hasGeoordinateFeatures(this.variables)
        );
      },
      set: () => {
        console.info("insufficient geocoordinate variables");
      },
    },
    showTimeModal: {
      get(): boolean {
        return (
          this.variables &&
          this.isTimeseries &&
          !this.isUpdating &&
          !hasTimeseriesFeatures(this.variables)
        );
      },
      set: () => {
        console.info("insufficient timeseries variables");
      },
    },
  },

  beforeMount() {
    console.log("hmm");
    datasetActions.fetchVariables(this.$store, { dataset: this.dataset });
  },
  methods: {
    getFacetByType: getFacetByType,
    async clearGrouping() {
      if (this.previewSummary) {
        await datasetActions.removeGrouping(this.$store, {
          dataset: this.dataset,
          variable: this.previewSummary.key,
        });
      }
    },
    idOptions(idCol): Object[] {
      const ID_COL_TYPES = {
        [TEXT_TYPE]: true,
        [ORDINAL_TYPE]: true,
        [CATEGORICAL_TYPE]: true,
      };
      const suggestions = this.variables
        .filter((v) => ID_COL_TYPES[v.colType])
        .filter((v) => v.colName === idCol || !this.isIDCol(v.colName))
        .map((v) => {
          return { value: v.colName, text: v.colDisplayName };
        });

      if (suggestions.length > 0) {
        const def = [{ value: null, text: "Choose ID", disabled: true }];
        return [].concat(def, suggestions);
      }
      return [];
    },
    onIdChange(arg) {
      const values = this.idCols.map((c) => c.value).filter((v) => v);
      if (values.length === this.prevIdCols) {
        return;
      }
      this.idCols.push({ value: null });
      this.hideIdCol.push(false);
      this.prevIdCols++;
      this.onChange();
    },
    removeIdCol(value) {
      this.idCols = this.idCols.filter((idCol) => idCol.value !== value);
      this.prevIdCols--;
      if (this.isReady) {
        this.submitGrouping(false);
      } else {
        this.clearGrouping();
      }
    },
    isIDCol(arg): boolean {
      return !!this.idCols.find((id) => id.value === arg);
    },
    isXCol(arg): boolean {
      return arg === this.xCol;
    },
    isYCol(arg): boolean {
      return arg === this.yCol;
    },
    isOtherCol(arg): boolean {
      return this.other.indexOf(arg) !== -1;
    },
    onChange() {
      if (this.isReady) {
        this.submitGrouping(false);
      }
    },
    onGroup() {
      this.submitGrouping(true);
    },
    getHiddenCols(idCol) {
      const hiddenCols = [this.xCol, this.yCol];
      if (idCol !== null) {
        hiddenCols.push(idCol);
      }
      return hiddenCols;
    },
    async submitGrouping(gotoTarget: boolean) {
      await this.clearGrouping();
      this.isUpdating = true;
      // Create a list of id values, filtering out the empty entry
      const ids = this.idCols.map((c) => c.value).filter((v) => v);

      // generate the grouping structure that describes how the vars will be grouped
      const idCol = this.isTimeseries ? getComposedVariableKey(ids) : null;

      const hiddenCols = gotoTarget ? this.getHiddenCols(idCol) : [];

      const grouping: Grouping = {
        type: this.groupingType,
        dataset: this.dataset,
        idCol: idCol,
        subIds: ids,
        hidden: hiddenCols,
      };

      if (this.isTimeseries) {
        const tsGrouping = grouping as TimeseriesGrouping;
        tsGrouping.xCol = this.xCol;
        tsGrouping.yCol = this.yCol;
        tsGrouping.clusterCol = null;
      } else if (this.isGeocoordinate) {
        const tsGrouping = grouping as GeoCoordinateGrouping;
        tsGrouping.xCol = this.xCol;
        tsGrouping.yCol = this.yCol;
      }

      await datasetActions.setGrouping(this.$store, {
        dataset: this.dataset,
        grouping: grouping,
      });

      // If this dataset contains multiple timeseries, then we need to request clustering be run on it
      if (this.isTimeseries && ids.length > 0) {
        await datasetActions.fetchClusters(this.$store, {
          dataset: this.dataset,
        });
      }

      if (gotoTarget) {
        this.gotoTargetSelection();
      }

      this.isUpdating = false;
    },
    async onClose() {
      await this.clearGrouping();
      this.gotoTargetSelection();
    },
    gotoTargetSelection() {
      this.$router.go(-1);
    },
  },
});
</script>

<style>
.var-grouping-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
}
.grouping-progress {
  margin: 6px 10%;
}
</style>
