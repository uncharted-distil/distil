<template>
  <b-card header="Recent Models">
    <div v-if="recentSolutions.length === 0">None</div>
    <b-list-group
      v-bind:key="solution.solutionId"
      v-for="solution in recentSolutions"
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
import { getters as modelGetters } from "../store/model/module";
import { Solution } from "../store/requests/index";
import Vue from "vue";
import _ from "lodash";
import moment from "moment";

export default Vue.extend({
  name: "recent-solutions",

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
    // Return recent solutions, filtering down to only those that have
    // been saved.  This is to ensure that the TA2 can re-load the fitted solution
    // for additional produce calls.
    recentSolutions(): Solution[] {
      // find solutions associated with exported models
      const savedModelsMap = _.mapKeys(
        modelGetters.getModels(this.$store),
        (m) => m.fittedSolutionId,
      );
      return requestGetters
        .getSolutions(this.$store)
        .filter((s) => savedModelsMap[s.fittedSolutionId])
        .sort((a, b) => moment(b.timestamp).unix() - moment(a.timestamp).unix())
        .slice(0, this.maxSolutions);
    },
  },
});
</script>

<style></style>
