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
  <b-navbar toggleable="lg" type="dark" class="fixed-top">
    <b-nav-toggle target="nav-collapse" />

    <!-- Branding -->
    <img
      src="/images/uncharted.svg"
      class="app-icon"
      :class="{ 'is-prototype': isPrototype }"
      height="36"
      width="36"
      :title="version"
      @click="onLogoClick"
    />
    <b-navbar-brand>Distil</b-navbar-brand>

    <!-- Left Side -->
    <b-collapse id="nav-collapse" is-nav>
      <b-navbar-nav>
        <b-nav-item :active="isActive(SEARCH_ROUTE)" @click="onSearch">
          <i class="fa fa-home nav-icon" /> Select Model or Dataset
        </b-nav-item>

        <!-- If search produces a model of interest, select it for reuse: will start Apply Model workflow. -->
        <template v-if="isApplyModel && !isActive(DATA_EXPLORER_ROUTE)">
          <b-nav-item
            :active="isActive(APPLY_MODEL_ROUTE)"
            @click="onApplyModel"
          >
            <i class="fa fa-table nav-icon" /> Apply Model: Select New Data
          </b-nav-item>

          <b-nav-item
            :active="isActive(PREDICTION_ROUTE)"
            :disabled="isActive(APPLY_MODEL_ROUTE)"
            @click="onPredictions"
          >
            <i class="fa fa-line-chart nav-icon" /> View Predictions
          </b-nav-item>
        </template>

        <!-- If no appropriate model exist, select a dataset: will start New Model workflow. -->
        <template v-else-if="hasDataset && !isActive(DATA_EXPLORER_ROUTE)">
          <b-nav-item
            :active="isActive(SELECT_TARGET_ROUTE)"
            @click="onSelectTarget"
          >
            <i class="fa fa-dot-circle-o nav-icon" /> New Model: Select Target
          </b-nav-item>

          <b-nav-item
            :active="isActive(SELECT_TRAINING_ROUTE)"
            :disabled="hasNoDatasetAndTarget"
            @click="onSelectData"
          >
            <i class="fa fa-sign-in nav-icon" /> Select Model Features
          </b-nav-item>

          <b-nav-item
            :active="isActive(RESULTS_ROUTE)"
            :disabled="hasNoDatasetAndTarget"
            @click="onResults"
          >
            <i class="fa fa-check-circle nav-icon" /> Check Models
          </b-nav-item>
        </template>
        <template v-else-if="isActive(DATA_EXPLORER_ROUTE)">
          <b-nav-item
            :active="explorerSelectState"
            @click="explorerNav('select')"
          >
            <i class="fa fa-dot-circle-o nav-icon" /> New Model
          </b-nav-item>
          <b-nav-item
            :active="explorerResultState"
            :disabled="hasNoDatasetAndTarget"
            @click="explorerNav('result')"
          >
            <i class="fa fa-check-circle nav-icon" /> Check Models
          </b-nav-item>
          <b-nav-item
            :active="explorerPredictionState"
            :disabled="hasNoDatasetAndTarget"
            @click="explorerNav('prediction')"
          >
            <i class="fa fa-line-chart nav-icon" /> View Predictions
          </b-nav-item>
        </template>
      </b-navbar-nav>
    </b-collapse>

    <!--<b-nav-item
      @click="onJoinDatasets"
      v-if="isJoinDatasets && isActive(JOIN_DATASETS_ROUTE)"
      :active="isActive(JOIN_DATASETS_ROUTE)"
    >
      <i class="fa fa-database nav-icon"></i> Join Datasets
    </b-nav-item>-->

    <!-- Right side -->
    <b-navbar-nav class="ml-auto">
      <b-nav-item :href="helpURL">Help</b-nav-item>
    </b-navbar-nav>
  </b-navbar>
</template>

<script lang="ts">
import "../assets/images/uncharted.svg";
import {
  gotoApplyModel,
  // gotoHome,
  gotoJoinDatasets,
  gotoPredictions,
  gotoResults,
  gotoSearch,
  gotoSelectData,
  gotoSelectTarget,
} from "../util/nav";
import {
  actions as appActions,
  getters as appGetters,
} from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import {
  APPLY_MODEL_ROUTE,
  // HOME_ROUTE,
  SEARCH_ROUTE,
  JOIN_DATASETS_ROUTE,
  SELECT_TARGET_ROUTE,
  SELECT_TRAINING_ROUTE,
  RESULTS_ROUTE,
  PREDICTION_ROUTE,
  DATA_EXPLORER_ROUTE,
} from "../store/route/index";
import { restoreView } from "../util/view";
import Vue from "vue";
import { ExplorerStateNames } from "../util/dataExplorer";
import { EventList } from "../util/events";

export default Vue.extend({
  name: "NavBar",

  data() {
    return {
      APPLY_MODEL_ROUTE: APPLY_MODEL_ROUTE,
      // HOME_ROUTE: HOME_ROUTE,
      SEARCH_ROUTE: SEARCH_ROUTE,
      JOIN_DATASETS_ROUTE: JOIN_DATASETS_ROUTE,
      SELECT_TARGET_ROUTE: SELECT_TARGET_ROUTE,
      SELECT_TRAINING_ROUTE: SELECT_TRAINING_ROUTE,
      RESULTS_ROUTE: RESULTS_ROUTE,
      PREDICTION_ROUTE: PREDICTION_ROUTE,
      DATA_EXPLORER_ROUTE: DATA_EXPLORER_ROUTE,
    };
  },

  computed: {
    path(): string {
      return routeGetters.getRoutePath(this.$store);
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    joinDatasets(): string[] {
      return routeGetters.getRouteJoinDatasets(this.$store);
    },

    joinDatasetsHash(): string {
      return routeGetters.getRouteJoinDatasetsHash(this.$store);
    },

    isJoinDatasets(): boolean {
      return this.joinDatasets.length === 2 || this.hasJoinDatasetView();
    },
    dataExplorerState(): ExplorerStateNames {
      return routeGetters.getDataExplorerState(this.$store);
    },
    isApplyModel(): boolean {
      /*
        Check if we requested in the route for an Apply Model navigation,
        or, in the case of a prediction a fitted solution ID.
       */
      return (
        routeGetters.isApplyModel(this.$store) ||
        !!routeGetters.getRouteFittedSolutionId(this.$store)
      );
    },
    explorerSelectState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.SELECT_VIEW;
    },
    explorerResultState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.RESULT_VIEW;
    },
    explorerPredictionState(): boolean {
      return this.dataExplorerState === ExplorerStateNames.PREDICTION_VIEW;
    },
    hasDataset(): boolean {
      return !!this.dataset;
    },

    helpURL(): string {
      return appGetters.getHelpURL(this.$store);
    },

    hasNoDatasetAndTarget(): boolean {
      return !(!!this.dataset && !!this.target);
    },

    version(): string {
      return appGetters.getAllSystemVersions(this.$store);
    },

    isPrototype(): boolean {
      return appGetters.isPrototype(this.$store);
    },
  },

  methods: {
    explorerNav(state: string) {
      this.$emit(EventList.EXPLORER.NAV_EVENT, state);
    },
    isActive(view) {
      return view === this.path;
    },
    isState(state: ExplorerStateNames): boolean {
      return state === this.dataExplorerState;
    },
    // onHome() {
    //   gotoHome(this.$router);
    // },

    onSearch() {
      gotoSearch(this.$router);
    },

    onJoinDatasets() {
      gotoJoinDatasets(this.$router);
    },

    onSelectTarget() {
      gotoSelectTarget(this.$router);
    },

    onSelectData() {
      gotoSelectData(this.$router);
    },

    onResults() {
      gotoResults(this.$router);
    },

    onApplyModel() {
      gotoApplyModel(this.$router);
    },

    onPredictions() {
      gotoPredictions(this.$router);
    },

    hasJoinDatasetView(): boolean {
      return !!restoreView(JOIN_DATASETS_ROUTE, this.joinDatasetsHash);
    },

    onLogoClick() {
      appActions.togglePrototype(this.$store);
    },
  },
});
</script>

<style scoped>
.navbar {
  background-color: var(--gray-900);
  box-shadow: 0 6px 12px 0 rgba(0, 0, 0, 0.1);
  justify-content: flex-start;
}

.app-icon {
  margin-right: 0.33em;
}

.app-icon.is-prototype {
  filter: invert(1);
}

.nav-item {
  font-weight: bold;
  letter-spacing: 0.01rem;
  white-space: nowrap;
}

/* Display an arrow if two link are next to each others. */
.navbar-collapse:not(.show) .nav-item + .nav-item .nav-link::before,
.navbar-collapse.show .nav-item + .nav-item::before {
  color: var(--gray-600);
  font-family: FontAwesome;
  font-weight: bold;
}

/* Horizontal arrow if the menu is visible (not collapsed). */
.navbar-collapse:not(.show) .nav-item + .nav-item .nav-link::before {
  content: "\f105"; /* angle-right => https://fontawesome.com/v4.7.0/cheatsheet/ */
  margin-right: 1em;
}

/* Change the arrow to be vertical if the menu is collapsed. */
.navbar-collapse.show .nav-item + .nav-item {
  position: relative;
  margin-top: 1em;
}
.navbar-collapse.show .nav-item + .nav-item::before {
  content: "\f107"; /* angle-down => https://fontawesome.com/v4.7.0/cheatsheet/ */
  left: 0.65em;
  position: absolute;
  top: -1em;
}

/* Icon. */
.nav-icon {
  border-radius: 50%;
  height: 30px;
  margin-right: 0.25em;
  padding: 7px;
  text-align: center;
  width: 30px;
}

/*
  In the following I use the ID #distil-app to overwrite the Bootstrap CSS
  by increasing the selectors specificity.
*/

/* Default colours */
#distil-app .nav-link {
  transition: color 0.25s;
  color: var(--gray-600);
}
#distil-app .nav-link .nav-icon {
  transition: background 0.25s, color 0.25s;
  background-color: var(--gray-800);
  color: var(--gray-400);
}

/* Active and non disabled on hover nav-item */
#distil-app .nav-link.active,
#distil-app .nav-link:not(.disable):hover {
  color: var(--gray-400);
}
#distil-app .nav-link.active .nav-icon,
#distil-app .nav-link:not(.disable):hover .nav-icon {
  background-color: var(--black);
}

/* Disabled Nav-item */
#distil-app .nav-link.disabled .nav-icon {
  background: none;
  color: var(--gray-600);
}
</style>
