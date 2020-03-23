<template>
  <div class="prediction-summaries">
    <p class="nav-link font-weight-bold">Results</p>
    <p></p>
    <p class="nav-link font-weight-bold">
      Predictions for Model
    </p>

    <div v-for="summary in summaries" :key="summary.key">
      <facet-entry
        enable-highlighting
        :summary="summary"
        :highlight="highlight"
        :enabled-type-changes="[]"
        :row-selection="rowSelection"
        :instanceName="instanceName"
      >
      </facet-entry>
    </div>

    <!-- TODO: For show right now.-->
    <b-button block variant="primary">
      Export Predictions
    </b-button>

    <b-modal id="export" title="Export">
      <div class="check-message-container">
        <i class="fa fa-check-circle fa-3x check-icon"></i>
        <div>
          This action will export predictions and return to the application
          start screen.
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import FacetEntry from "../components/FacetEntry";
import FileUploader from "../components/FileUploader";
import { getSolutionById } from "../util/solutions";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as predictionsGetters } from "../store/predictions/module";
import {
  actions as appActions,
  getters as appGetters
} from "../store/app/module";
import store from "../store/store";
import {
  EXPORT_SUCCESS_ROUTE,
  ROOT_ROUTE,
  PREDICTION_ROUTE
} from "../store/route/index";
import {
  Variable,
  TaskTypes,
  VariableSummary,
  Highlight,
  RowSelection
} from "../store/dataset/index";
import Vue from "vue";
import { Solution } from "../store/requests/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { createRouteEntry } from "../util/routes";
import { PREDICTION_UPLOAD } from "../util/uploads";
import { sum } from "d3";

export default Vue.extend({
  name: "prediction-summaries",

  components: {
    FacetEntry,
    FileUploader
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      exportFailureMsg: "",
      file: null,
      uploadData: {},
      uploadStatus: "",
      uploadType: PREDICTION_UPLOAD
    };
  },

  computed: {
    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },

    fittedSolutionId(): string {
      const predictions = requestGetters
        .getPredictions(this.$store)
        .find(p => p.requestId === this.produceRequestId);
      return predictions.fittedSolutionId || "";
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    targetType(): string {
      const targetName = this.target;
      const variables = datasetGetters.getVariables(this.$store);
      return variables.find(v => v.colName === targetName).colType;
    },

    instanceName(): string {
      return "predictions";
    },

    summaries(): VariableSummary[] {
      return predictionsGetters
        .getPredictionSummaries(this.$store)
        .filter(s => s.solutionId === this.solutionId);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    }
  }
});
</script>

<style>
.prediction-summaries {
  margin-bottom: 12px;
  overflow-x: hidden;
  overflow-y: auto;
}

.prediction-summaries .facets-facet-base {
  overflow: visible;
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
