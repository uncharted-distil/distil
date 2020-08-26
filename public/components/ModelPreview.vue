<template>
  <div class="card card-result" @click="onResult()">
    <div class="model-header hover card-header" variant="dark">
      <a class="nav-link">
        <i class="fa fa-connectdevelop"></i> <b>Model Name:</b>
        {{ model.modelName }}
      </a>
      <a class="nav-link"><b>Dateset Name:</b> {{ model.datasetName }}</a>
      <a class="nav-link"><b>Features:</b> {{ model.variables.length }}</a>
      <a class="nav-link"><b>Target:</b> {{ model.target }}</a>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-4">
          <span><b>Features:</b></span>
          <ul>
            <li :key="variable.name" v-for="variable in model.variables">
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
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import { Model } from "../store/model/index";
import { openModelSolution } from "../util/solutions";

export default Vue.extend({
  name: "model-preview",

  props: {
    model: Object as () => Model,
  },

  methods: {
    onResult() {
      openModelSolution(this.$router, {
        datasetId: this.model.datasetId,
        targetFeature: this.model.target,
        fittedSolutionId: this.model.fittedSolutionId,
        variableFeatures: this.model.variables,
      });
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
