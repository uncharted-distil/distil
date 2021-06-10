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
    <div class="dataset-card-header hover card-header" variant="dark">
      <b>{{ dataset.name }}</b>
      <b-button
        size="sm"
        variant="secondary"
        class="remove-from-join-button"
        @click="removeFromJoin(dataset.id)"
      >
        <i class="fa fa-times" />
      </b-button>
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
import Vue from "vue";
import {
  // sortVariablesByPCARanking,
  filterVariablesByFeature,
} from "../util/data";
import { formatBytes } from "../util/bytes";
import { Dataset, Variable } from "../store/dataset/index";
import { EventList } from "../util/events";
const NUM_TOP_FEATURES = 5;

export default Vue.extend({
  name: "DatasetPreviewCard",

  props: {
    dataset: Object as () => Dataset,
  },

  // computed: {
  //   topVariables(): Variable[] {
  //     const variables = this.dataset.variables.slice(0);
  //     return sortVariablesByPCARanking(variables).slice(0, NUM_TOP_FEATURES);
  //   },
  // },

  methods: {
    formatBytes(n: number): string {
      return formatBytes(n);
    },

    filterVariablesByFeature(variables: Variable[]): Variable[] {
      return filterVariablesByFeature(variables);
    },

    removeFromJoin(arg) {
      this.$emit(EventList.JOIN.REMOVE_EVENT, arg);
    },
  },
});
</script>

<style scoped>
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
