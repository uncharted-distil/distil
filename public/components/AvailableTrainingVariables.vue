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
  <div class="available-training-variables">
    <p class="nav-link font-weight-bold">
      {{ title }}
      <i class="float-right fa fa-angle-right fa-lg" />
    </p>
    <variable-facets
      ref="facets"
      enable-highlighting
      enable-search
      enable-type-change
      :facet-count="variables && variables.length"
      :html="html"
      :is-available-features="isAvailableFeatures"
      :is-features-to-model="!isAvailableFeatures"
      :instance-name="instanceName"
      :pagination="variables && variables.length > numRowsPerPage"
      :rows-per-page="numRowsPerPage"
      :summaries="summaries"
      :enable-color-scales="geoVarExists"
      :include="include"
    >
      <div
        class="d-flex flex-row justify-content-between align-items-center my-2 mx-1"
      >
        <div>
          {{ subtitleInfo }}
        </div>
        <b-button
          v-if="displayAddAll"
          size="sm"
          variant="outline-secondary"
          @click="addAll"
        >
          {{ groupBtnTitle }}
        </b-button>
      </div>
    </variable-facets>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Variable, VariableSummary } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { Group } from "../util/facets";
import { NUM_PER_PAGE } from "../util/data";
import VariableFacets from "./facets/VariableFacets.vue";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { DISTIL_ROLES } from "../util/types";
import { isGeoLocatedType } from "../util/types";
import { overlayRouteEntry } from "../util/routes";
import { EventList, GroupChangeParams } from "../util/events";
export default Vue.extend({
  name: "AvailableTrainingVariables",

  components: {
    VariableFacets,
  },
  props: {
    variables: {
      type: Array as () => Variable[],
      default: [] as Variable[],
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: [] as VariableSummary[],
    },
    title: { type: String as () => string, default: "" },
    groupBtnTitle: { type: String as () => string, default: "" },
    btnTitle: { type: String as () => string, default: "" },
    instanceName: { type: String as () => string, default: "" },
    subtitle: { type: String as () => string, default: "" },
    checkGeoType: { type: Boolean as () => boolean, default: false },
    isAvailableFeatures: { type: Boolean as () => boolean, default: false },
    include: { type: Boolean as () => boolean, default: true },
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    availableTrainingVarsSearch(): string {
      return routeGetters.getRouteAvailableTrainingVarsSearch(this.$store);
    },

    availableVariableSummariesForTraining(): VariableSummary[] {
      return this.summaries.filter(
        (variable) => variable.distilRole != DISTIL_ROLES.Augmented
      );
    },

    displayAddAll(): boolean {
      return this.availableVariableSummariesForTraining.length > 0;
    },

    subtitleInfo(): string {
      const total = this.availableVariableSummariesForTraining?.length ?? 0;
      if (total < 1) return;
      return `${total} ${this.subtitle}`;
    },
    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },

    targetVariable(): Variable {
      return routeGetters.getTargetVariable(this.$store);
    },

    html(): (group: Group) => HTMLElement {
      return (group: Group) => {
        const trainingElem = document.createElement("button");
        trainingElem.className += "btn btn-sm btn-outline-secondary mb-2";
        trainingElem.textContent = this.btnTitle;

        // In the case of a categorical variable with a timeserie selected.
        const isCategorical: boolean = group.type === "categorical";
        if (this.isTimeseries && isCategorical) {
          // Change the meaning of the button as this action is different than the default one.
          trainingElem.textContent = "Add to Timeseries";
        }

        trainingElem.addEventListener("click", async () => {
          // log UI event on server
          appActions.logUserEvent(this.$store, {
            feature: Feature.ADD_FEATURE,
            activity: Activity.DATA_PREPARATION,
            subActivity: SubActivity.DATA_TRANSFORMATION,
            details: { feature: group.key },
          });
          this.variableChange(group);
        });

        return trainingElem;
      };
    },
    geoVarExists(): boolean {
      if (this.checkGeoType) {
        return this.summaries.some((v) => {
          return isGeoLocatedType(v.type);
        });
      }
      return false;
    },
  },
  watch: {
    geoVarExists() {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: this.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
  methods: {
    variableChange(group: Group) {
      this.$emit(EventList.VARIABLES.VAR_SET_CHANGE_EVENT, group);
    },
    addAll() {
      // log UI event on server
      appActions.logUserEvent(this.$store, {
        feature: Feature.ADD_ALL_FEATURES,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });

      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store
      );

      this.variables.forEach((variable) => {
        if (variable.distilRole === DISTIL_ROLES.Augmented) return;
        training.push(variable.key);
      });
      const dataset = routeGetters.getRouteDataset(this.$store);
      const targetName = routeGetters.getRouteTargetVariable(this.$store);
      this.$emit(EventList.VARIABLES.VAR_SET_GROUP_CHANGE_EVENT, {
        dataset,
        targetName,
        variableNames: training,
      } as GroupChangeParams);
    },
  },
});
</script>

<style scoped>
.available-training-variables {
  display: flex;
  flex-direction: column;
}

.available-training-variables /deep/ .variable-facets-wrapper {
  height: 100%;
}
</style>
