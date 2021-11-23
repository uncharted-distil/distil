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
  <b-button
    v-if="displayTarget"
    class="toggle btn-sm d-flex align-items-center shadow-none"
    :variant="this.isTarget ? 'outline-secondary' : 'primary'"
    @click="updateTarget"
  >
    {{ isTarget ? "Remove Target" : "Select Target" }}
  </b-button>
  <b-button
    v-else-if="displayTraining"
    class="toggle btn-sm d-flex align-items-center shadow-none"
    :variant="isTraining ? 'outline-primary' : 'primary'"
    @click="updateTraining"
  >
    {{ isTraining ? "Remove Training" : "Select Training" }}
  </b-button>
</template>

<script lang="ts">
import Vue from "vue";

import { SummaryMode, TaskTypes, Variable } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { actions as appActions } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { RouteArgs, overlayRouteEntry, varModesToString } from "../util/routes";
import { requestActions } from "../store";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { hasRole } from "../util/data";
import { isUnsupportedTargetVar, DISTIL_ROLES } from "../util/types";

export default Vue.extend({
  name: "ButtonTrainingTarget",
  props: {
    variable: String as () => string,
    datasetName: String as () => string,
    activeVariables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
  },
  computed: {
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    hasTarget(): boolean {
      return !!routeGetters.getTargetVariable(this.$store);
    },

    isTarget(): boolean {
      return this.variable === this.target;
    },

    displayTarget(): boolean {
      const activeVariable = this.activeVariables.find(
        (v) => v.key === this.variable
      );
      return (
        !hasRole(activeVariable, DISTIL_ROLES.Augmented) &&
        (!this.hasTarget || this.isTarget) &&
        !this.isUnsupported
      );
    },

    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },

    isTraining(): boolean {
      return this.training?.includes(this.variable) ?? false;
    },

    displayTraining(): boolean {
      const activeVariable = this.activeVariables.find(
        (v) => v.key === this.variable
      );
      return (
        !hasRole(activeVariable, DISTIL_ROLES.Augmented) &&
        this.hasTarget &&
        !this.isTarget
      );
    },

    unsupportedTargets(): Set<string> {
      return new Set(
        this.activeVariables
          .filter((v) => isUnsupportedTargetVar(v.key, v.colType))
          .map((v) => v.key)
      );
    },

    isUnsupported(): boolean {
      return this.unsupportedTargets.has(this.variable);
    },

    buttonText(): string {
      return `Text`;
    },
  },
  methods: {
    async updateTarget(): Promise<void> {
      const target = this.variable;

      // Is the variable the current target?
      if (this.isTarget) {
        // Remove the variable as target
        this.updateRoute({ target: null, task: null });
        return;
      }

      const args = {} as RouteArgs;
      args.target = target;

      // Filter it out of the training
      const training = this.training.filter((v) => v !== target);

      // Get Variables Grouping and check if our target is one of them
      const groupings = datasetGetters.getGroupings(this.$store);
      const targetGrouping = groupings?.find((g) => g.key === target)?.grouping;
      if (!!targetGrouping) {
        if (targetGrouping.subIds.length > 0) {
          targetGrouping.subIds.forEach((subId) => {
            if (!training.find((t) => t === subId)) {
              training.push(subId);
            }
          });
        } else {
          if (!training.find((t) => t === targetGrouping.idCol)) {
            training.push(targetGrouping.idCol);
          }
        }
      }

      // Get the var modes
      const varModesMap = routeGetters.getDecodedVarModes(this.$store);
      args.varModes = varModesToString(varModesMap);

      // Fetch the task
      try {
        const response = await datasetActions.fetchTask(this.$store, {
          dataset: this.datasetName,
          targetName: target,
          variableNames: [],
        });
        args.task = response.data.task.join(",") ?? "";

        // Update the training variable
        if (args.task.includes("timeseries")) {
          training.forEach((variable) => {
            if (variable !== target) {
              varModesMap.set(variable, SummaryMode.Timeseries);
            }
          });
        }
        await datasetActions.fetchModelingMetrics(this.$store, {
          task: args.task,
        });
      } catch (error) {
        console.log(error);
      }

      // Make the list of training variables' name a string.
      args.training = training.join(",");

      appActions.logUserEvent(this.$store, {
        feature: Feature.SELECT_TARGET,
        activity: Activity.PROBLEM_DEFINITION,
        subActivity: SubActivity.PROBLEM_SPECIFICATION,
        details: { target },
      });
      // fetch existing solutions
      await requestActions.fetchSolutions(this.$store, {
        dataset: this.datasetName,
        target,
      });
      this.updateRoute(args);
      datasetActions.fetchVariableRankings(this.$store, {
        dataset: this.datasetName,
        target,
      });
    },

    async updateTraining(): Promise<void> {
      let args = {} as RouteArgs;
      if (this.isTraining) {
        args.training = this.training
          .filter((v) => v !== this.variable)
          .join(",");
      } else {
        args = await this.addTrainingVariables([this.variable]);
      }

      this.updateRoute(args);
    },

    updateRoute(args: RouteArgs) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    async addTrainingVariables(variables: string[]): Promise<RouteArgs> {
      const args = {} as RouteArgs;
      const training = this.training.concat(variables);
      args.training = training.join(",");
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: this.datasetName,
        targetName: this.target,
        variableNames: training,
      });
      const task = taskResponse.data.task.join(",");
      args.task = task;
      if (task.includes(TaskTypes.REMOTE_SENSING)) {
        const available = routeGetters.getAvailableVariables(this.$store);
        const varModesMap = routeGetters.getDecodedVarModes(this.$store);
        training.forEach((v) => {
          varModesMap.set(v, SummaryMode.MultiBandImage);
        });

        available.forEach((v) => {
          varModesMap.set(v.key, SummaryMode.MultiBandImage);
        });

        varModesMap.set(
          routeGetters.getRouteTargetVariable(this.$store),
          SummaryMode.MultiBandImage
        );
        const varModesStr = varModesToString(varModesMap);
        args.varModes = varModesStr;
      }
      return args;
    },
  },
});
</script>

<style scoped>
.dropdown,
.toggle {
  height: 22px;
}
</style>
