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
  <div class="h-100">
    <header v-if="enableFooter">
      <b-button size="sm" @click="showAll">
        <i class="fas fa-eye" /> Show All
      </b-button>
      <b-button size="sm" @click="hideAll">
        <i class="fas fa-eye-slash" /> Hide All
      </b-button>
      <b-button
        v-if="hasTarget"
        variant="primary"
        size="sm"
        class="float-right"
        @click="selectAllTraining"
      >
        Select All Training
      </b-button>
      <positive-label
        v-if="labels && isTargetPanel"
        class="pt-2"
        :labels="labels"
      />
    </header>
    <variable-facets
      enable-highlighting
      enable-search
      enable-type-change
      enable-type-filtering
      ignore-highlights
      class="mh-var-list"
      :is-available-features="isSelectedView"
      :is-result-features="isResultView"
      :include="include"
      :enable-color-scales="enableColorScales"
      :facet-count="searchedActiveVariables.length"
      :instance-name="instanceName"
      :log-activity="problemDefinition"
      :rows-per-page="numRowsPerPage"
      :pagination="
        searchedActiveVariables &&
        searchedActiveVariables.length > numRowsPerPage
      "
      :disabled-color-scales="disabledColorScales"
      :summaries="activeSummaries"
      :dataset-name="dataset"
      :active-variables="variables"
      @search="onSearch"
      @type-change="onTypeChange"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "../../components/facets/VariableFacets.vue";
import PositiveLabel from "../buttons/PositiveLabel.vue";
import {
  SummaryMode,
  TaskTypes,
  Variable,
  VariableSummary,
} from "../../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../../store/route/index";
import { getters as routeGetters } from "../../store/route/module";

import { NUM_PER_PAGE, searchVariables } from "../../util/data";
import {
  getRouteFacetPage,
  overlayRouteEntry,
  RouteArgs,
  varModesToString,
} from "../../util/routes";
import { Activity } from "../../util/userEvents";
import { ExplorerStateNames } from "../../util/explorer";
import { EventList } from "../../util/events";

export default Vue.extend({
  name: "FacetListPane",

  components: {
    PositiveLabel,
    VariableFacets,
  },

  props: {
    variables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: () => [] as VariableSummary[],
    },
    enableColorScales: { type: Boolean as () => boolean, default: false },
    include: { type: Boolean as () => boolean, default: true },
    enableFooter: { type: Boolean as () => boolean, default: false },
    isTargetPanel: { type: Boolean as () => boolean, default: false },
    dataset: { type: String as () => string, default: "" },
    instanceName: {
      type: String as () => string,
      default: DATA_EXPLORER_VAR_INSTANCE,
    },
  },

  data() {
    return {
      numRowsPerPage: NUM_PER_PAGE,
      search: "",
    };
  },

  computed: {
    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store)(this.include);
    },
    labels(): string[] {
      // make sure we are only on a binary classification task
      if (!routeGetters.isBinaryClassification(this.$store)) return;

      // retreive the target variable buckets
      const buckets = this.targetSummaries?.[0]?.baseline?.buckets;
      if (!buckets) return;

      // use the buckets keys as labels
      return buckets.map((bucket) => bucket.key);
    },
    hasTarget(): boolean {
      return !!routeGetters.getTargetVariable(this.$store);
    },
    isSelectedView(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.SELECT_VIEW
      );
    },
    disabledColorScales(): Map<string, boolean> {
      const result = new Map();
      // if footer is disabled there is no hide/display feature
      if (!this.enableFooter) {
        return result;
      }
      // check if the variable is in the state of "display"
      this.variables.forEach((v) => {
        result.set(v.key, !this.isExplore(v.key));
      });
      return result;
    },
    isResultView(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.RESULT_VIEW
      );
    },
    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
    },

    groupedFeatures(): string[] {
      // Fetch the grouped features.
      const groupedFeatures = datasetGetters
        .getGroupings(this.$store)
        .filter((group) => Array.isArray(group.grouping.subIds))
        .map((group) => group.grouping.subIds)
        .flat();
      return groupedFeatures;
    },

    problemDefinition(): string {
      return Activity.PROBLEM_DEFINITION;
    },
    activeVariables(): Variable[] {
      // remove variables used in groupedFeature;
      return this.variables.filter(
        (v) => !this.groupedFeatures.includes(v.key)
      );
    },
    searchedActiveVariables(): Variable[] {
      return searchVariables(this.activeVariables, this.search);
    },
    colorScaleVar(): string {
      return routeGetters.getColorScaleVariable(this.$store);
    },
    activeSummaries(): VariableSummary[] {
      const searchedMap = new Map(
        this.searchedActiveVariables.map((v, idx) => {
          return [v.key, idx];
        })
      );
      const pageId = this.instanceName + ROUTE_PAGE_SUFFIX;
      const page = getRouteFacetPage(pageId, this.$route);
      const begin = (page - 1) * this.numRowsPerPage;
      const currentSummaries = this.summaries
        .filter((s) => {
          return searchedMap.has(s.key);
        })
        .sort((a, b) => {
          return searchedMap.get(a.key) - searchedMap.get(b.key);
        });
      const end = Math.min(page * this.numRowsPerPage, currentSummaries.length);
      return currentSummaries.slice(begin, end);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },

    varsPage(): number {
      return routeGetters.getRouteDataExplorerVarsPage(this.$store);
    },

    varsSearch(): string {
      return routeGetters.getRouteDataExplorerVarsSearch(this.$store);
    },
  },

  watch: {
    varsPage() {
      this.$emit(EventList.SUMMARIES.FETCH_SUMMARIES_EVENT);
    },

    varsSearch() {
      this.$emit(EventList.SUMMARIES.FETCH_SUMMARIES_EVENT);
    },
  },

  methods: {
    onTypeChange() {
      this.$emit(EventList.VARIABLES.TYPE_CHANGE);
    },
    onSearch(term): void {
      this.search = term;
    },
    async selectAllTraining() {
      const list = this.activeVariables.filter((v) => {
        return !this.isTraining(v.key) && !this.isTarget(v.key);
      });
      const args = await this.addTrainingVariables(list.map((v) => v.key));
      this.updateRoute(args);
    },
    showAll() {
      const args = {} as RouteArgs;
      const list = [] as string[];
      this.activeVariables.forEach((variable) => {
        if (!this.isExplore(variable.key)) {
          list.push(variable.key);
        }
      });
      if (!list.length) {
        return;
      }
      args.explore = this.explore.concat(list).join(",");
      this.updateRoute(args);
    },
    hideAll() {
      const args = {} as RouteArgs;
      const map = new Map(
        this.activeVariables
          .filter((v) => {
            return this.isExplore(v.key);
          })
          .map((v) => {
            return [v.key, true];
          })
      );

      args.explore = this.explore.filter((v) => !map.has(v)).join(",");
      this.updateRoute(args);
    },

    updateRoute(args: RouteArgs) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    isTarget(variable: string): boolean {
      return variable === this.target;
    },

    isTraining(variable: string): boolean {
      return this.training?.includes(variable) ?? false;
    },

    isExplore(variable: string): boolean {
      return this.explore.includes(variable);
    },

    async addTrainingVariables(variables: string[]): Promise<RouteArgs> {
      const args = {} as RouteArgs;
      const training = this.training.concat(variables);
      args.training = training.join(",");
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: this.dataset,
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
.mh-var-list {
  max-height: calc(100% - 30px);
}
</style>
