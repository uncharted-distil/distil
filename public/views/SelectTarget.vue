<template>
  <div class="container-fluid d-flex flex-column h-100">
    <!-- Spacer for the App.vue <navigation> component -->
    <div class="row flex-0-nav"></div>

    <!-- Title of the page -->
    <header class="header row justify-content-center">
      <b-col cols="12" md="10">
        <h5 class="header-title">
          Dataset Overview: Select Feature to Predict
        </h5>
      </b-col>
    </header>

    <!-- Information -->
    <section class="sub-header row justify-content-center">
      <b-col cols="12" md="10">
        <b-row no-gutters>
          <b-col cols="12" md="7" class="mr-auto">
            <h6 class="sub-header-title">
              Select feature to infer below (target).
            </h6>
            If you want to predict a value over time, create
            a&nbsp;<strong>Timeseries</strong>. If you have geospatial data, you
            can plot it on a&nbsp;<strong>Map</strong>.
          </b-col>
          <span class="sub-header-action">
            <b-button @click="onTimeseriesClick" variant="dark">
              <i class="fa fa-area-chart"></i> Timeseries
            </b-button>
            <b-button @click="onMapClick" variant="dark">
              <i class="fa fa-globe"></i> Map
            </b-button>
          </span>
        </b-row>
      </b-col>
    </section>

    <!-- List of features -->
    <section class="available-target row justify-content-center">
      <div class="available-target-variables col-12 col-md-10">
        <variable-facets
          enable-search
          enable-type-change
          ignore-highlights
          enable-typefiltering
          :instance-name="instanceName"
          :rows-per-page="numRowsPerPage"
          :summaries="summaries"
          :html="html"
          :logActivity="problemDefinition"
        />
      </div>
    </section>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "../components/facets/VariableFacets.vue";
import { actions as viewActions } from "../store/view/module";
import { actions as appActions } from "../store/app/module";
import { Variable, VariableSummary, SummaryMode } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import {
  AVAILABLE_TARGET_VARS_INSTANCE,
  GROUPING_ROUTE,
  SELECT_TRAINING_ROUTE,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { filterSummariesByDataset, NUM_PER_TARGET_PAGE } from "../util/data";
import { Group } from "../util/facets";
import { createRouteEntry, varModesToString } from "../util/routes";
import {
  isUnsupportedTargetVar,
  GEOCOORDINATE_TYPE,
  TIMESERIES_TYPE,
} from "../util/types";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "select-target-view",

  components: {
    VariableFacets,
  },

  computed: {
    availableTargetVarsPage(): number {
      return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    html(): (group: Group) => HTMLDivElement {
      return (group: Group) => {
        const container = document.createElement("div");
        const targetElem = document.createElement("button");

        const unsupported = this.unsupportedTargets.has(group.colName);
        targetElem.className += "btn btn-sm btn-success mb-2";
        if (unsupported) {
          targetElem.className += " disabled";
        }

        targetElem.innerHTML = "Select Target";
        if (!unsupported) {
          // only add listener on supported target types
          targetElem.addEventListener("click", () => {
            const target = group.colName;
            // remove from training
            const training = routeGetters.getDecodedTrainingVariableNames(
              this.$store
            );
            const index = training.indexOf(target);
            if (index !== -1) {
              training.splice(index, 1);
            }

            const v = this.variables.find((v) => {
              return v.colName === group.colName;
            });
            if (v && v.grouping) {
              if (v.grouping.subIds.length > 0) {
                v.grouping.subIds.forEach((subId) => {
                  const exists = training.find((t) => {
                    return t === subId;
                  });
                  if (!exists) {
                    training.push(subId);
                  }
                });
              } else {
                const exists = training.find((t) => {
                  return t === v.grouping.idCol;
                });
                if (!exists) {
                  training.push(v.grouping.idCol);
                }
              }
            }

            // kick off the fetch task and wait for the result - when we've got it, update the route with info
            const dataset = routeGetters.getRouteDataset(this.$store);
            datasetActions
              .fetchTask(this.$store, {
                dataset: dataset,
                targetName: group.colName,
                variableNames: [],
              })
              .then((response) => {
                const task = response.data.task.join(",");

                const varModesMap = routeGetters.getDecodedVarModes(
                  this.$store
                );
                if (task.includes("timeseries")) {
                  training.forEach((v) => {
                    if (v !== group.colName) {
                      varModesMap.set(v, SummaryMode.Timeseries);
                    }
                  });
                }
                const varModesStr = varModesToString(varModesMap);

                const routeArgs = {
                  target: group.colName,
                  dataset: dataset,
                  filters: routeGetters.getRouteFilters(this.$store),
                  training: training.join(","),
                  task: task,
                  varModes: varModesStr,
                };

                appActions.logUserEvent(this.$store, {
                  feature: Feature.SELECT_TARGET,
                  activity: Activity.PROBLEM_DEFINITION,
                  subActivity: SubActivity.PROBLEM_SPECIFICATION,
                  details: { target: group.colName },
                });

                const entry = createRouteEntry(
                  SELECT_TRAINING_ROUTE,
                  routeArgs
                );
                this.$router.push(entry);
              })
              .catch((error) => {
                console.error(error);
              });
          });
        }
        container.appendChild(targetElem);
        return container;
      };
    },

    instanceName(): string {
      return AVAILABLE_TARGET_VARS_INSTANCE;
    },

    numRowsPerPage(): number {
      return NUM_PER_TARGET_PAGE;
    },

    problemDefinition(): string {
      return Activity.PROBLEM_DEFINITION;
    },

    summaries(): VariableSummary[] {
      let summaries = datasetGetters.getVariableSummaries(this.$store);
      summaries = filterSummariesByDataset(summaries, this.dataset);

      // Fetch the grouped features.
      const groupedFeatures = datasetGetters
        .getGroupings(this.$store)
        .filter((group) => Array.isArray(group.grouping.subIds))
        .map((group) => group.grouping.subIds)
        .flat();

      // Remove summaries of features used in a grouping.
      summaries = summaries.filter(
        (summary) => !groupedFeatures.includes(summary.key)
      );

      return summaries;
    },

    unsupportedTargets(): Set<string> {
      return new Set(
        this.variables
          .filter((v) => isUnsupportedTargetVar(v.colName, v.colType))
          .map((v) => v.colName)
      );
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
  },

  watch: {
    availableTargetVarsPage() {
      viewActions.fetchSelectTargetData(this.$store, false);
    },
  },

  beforeMount() {
    viewActions.fetchSelectTargetData(this.$store, true);
  },

  methods: {
    groupingClick(type) {
      const entry = createRouteEntry(GROUPING_ROUTE, {
        dataset: routeGetters.getRouteDataset(this.$store),
        groupingType: type,
      });
      this.$router.push(entry);
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

<style scoped>
.sub-header-action {
  align-self: end;
  margin-top: 1em;
}

.sub-header-action /deep/ .btn {
  font-weight: bold;
}

.sub-header-action /deep/ .btn + .btn {
  margin-left: 0.5em;
}

.sub-header-action /deep/ .fa {
  margin-right: 0.5em;
}

/* List of targets */
.available-target {
  padding-bottom: 1rem;
}

/* Make all those elements full-height to fit the non-scrollable page design. */
.available-target,
.available-target-variables,
.available-target-variables /deep/ .variable-facets {
  height: 100%;
}

/* Render items as columns */
.available-target-variables /deep/ .variable-facets-container {
  column-count: 3;
  column-gap: 1rem;
}

.available-target-variables /deep/ .variable-facets-item {
  break-inside: avoid;
  display: inline-block;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
  page-break-inside: avoid;
  width: 100%;
  -webkit-backface-visibility: hidden;
  -webkit-column-break-inside: avoid;
}

.available-target-variables
  /deep/
  .facets-group
  .facets-facet-horizontal
  .facet-range {
  cursor: pointer !important;
}

.available-target-variables /deep/ .facet-filters {
  padding: 1rem 0;
}
</style>
