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

    <!-- Spacer for the App.vue <navigation> component -->
    <div class="row flex-0-nav" />

    <!-- Title of the page -->
    <header class="header row justify-content-center">
      <b-col cols="12" md="10">
        <h5 class="header-title">
          Configure
          <span v-if="isTimeseries">Time Series</span>
          <span v-else-if="isLabeling">Labeling</span>
          <span v-else>Geocoordinate</span>
        </h5>
      </b-col>
    </header>

    <!-- Information -->
    <section class="sub-header row justify-content-center">
      <b-col cols="12" md="10">
        <template v-if="isTimeseries">
          To predict a value over time <strong>(forecasting)</strong> your
          target should be a <strong>timeseries.</strong><br />
          Select a <strong>time</strong> column, a <strong>value</strong> column
          and if available, optionally add one or more
          <strong>series id</strong> column(s) to create multiple timeseries.
        </template>
        <template v-if="isLabeling">
          To create labels for your data please annotate images with the
          following criteria: if an image is your label give it a
          <strong>positive</strong> annotation, if the image is not your label
          give it a <strong>negative</strong> annotation. For the best results
          make sure to include <strong>positive</strong> and
          <strong>negative</strong> annotations.
        </template>
        <template v-else>
          If your data contains geocoordinate data (<strong
            >latitude, longitude</strong
          >) in separate columns, select those to display the location data on a
          map.
        </template>
      </b-col>
    </section>
    <section v-if="isLabeling" class="h-100">
      <labeling-view />
    </section>
    <!-- Form -->
    <section v-if="!isLabeling" class="mt-3 container">
      <b-row>
        <b-col cols="6">
          <!-- X column -->
          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b v-if="isTimeseries">Time Column:</b>
              <b v-else>Longitude Column:</b>
            </b-col>
            <b-col cols="7">
              <b-form-select
                v-model="xCol"
                :options="xColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>

          <!-- Y column -->
          <b-row class="mt-1 mb-1">
            <b-col cols="5">
              <b v-if="isTimeseries">Value Column:</b>
              <b v-else>Latitude Column:</b>
            </b-col>
            <b-col cols="7">
              <b-form-select
                v-model="yCol"
                :options="yColOptions"
                @input="onChange"
              />
            </b-col>
          </b-row>

          <!-- ID columns -->
          <template v-if="isTimeseries && idCols.length > 0">
            <b-row
              v-for="(idCol, index) in idCols"
              :key="idCol.value"
              class="mt-1 mb-1"
            >
              <b-col cols="5">
                <template
                  v-if="index === 0 && idOptions(idCol.value).length !== 0"
                >
                  <b>Series ID Column(s):</b>
                </template>
              </b-col>

              <b-col
                v-if="idOptions(idCol.value).length !== 0"
                cols="7"
                class="d-flex align-content-center"
              >
                <b-form-select
                  v-model="idCol.value"
                  class="mr-auto"
                  :options="idOptions(idCol.value)"
                  @input="onIdChange"
                />
                <b-button
                  v-if="idCol.value"
                  class="ml-1"
                  variant="outline-danger"
                  title="Clear Selection"
                  @click="removeIdCol(idCol.value)"
                >
                  <i class="fa fa-times-circle" />
                </b-button>
              </b-col>
            </b-row>
          </template>
        </b-col>

        <!-- Facet Preview -->
        <b-col cols="6">
          <div v-if="isReady && previewSummary && previewSummary.baseline">
            <component
              :is="getFacetByType(groupingType)"
              :summary="previewSummary"
            />
          </div>
          <div v-else>
            <facet-loading :summary="{ label: 'pending' }" />
          </div>
        </b-col>
      </b-row>

      <!-- Buttons -->
      <b-row align-h="center">
        <b-btn
          class="mt-3 grouping-button"
          variant="outline-secondary"
          :disabled="isPending"
          @click="onClose"
        >
          <div class="row justify-content-center">
            <i class="fa fa-times-circle fa-2x mr-2" />
            <b>Cancel</b>
          </div>
        </b-btn>
        <b-btn
          class="mt-3 grouping-button"
          variant="primary"
          :disabled="isPending || !isReady"
          @click="onGroup"
        >
          <div class="row justify-content-center">
            <i class="fa fa-check-circle fa-2x mr-2" />
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
        />
      </b-row>
    </section>
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
  LABELING_TYPE,
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
import LabelingView from "../views/Labeling.vue";

export default Vue.extend({
  name: "variable-grouping",

  components: {
    FacetLoading,
    FacetTimeseries,
    GeocoordinateFacet,
    LabelingView,
  },

  data() {
    return {
      idCols: [{ value: null as string }],
      prevIdCols: 0,
      xCol: null as string,
      yCol: null as string,
      hideIdCol: [false],
      hideXCol: true,
      hideYCol: true,
      hideClusterCol: true,
      other: [] as string[],
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
    isLabeling(): boolean {
      return this.groupingType === LABELING_TYPE;
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
        .filter((v) => !this.isIDCol(v.key))
        .filter((v) => !this.isYCol(v.key))
        .map((v) => {
          return { value: v.key, text: v.colDisplayName };
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
        .filter((v) => !this.isIDCol(v.key))
        .filter((v) => !this.isXCol(v.key))
        .map((v) => {
          return { value: v.key, text: v.colDisplayName };
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
      return summaryDictionary?.[previewKey]?.[minKey];
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
    idOptions(idCol: string): Object[] {
      const ID_COL_TYPES = {
        [TEXT_TYPE]: true,
        [ORDINAL_TYPE]: true,
        [CATEGORICAL_TYPE]: true,
      };
      const suggestions = this.variables
        .filter((v) => ID_COL_TYPES[v.colType])
        .filter((v) => v.key === idCol || !this.isIDCol(v.key))
        .map((v) => {
          return { value: v.key, text: v.colDisplayName };
        });

      if (suggestions.length > 0) {
        const def = [{ value: null, text: "Choose ID", disabled: true }];
        return [].concat(def, suggestions);
      }
      return [];
    },

    onIdChange(arg: string) {
      const values = this.idCols.map((c) => c.value).filter((v) => v);
      if (values.length === this.prevIdCols) {
        return;
      }
      this.idCols.push({ value: null });
      this.hideIdCol.push(false);
      this.prevIdCols++;
      this.onChange();
    },

    removeIdCol(value: string) {
      this.idCols = this.idCols.filter((idCol) => idCol.value !== value);
      this.prevIdCols--;
      if (this.isReady) {
        this.submitGrouping(false);
      } else {
        this.clearGrouping();
      }
    },

    isIDCol(arg: string): boolean {
      return !!this.idCols.find((id) => id.value === arg);
    },

    isXCol(arg: string): boolean {
      return arg === this.xCol;
    },

    isYCol(arg: string): boolean {
      return arg === this.yCol;
    },

    isOtherCol(arg: string): boolean {
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

    getHiddenCols(idCol: string, xCol: string, yCol: string): string[] {
      const hiddenCols = [xCol, yCol];
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

      const hidden = gotoTarget
        ? this.getHiddenCols(idCol, this.xCol, this.yCol)
        : [];
      const grouping: Grouping = {
        type: this.groupingType,
        dataset: this.dataset,
        idCol: idCol,
        subIds: ids,
        hidden: hidden,
      };

      const groupings = [] as Grouping[];

      if (this.isTimeseries) {
        const tsGrouping = _.cloneDeep(grouping) as TimeseriesGrouping;
        tsGrouping.xCol = this.xCol;
        tsGrouping.yCol = this.yCol;
        tsGrouping.clusterCol = null;
        groupings.push(tsGrouping);

        if (gotoTarget) {
          // We want to take any other numeric values and create groups for them as well.
          const Y_COL_TYPES = {
            [INTEGER_TYPE]: true,
            [REAL_TYPE]: true,
          };
          const yCols = this.variables
            .filter((v) => Y_COL_TYPES[v.colType])
            .filter((v) => !this.isIDCol(v.colName))
            .filter((v) => !this.isXCol(v.colName))
            .filter((v) => !this.isYCol(v.colName))
            .map((v) => v.colName)
            .forEach((v) => {
              // create a new grouping entry for each value variable and add it to
              // the list to create
              const tsGrouping = _.cloneDeep(grouping) as TimeseriesGrouping;
              tsGrouping.hidden = gotoTarget
                ? this.getHiddenCols(idCol, this.xCol, v)
                : [];
              tsGrouping.xCol = this.xCol;
              tsGrouping.yCol = v;
              tsGrouping.clusterCol = null;
              tsGrouping.hidden = this.getHiddenCols(idCol, this.xCol, v);
              groupings.push(tsGrouping);
            });
        }
      } else if (this.isGeocoordinate) {
        const geoGrouping = grouping as GeoCoordinateGrouping;
        geoGrouping.xCol = this.xCol;
        geoGrouping.yCol = this.yCol;

        groupings.push(geoGrouping);
      }

      // Create all of the necessary groupings
      for (const grouping of groupings) {
        // CDB: This needs to be converted into an API call that can handle creation of
        // multiple groups because the UI goes spastic updating after each individual operation.
        await datasetActions.setGrouping(this.$store, {
          dataset: this.dataset,
          grouping: grouping,
        });
      }

      if (gotoTarget) {
        // If this dataset contains multiple timeseries,
        // and we're doing the final submit,
        // then we need to request clustering be run on it.
        if (this.isTimeseries && ids.length > 0) {
          await datasetActions.fetchClusters(this.$store, {
            dataset: this.dataset,
          });
        }
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

<style scoped>
.grouping-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
}

.grouping-progress {
  margin: 6px 10%;
}
</style>
