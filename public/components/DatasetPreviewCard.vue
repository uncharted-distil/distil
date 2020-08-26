<template>
  <div class="card card-result">
    <div class="dataset-card-header hover card-header" variant="dark">
      <b>{{ dataset.name }}</b>
      <b-button
        size="sm"
        variant="secondary"
        class="remove-from-join-button"
        @click="removeFromJoin(dataset.id)"
        ><i class="fa fa-times"></i
      ></b-button>
    </div>
    <div class="card-body">
      <div class="row align-items-center justify-content-center">
        <div class="col-6">
          <div>
            <b>Features:</b>
            {{ filterVariablesByFeature(dataset.variables).length }}
          </div>
          <div><b>Rows:</b> {{ dataset.numRows }}</div>
          <div><b>Size:</b> {{ formatBytes(dataset.numBytes) }}</div>
        </div>
        <!-- <div class='col-6'>
					<span><b>Top features:</b></span>
					<ul>
						<li :key="variable.name" v-for='variable in topVariables'>
							{{variable.colDisplayName}}
						</li>
					</ul>
				</div> -->
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import {
  sortVariablesByImportance,
  filterVariablesByFeature,
} from "../util/data";
import { formatBytes } from "../util/bytes";
import { Dataset, Variable } from "../store/dataset/index";

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
  name: "dataset-preview-card",

  props: {
    dataset: Object as () => Dataset,
  },

  computed: {
    topVariables(): Variable[] {
      return sortVariablesByImportance(this.dataset.variables.slice(0)).slice(
        0,
        NUM_TOP_FEATURES,
      );
    },
  },

  methods: {
    formatBytes(n: number): string {
      return formatBytes(n);
    },
    filterVariablesByFeature(variables: Variable[]): Variable[] {
      return filterVariablesByFeature(variables);
    },
    removeFromJoin(arg) {
      this.$emit("remove-from-join", arg);
    },
  },
});
</script>

<style>
.dataset-card-header {
  display: flex;
  padding: 4px 8px;
  color: white;
  justify-content: space-between;
  border: none;
}
.card-result .card-header {
  background-color: #424242;
}

.remove-from-join-button {
  position: absolute;
  top: 8px;
  right: 8px;
  cursor: pointer;
}
</style>
