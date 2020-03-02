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
import { Solution } from "../store/requests/index";
import Vue from "vue";

export default Vue.extend({
  name: "recent-solutions",

  components: {
    SolutionPreview
  },

  props: {
    maxSolutions: {
      default: 20,
      type: Number as () => number
    }
  },

  computed: {
    recentSolutions(): Solution[] {
      return requestGetters
        .getSolutions(this.$store)
        .slice(0, this.maxSolutions);
    }
  }
});
</script>

<style></style>
