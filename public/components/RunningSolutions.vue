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
  <b-card header="Pending Models">
    <div v-if="runningSolutions.length === 0">None</div>
    <b-list-group
      v-bind:key="solution.solutionId"
      v-for="solution in runningSolutions"
    >
      <b-list-group-item href="#" v-bind:key="solution.solutionId">
        <solution-preview :solution="solution"></solution-preview>
      </b-list-group-item>
    </b-list-group>
  </b-card>
</template>

<script lang="ts">
import SolutionPreview from "../components/SolutionPreview";
import { getters as requestGetters } from "../store/requests/module";
import { Solution } from "../store/requests/index";
import Vue from "vue";

export default Vue.extend({
  name: "running-solutions",

  components: {
    SolutionPreview,
  },

  props: {
    maxSolutions: {
      default: 20,
      type: Number as () => number,
    },
  },

  computed: {
    runningSolutions(): Solution[] {
      return requestGetters
        .getRunningSolutions(this.$store)
        .slice(0, this.maxSolutions);
    },
  },
});
</script>

<style></style>
