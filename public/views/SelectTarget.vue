<template>
  <div class="container-fluid d-flex flex-column h-100 select-view">
    <!-- Spacer for the App.vue <navigation> component. -->
    <div class="row flex-0-nav"></div>

    <!-- Title of the page. -->
    <header class="header row">
      <div class="col-12 col-md-10">
        <h5 class="header-title">
          Dataset Overview: Select Feature to Predict
        </h5>
      </div>
    </header>

    <!--
      <b-button @click="onGroupingClick">
        Create Variable Grouping
      </b-button>
    -->

    <!-- List of features -->
    <section class="target-container row justify-content-center">
      <div class="col-12 col-md-10">
        <available-target-variables />
      </div>
    </section>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Variable } from "../store/dataset/index";
import AvailableTargetVariables from "../components/AvailableTargetVariables";
import { actions as viewActions } from "../store/view/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { isTimeType } from "../util/types";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";
import { GROUPING_ROUTE } from "../store/route";

export default Vue.extend({
  name: "select-target-view",

  components: {
    AvailableTargetVariables
  },

  computed: {
    availableTargetVarsPage(): number {
      return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    hasTimeVariable(): boolean {
      return this.variables.filter(v => isTimeType(v.colType)).length > 0;
    }
  },

  watch: {
    availableTargetVarsPage() {
      viewActions.fetchSelectTargetData(this.$store, false);
    }
  },

  beforeMount() {
    viewActions.fetchSelectTargetData(this.$store, true);
  },

  methods: {
    onGroupingClick() {
      const entry = createRouteEntry(GROUPING_ROUTE, {
        dataset: routeGetters.getRouteDataset(this.$store)
      });
      this.$router.push(entry);
    }
  }
});
</script>

<style>
.select-view .nav-link {
  padding: 1rem 0 0.25rem 0;
  border-bottom: 1px solid #e0e0e0;
  color: rgba(0, 0, 0, 0.87);
}
.select-view .variable-facets {
  height: 100%;
}
.select-view .nav-tabs .nav-item a {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}
.select-view .nav-tabs .nav-link {
  color: #757575;
}
.select-view .nav-tabs .nav-link.active {
  color: rgba(0, 0, 0, 0.87);
}

.select-view .target-container {
  height: 100%;
  padding-bottom: 1rem;
}
</style>
