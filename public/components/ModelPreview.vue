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
  <div class="card card-result">
    <div
      class="model-header hover card-header"
      variant="dark"
      @click="onResult()"
    >
      <a class="nav-link">
        <i class="fa fa-connectdevelop" /> <b>Model Name:</b>
        {{ model.modelName }}
      </a>
      <a class="nav-link"><b>Dateset Name:</b> {{ model.datasetName }}</a>
      <a class="nav-link"><b>Features:</b> {{ model.variables.length }}</a>
      <a class="nav-link"><b>Target:</b> {{ model.target.displayName }}</a>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-4">
          <span><b>Features:</b></span>
          <ul>
            <li v-for="variable in topVariables" :key="variable">
              {{ variable }}
            </li>
          </ul>
        </div>
        <div class="col-8">
          <span><b>Description:</b></span>
          <p class="small-text">
            {{ model.modelDescription || "n/a" }}
          </p>
        </div>
      </div>

      <div class="row mt-1">
        <div v-if="!expanded" class="col-12">
          <b-button
            class="full-width hover"
            variant="outline-secondary"
            @click="toggleExpansion()"
          >
            More Details...
          </b-button>
        </div>
        <div v-if="expanded" class="col-12">
          <span><b>All Variables:</b></span>
          <p>
            <span
              v-for="(variable, i) in sortedVariables"
              :key="variable.displayName"
            >
              {{
                variable.displayName +
                (i !== model.variables.length - 1 ? ", " : ".")
              }}
            </span>
          </p>
          <b-button
            class="full-width hover"
            variant="outline-secondary"
            @click="toggleExpansion()"
          >
            Less Details...
          </b-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import { Model, VariableDetail } from "../store/model/index";
import { openModelSolution } from "../util/solutions";

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
  name: "model-preview",

  props: {
    model: Object as () => Model,
  },

  data() {
    return {
      expanded: false,
    };
  },

  computed: {
    sortedVariables(): VariableDetail[] {
      return this.model.variableDetails.slice().sort((a, b) => b.rank - a.rank);
    },
    topVariables(): string[] {
      return this.sortedVariables
        .slice(0, NUM_TOP_FEATURES)
        .map((a) => a.displayName);
    },
  },

  methods: {
    onResult() {
      openModelSolution(this.$router, {
        datasetId: this.model.datasetId,
        targetFeature: this.model.target.key,
        fittedSolutionId: this.model.fittedSolutionId,
        variableFeatures: this.model.variables,
      });
    },
    toggleExpansion() {
      this.expanded = !this.expanded;
    },
  },
});
</script>

<style>
.highlight {
  background-color: #87cefa;
}
.model-header {
  display: flex;
  padding: 4px 8px;
  color: white;
  justify-content: space-between;
  border: none;
  border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}
.card-result .card-header {
  background-color: #424242;
}
.card-result .card-header:hover {
  color: #fff;
  background-color: #535353;
}
.model-header:hover {
  text-decoration: underline;
}
</style>
