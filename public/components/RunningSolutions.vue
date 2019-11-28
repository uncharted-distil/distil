<template>
  <b-card header="Pending Models">
    <div v-if="runningSolutions.length === 0">None</div>
    <b-list-group
      v-bind:key="solution.solutionId"
      v-for="solution in runningSolutions"
    >
      <solution-preview :result="solution"></solution-preview>
    </b-list-group>
  </b-card>
</template>

<script lang="ts">
import SolutionPreview from "../components/SolutionPreview";
import { getters } from "../store/solutions/module";
import { Solution } from "../store/solutions/index";
import Vue from "vue";

export default Vue.extend({
  name: "running-solutions",

  props: {
    maxSolutions: {
      default: 20,
      type: Number as () => number
    }
  },

  components: {
    SolutionPreview
  },

  computed: {
    runningSolutions(): Solution[] {
      return getters
        .getRunningSolutions(this.$store)
        .slice(0, this.maxSolutions);
    }
  }
});
</script>

<style></style>
