<template>
  <div class="result-summaries">
    <p class="nav-link font-weight-bold">Results</p>
    <p></p>
    <div v-if="regressionEnabled" class="result-summaries-error">
      <error-threshold-slider></error-threshold-slider>
    </div>
    <p class="nav-link font-weight-bold">Predictions by Model</p>
    <result-facets :regression="regressionEnabled"> </result-facets>
    <b-btn v-b-modal.export variant="primary" class="check-button"
      >Export Model</b-btn
    >

    <b-modal id="export" title="Export" @ok="onExport">
      <div class="check-message-container">
        <i class="fa fa-check-circle fa-3x check-icon"></i>
        <div>
          This action will export solution <b>{{ activeSolutionName }}</b> and
          return to the application start screen.
        </div>
      </div>
    </b-modal>

    <b-modal
      ref="exportSuccessModal"
      title="Export Succeeded"
      cancel-disabled
      hide-header
      hide-footer
    >
      <div class="check-message-container">
        <i class="fa fa-check-circle fa-3x check-icon"></i>
        <div>Export Succeeded.</div>
        <b-btn
          class="mt-3 ml-3 close-modal"
          variant="success"
          block
          @click="hideSuccessModal"
          >OK</b-btn
        >
      </div>
    </b-modal>

    <b-modal
      ref="exportFailModal"
      title="Export Failed"
      cancel-disabled
      hide-header
      hide-footer
    >
      <div class="check-message-container">
        <i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
        <div><b>Export Failed:</b> {{ exportFailureMsg }}</div>
        <b-btn
          class="mt-3 ml-3 close-modal"
          variant="success"
          block
          @click="hideFailureModal"
          >OK</b-btn
        >
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider";
import { getSolutionById } from "../util/solutions";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as solutionGetters } from "../store/solutions/module";
import {
  actions as appActions,
  getters as appGetters
} from "../store/app/module";
import { EXPORT_SUCCESS_ROUTE, ROOT_ROUTE } from "../store/route/index";
import { Variable, TaskTypes } from "../store/dataset/index";
import vueSlider from "vue-slider-component";
import Vue from "vue";
import { Solution } from "../store/solutions/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "result-summaries",

  components: {
    ResultFacets,
    ErrorThresholdSlider,
    vueSlider
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      exportFailureMsg: "",
      symmetricSlider: true
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    regressionEnabled(): boolean {
      const tasks = routeGetters.getRouteTask(this.$store).split(',');
      return tasks.indexOf(TaskTypes.REGRESSION) > -1;
    },

    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    activeSolution(): Solution {
      return getSolutionById(this.$store.state.solutionModule, this.solutionId);
    },

    activeSolutionName(): string {
      return this.activeSolution ? this.activeSolution.feature : "";
    },

    instanceName(): string {
      return "groundTruth";
    }
  },

  methods: {
    onExport() {
      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPORT,
        details: {
          solution: this.activeSolution.solutionId,
          score: this.activeSolution.scores.map(s => ({
            metric: s.metric,
            value: s.value
          }))
        }
      });
      appActions
        .exportSolution(this.$store, {
          solutionId: this.activeSolution.solutionId
        })
        .then(err => {
          if (err) {
            // failed, this is because the wrong variable was selected
            const modal = this.$refs.exportFailModal as any;
            this.exportFailureMsg = err.message;
            modal.show();
          } else {
            const modal = this.$refs.exportSuccessModal as any;
            modal.show();
          }
        });
    },

    hideFailureModal() {
      const modal = this.$refs.exportFailModal as any;
      modal.hide();
    },

    hideSuccessModal() {
      const modal = this.$refs.exportSuccessModal as any;
      modal.hide();
      this.$router.replace(ROOT_ROUTE);
      this.$router.go(0);
    }
  }
});
</script>

<style>
.result-summaries {
  overflow-x: hidden;
  overflow-y: auto;
}

.result-summaries .facets-facet-base {
  overflow: visible;
}

.result-summaries-error {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  margin-bottom: 30px;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #007bff;
}

.check-message-container {
  display: flex;
  justify-content: flex-start;
  flex-direction: row;
  align-items: center;
}

.check-icon {
  display: flex;
  flex-shrink: 0;
  color: #00c851;
  padding-right: 15px;
}

.fail-icon {
  display: flex;
  flex-shrink: 0;
  color: #ee0701;
  padding-right: 15px;
}

.check-button {
  width: 60%;
  margin: 0 20%;
}
</style>
