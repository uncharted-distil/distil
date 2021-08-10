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
  <div>
    <div class="solution-preview" @click="onResult()">
      <div class="solution-header">
        <div><strong>Dataset:</strong> {{ solution.dataset }}</div>
        <div><strong>Date:</strong> {{ formattedTime }}</div>
      </div>
      <div class="solution-body">
        <div><strong>Feature:</strong> {{ solution.feature }}</div>
        <div>
          <b-badge v-if="isPending">
            {{ status }}
          </b-badge>
          <b-badge variant="info" v-if="isRunning">
            {{ status }}
          </b-badge>
          <div v-if="isCompleted">
            <b-badge
              variant="info"
              v-bind:key="score.metric"
              v-for="score in solution.scores"
            >
              {{ score.label }}: {{ score.value.toFixed(2) }}
            </b-badge>
          </div>
          <div v-if="isErrored">
            <b-badge variant="danger"> ERROR </b-badge>
          </div>
        </div>
      </div>
    </div>
    <div class="solution-progress">
      <b-progress
        v-if="isRunning"
        :value="percentComplete"
        variant="outline-secondary"
        striped
        :animated="true"
      ></b-progress>
    </div>
  </div>
</template>

<script lang="ts">
import moment from "moment";
import { createRouteEntry } from "../util/routes";
import { getters as routeGetters } from "../store/route/module";
import { actions as dataActions } from "../store/dataset/module";
import {
  SOLUTION_PENDING,
  SOLUTION_FITTING,
  SOLUTION_SCORING,
  SOLUTION_PRODUCING,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED,
  Solution,
} from "../store/requests/index";
import { APPLY_MODEL_ROUTE, RESULTS_ROUTE } from "../store/route/index";
import Vue from "vue";
import { Location } from "vue-router";
import { Dictionary } from "lodash";
import {
  SOLUTION_PROGRESS,
  SOLUTION_LABELS,
  openModelSolution,
} from "../util/solutions";

export default Vue.extend({
  name: "solution-preview",

  props: {
    solution: Object as () => Solution,
  },

  computed: {
    percentComplete(): number {
      return SOLUTION_PROGRESS[this.solution.progress];
    },
    formattedTime(): string {
      const t = moment(this.solution.timestamp);
      return t.format("MMM Do YYYY, h:mm:ss a");
    },
    status(): string {
      return SOLUTION_LABELS[this.solution.progress];
    },
    isPending(): boolean {
      return this.solution.progress === SOLUTION_PENDING;
    },
    isRunning(): boolean {
      return (
        this.solution.progress === SOLUTION_FITTING ||
        this.solution.progress === SOLUTION_SCORING ||
        this.solution.progress === SOLUTION_PRODUCING
      );
    },
    isCompleted(): boolean {
      return this.solution.progress === SOLUTION_COMPLETED;
    },
    isErrored(): boolean {
      return this.solution.progress === SOLUTION_ERRORED;
    },
    isBad(): boolean {
      return this.solution.isBad;
    },
  },

  methods: {
    async onResult(): Promise<void> {
      const args = await openModelSolution(this.$router, {
        datasetId: this.solution.dataset,
        targetFeature: this.solution.feature,
        solutionId: this.solution.solutionId,
        variableFeatures: this.solution.features.map((f) => f.featureName),
      });
      const entry = createRouteEntry(APPLY_MODEL_ROUTE, args);
      this.$router.push(entry).catch((err) => console.debug(err));
    },
  },
});
</script>

<style>
.solution-preview {
  display: flex;
  flex-direction: column;
}
.solution-header {
  display: flex;
  justify-content: space-between;
}
.solution-body {
  display: flex;
  justify-content: space-between;
}
.solution-preview .badge {
  display: block;
  margin: 4px 0;
}
.solution-progress {
  margin: 6px 0;
}
</style>
