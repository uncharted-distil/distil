<template>
  <div class="container-fluid d-flex flex-column h-100">
    <div class="row flex-0-nav"></div>

    <div class="row flex-shrink-0 align-items-center bg-white">
      <div v-if="isTimeseries" class="col-4 offset-md-1">
        <h5 class="header-label">Configure Time Series</h5>
      </div>
      <div v-if="isGeocoordinate" class="col-4 offset-md-1">
        <h5 class="header-label">Configure Geocoordinate</h5>
      </div>
    </div>

    <div class="row justify-content-center h-100 p-3">
      <div class="col-12 col-md-8 flex-column d-flex h-100">
        <div v-if="isTimeseries">
          <div
            class="row mt-1 mb-1"
            v-for="(idCol, index) in idCols"
            :key="idCol.value"
          >
            <div class="col-3">
              <template v-if="index === 0">
                <b>Series ID Column(s):</b>
              </template>
            </div>

            <div class="col-5">
              <b-form-select
                v-model="idCol.value"
                :options="idOptions(idCol.value)"
                @input="onIdChange"
              />
            </div>
          </div>

          <div class="row mt-1 mb-1" v-if="isTimeseries">
            <div class="col-3">
              <b>Time Column:</b>
            </div>

            <div class="col-5">
              <b-form-select v-model="xCol" :options="xColOptions" />
            </div>
          </div>

          <div class="row mt-1 mb-1" v-if="isTimeseries">
            <div class="col-3">
              <b>Value Column:</b>
            </div>

            <div class="col-5">
              <b-form-select v-model="yCol" :options="yColOptions" />
            </div>
          </div>
        </div>

        <div class="row mt-1 mb-1" v-if="isGeocoordinate">
          <div class="col-3">
            <b>Longitude Column:</b>
          </div>

          <div class="col-5">
            <b-form-select v-model="xCol" :options="xColOptions" />
          </div>
        </div>

        <div class="row mt-1 mb-1" v-if="isGeocoordinate">
          <div class="col-3">
            <b>Latitude Column:</b>
          </div>

          <div class="col-5">
            <b-form-select v-model="yCol" :options="yColOptions" />
          </div>
        </div>

        <div v-if="isReady" class="row justify-content-center">
          <b-btn
            class="mt-3 var-grouping-button"
            variant="outline-success"
            :disabled="isPending"
            @click="onGroup"
          >
            <div class="row justify-content-center">
              <i class="fa fa-check-circle fa-2x mr-2"></i>
              <b>Submit</b>
            </div>
          </b-btn>
          <b-btn
            class="mt-3 var-grouping-button"
            variant="outline-danger"
            :disabled="isPending"
            @click="onClose"
          >
            <div class="row justify-content-center">
              <i class="fa fa-times-circle fa-2x mr-2"></i>
              <b>Cancel</b>
            </div>
          </b-btn>
        </div>

        <div class="grouping-progress">
          <b-progress
            v-if="isPending"
            :value="percentComplete"
            variant="outline-secondary"
            striped
            :animated="true"
          ></b-progress>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import { Variable } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions
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
  LONGITUDE_TYPE
} from "../util/types";
import { getComposedVariableKey } from "../util/data";
import { SELECT_TARGET_ROUTE } from "../store/route/index";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";

export default Vue.extend({
  name: "variable-grouping",

  components: {},

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
      percentComplete: 100
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
      if (this.isGeocoordinate) {
        const X_COL_TYPES = {
          [LONGITUDE_TYPE]: true,
          [REAL_TYPE]: true
        };
        const def = [
          {
            value: null,
            text: `Choose ${LONGITUDE_TYPE} column`,
            disabled: true
          }
        ];
        const suggestions = this.variables
          .filter(v => X_COL_TYPES[v.colType])
          .filter(v => !this.isIDCol(v.colName))
          .filter(v => !this.isYCol(v.colName))
          .map(v => {
            return { value: v.colName, text: v.colDisplayName };
          });

        if (suggestions.length === 1) {
          this.xCol = suggestions[0].value;
          return suggestions;
        }
        return [].concat(def, suggestions);
      } else if (this.isTimeseries) {
        const X_COL_TYPES = {
          [INTEGER_TYPE]: true,
          [DATE_TIME_TYPE]: true,
          [TIMESTAMP_TYPE]: true
        };
        const def = [{ value: null, text: "Choose column", disabled: true }];

        const suggestions = this.variables
          .filter(v => X_COL_TYPES[v.colType])
          .filter(v => !this.isIDCol(v.colName))
          .filter(v => !this.isYCol(v.colName))
          .map(v => {
            return { value: v.colName, text: v.colDisplayName };
          });

        if (suggestions.length === 1) {
          this.xCol = suggestions[0].value;
          return suggestions;
        }

        return [].concat(def, suggestions);
      }
    },

    yColOptions(): Object[] {
      if (this.isGeocoordinate) {
        const Y_COL_TYPES = {
          [LATITUDE_TYPE]: true,
          [REAL_TYPE]: true
        };
        const def = [
          {
            value: null,
            text: `Choose ${LATITUDE_TYPE} column`,
            disabled: true
          }
        ];

        const suggestions = this.variables
          .filter(v => Y_COL_TYPES[v.colType])
          .filter(v => !this.isIDCol(v.colName))
          .filter(v => !this.isXCol(v.colName))
          .map(v => {
            return { value: v.colName, text: v.colDisplayName };
          });

        if (suggestions.length === 1) {
          this.yCol = suggestions[0].value;
          return suggestions;
        }

        return [].concat(def, suggestions);
      } else if (this.isTimeseries) {
        const Y_COL_TYPES = {
          [INTEGER_TYPE]: true,
          [REAL_TYPE]: true
        };
        const def = [{ value: null, text: "Choose column", disabled: true }];

        const suggestions = this.variables
          .filter(v => Y_COL_TYPES[v.colType])
          .filter(v => !this.isIDCol(v.colName))
          .filter(v => !this.isXCol(v.colName))
          .map(v => {
            return { value: v.colName, text: v.colDisplayName };
          });

        if (suggestions.length === 1) {
          this.yCol = suggestions[0].value;
          return suggestions;
        }

        return [].concat(def, suggestions);
      }
    },

    isReady(): boolean {
      return this.xCol !== null && this.groupingType !== null;
      // return this.idCols.length > 1 && this.xCol && this.yCol && this.groupingType;
    }
  },

  beforeMount() {
    viewActions.fetchSelectTargetData(this.$store, false);
  },
  methods: {
    idOptions(idCol): Object[] {
      const ID_COL_TYPES = {
        [TEXT_TYPE]: true,
        [ORDINAL_TYPE]: true,
        [CATEGORICAL_TYPE]: true
      };
      const def = [{ value: null, text: "Choose ID", disabled: true }];
      const suggestions = this.variables
        .filter(v => ID_COL_TYPES[v.colType])
        .filter(v => v.colName === idCol || !this.isIDCol(v.colName))
        .filter(v => !this.isXCol(v.colName))
        .filter(v => !this.isYCol(v.colName))
        .map(v => {
          return { value: v.colName, text: v.colDisplayName };
        });

      return [].concat(def, suggestions);
    },
    onIdChange(arg) {
      const values = this.idCols.map(c => c.value).filter(v => v);
      if (values.length === this.prevIdCols) {
        return;
      }
      this.idCols.push({ value: null });
      this.hideIdCol.push(false);
      this.prevIdCols++;
    },
    isIDCol(arg): boolean {
      return !!this.idCols.find(id => id.value === arg);
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
    onGroup() {
      this.submitGrouping();
    },
    submitGrouping() {
      // create the list of IDs that we're going to be grouping
      const hidden = {
        [this.xCol]: true,
        [this.yCol]: true
      };

      const ids = this.idCols.map(c => c.value).filter(v => v);
      ids.forEach((id, index) => {
        hidden[id] = this.hideIdCol[index];
      });

      // generate the grouping structure that describes how the vars will be grouped
      const grouping = {
        type: this.groupingType,
        dataset: this.dataset,
        idCol: getComposedVariableKey(ids),
        subIds: ids,
        hidden: Object.keys(hidden).filter(v => hidden[v]),
        properties: {
          xCol: this.xCol,
          yCol: this.yCol
        }
      };

      datasetActions
        .setGrouping(this.$store, {
          dataset: this.dataset,
          grouping: grouping
        })
        .then(() => {
          // If this is a timeseries, then we need to request clustering be run on it
          if (this.isTimeseries) {
            datasetActions.clusterData(this.$store, {
              dataset: this.dataset,
              variable: getComposedVariableKey(ids)
            });
          }
          this.gotoTargetSelection();
        });
    },
    onClose() {
      this.gotoTargetSelection();
    },
    gotoTargetSelection() {
      this.$router.go(-1);
    }
  }
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
