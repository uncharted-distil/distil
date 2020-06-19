<template>
  <div class="available-target-variables">
    <variable-facets
      enable-search
      enable-type-change
      enable-title
      ignore-highlights
      enable-typefiltering
      :instance-name="instanceName"
      :rows-per-page="numRowsPerPage"
      :summaries="summaries"
      :html="html"
      :logActivity="problemDefinition"
    >
    </variable-facets>
  </div>
</template>

<script lang="ts">
import "jquery";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { createRouteEntry, varModesToString } from "../util/routes";
import { filterSummariesByDataset, getComposedVariableKey } from "../util/data";
import VariableFacets from "../components/VariableFacets.vue";
import {
  Grouping,
  Variable,
  VariableSummary,
  SummaryMode
} from "../store/dataset/index";
import { hasComputedVarPrefix, isUnsupportedTargetVar } from "../util/types";
import {
  AVAILABLE_TARGET_VARS_INSTANCE,
  SELECT_TRAINING_ROUTE
} from "../store/route/index";
import { Group } from "../util/facets";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import Vue from "vue";

// 9 so it makes a nice clean grid
const NUM_TARGET_PER_PAGE = 9;

export default Vue.extend({
  name: "available-target-variables",

  components: {
    VariableFacets
  },

  computed: {
    problemDefinition(): string {
      return Activity.PROBLEM_DEFINITION;
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    summaries(): VariableSummary[] {
      let summaries = datasetGetters.getVariableSummaries(this.$store);
      summaries = filterSummariesByDataset(summaries, this.dataset);

      // Fetch the grouped features.
      const groupedFeatures = datasetGetters
        .getGroupings(this.$store)
        .filter(group => Array.isArray(group.grouping.subIds))
        .map(group => group.grouping.subIds)
        .flat();

      // Remove summaries of features used in a grouping.
      summaries = summaries.filter(
        summary => !groupedFeatures.includes(summary.key)
      );

      return summaries;
    },

    numRowsPerPage(): number {
      return NUM_TARGET_PER_PAGE;
    },
    instanceName(): string {
      return AVAILABLE_TARGET_VARS_INSTANCE;
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    unsupportedTargets(): Set<string> {
      return new Set(
        this.variables
          .filter(v => isUnsupportedTargetVar(v.colName, v.colType))
          .map(v => v.colName)
      );
    },
    html(): (group: Group) => HTMLDivElement {
      return (group: Group) => {
        const container = document.createElement("div");
        const targetElem = document.createElement("button");

        const unsupported = this.unsupportedTargets.has(group.colName);
        targetElem.className += "btn btn-sm btn-success ml-2 mr-2 mb-2";
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

            const v = this.variables.find(v => {
              return v.colName === group.colName;
            });
            if (v && v.grouping) {
              if (v.grouping.subIds.length > 0) {
                v.grouping.subIds.forEach(subId => {
                  const exists = training.find(t => {
                    return t === subId;
                  });
                  if (!exists) {
                    training.push(subId);
                  }
                });
              } else {
                const exists = training.find(t => {
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
                variableNames: []
              })
              .then(response => {
                const task = response.data.task.join(",");

                const varModesMap = routeGetters.getDecodedVarModes(
                  this.$store
                );
                if (task.includes("timeseries")) {
                  training.forEach(v => {
                    if (v !== group.colName) {
                      varModesMap.set(v, SummaryMode.Timeseries);
                    }
                  });
                } else if (task.includes("remoteSensing")) {
                  training.forEach(v => {
                    if (v !== group.colName) {
                      varModesMap.set(v, SummaryMode.RemoteSensing);
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
                  varModes: varModesStr
                };

                appActions.logUserEvent(this.$store, {
                  feature: Feature.SELECT_TARGET,
                  activity: Activity.PROBLEM_DEFINITION,
                  subActivity: SubActivity.PROBLEM_SPECIFICATION,
                  details: { target: group.colName }
                });

                const entry = createRouteEntry(
                  SELECT_TRAINING_ROUTE,
                  routeArgs
                );
                this.$router.push(entry);
              })
              .catch(error => {
                console.error(error);
              });
          });
        }
        container.appendChild(targetElem);
        return container;
      };
    }
  }
});
</script>

<style>
.available-target-variables {
  height: 100%;
}

/* Render items as columns */
.available-target-variables .variable-facets-container {
  column-count: 3;
  column-gap: 1rem;
}

.available-target-variables .variable-facets-item {
  display: inline-block;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
  width: 100%;
}

.available-target-variables
  .facets-group
  .facets-facet-horizontal
  .facet-range {
  cursor: pointer !important;
}

.available-target-variables .facet-filters {
  padding: 2rem;
}
</style>
