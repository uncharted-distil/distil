<template>
  <div class="status-panel" v-if="isOpen">
    <div class="d-flex flex-column h-100">
      <div class="heading">
        <h4 class="title">{{ contentData.title }}</h4>
        <div class="close-button" @click="close()">
          <i class="fa fa-2x fa-times" aria-hidden="true"></i>
        </div>
      </div>
      <div class="content">
        <div v-if="!requestData">There is no new update.</div>
        <div v-else-if="isPending" class="spinner">
          <div class="circle-spinner"></div>
        </div>
        <div v-else-if="isResolved" class="h-100">
          <status-panel-join v-if="this.statusType === 'JOIN_SUGGESTION'" />
          <div v-else>
            <div>
              <p>
                {{ contentData.resolvedMsg }}
              </p>
            </div>
            <b-button variant="primary" @click="applyChange">Apply</b-button>
            <b-button variant="secondary" @click="clearData">Discard</b-button>
          </div>
        </div>
        <div v-else-if="isError">
          <div>
            <p>
              {{ contentData.errorMsg }}
            </p>
          </div>
          <b-button variant="secondary" @click="clearData">Ok</b-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import StatusPanelJoin from "../components/StatusPanelJoin.vue";
import {
  DatasetPendingRequest,
  DatasetPendingRequestType,
  VariableRankingPendingRequest,
  DatasetPendingRequestStatus,
  GeocodingPendingRequest,
  SummaryMode,
  DataMode,
  OutlierPendingRequest,
} from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import {
  actions as appActions,
  getters as appGetters,
} from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { StatusPanelState, StatusPanelContentType } from "../store/app";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { overlayRouteEntry, varModesToString } from "../util/routes";
import { IMAGE_TYPE, isClusterType } from "../util/types";
import { $enum } from "ts-enum-util";

const STATUS_USER_EVENT = new Map<DatasetPendingRequestType, Feature>([
  [DatasetPendingRequestType.VARIABLE_RANKING, Feature.RANK_FEATURES],
  [DatasetPendingRequestType.GEOCODING, Feature.GEOCODE_FEATURES],
  [DatasetPendingRequestType.CLUSTERING, Feature.CLUSTER_DATA],
  [DatasetPendingRequestType.JOIN_SUGGESTION, Feature.JOIN_DATASETS],
  [DatasetPendingRequestType.OUTLIER, Feature.OUTLIER_FEATURES],
]);

export default Vue.extend({
  name: "StatusPanel",

  components: {
    StatusPanelJoin,
  },

  mounted: function () {
    if (routeGetters.getRouteIsClusterGenerated(this.$store)) {
      this.applyClusteringChange();
    }
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    statusPanelState(): StatusPanelState {
      return appGetters.getStatusPanelState(this.$store);
    },
    isOpen(): boolean {
      return this.statusPanelState.isOpen;
    },
    statusType(): StatusPanelContentType {
      return this.statusPanelState.contentType;
    },
    requestData(): DatasetPendingRequest {
      const request = datasetGetters
        .getPendingRequests(this.$store)
        .find(
          (request) =>
            request.dataset === this.dataset && request.type === this.statusType
        );
      return request;
    },
    isPending(): boolean {
      return this.requestData.status === DatasetPendingRequestStatus.PENDING;
    },
    isResolved(): boolean {
      return (
        this.requestData.status === DatasetPendingRequestStatus.RESOLVED ||
        this.requestData.status === DatasetPendingRequestStatus.REVIEWED
      );
    },
    isError(): boolean {
      return (
        this.requestData.status === DatasetPendingRequestStatus.ERROR ||
        this.requestData.status === DatasetPendingRequestStatus.ERROR_REVIEWED
      );
    },
    contentData(): {
      title: string;
      pendingMsg?: string;
      resolvedMsg?: string;
      defaultMsg?: string;
      errorMsg?: string;
    } {
      switch (this.statusType) {
        case DatasetPendingRequestType.VARIABLE_RANKING:
          return {
            title: "Variable Ranking",
            pendingMsg: "Computing variable rankings...",
            resolvedMsg:
              "Variable ranking has been updated. Would you like to apply the changes to the feature list?",
            errorMsg:
              "Unexpected error has happened while calculating variable rankings",
          };
        case DatasetPendingRequestType.GEOCODING:
          return {
            title: "Geo Coding",
            pendingMsg: "Geocoding place names...",
            resolvedMsg:
              "Geocoding has been processed. Would you like to apply the change to the feature list?",
            errorMsg: "Unexpected error has happened while geocoding",
          };
        case DatasetPendingRequestType.CLUSTERING:
          return {
            title: "Data Clustering",
            pendingMsg: "Computing data clusters...",
            resolvedMsg:
              "Data clusters have been generated. Would you like to apply the change to the dataset?",
            errorMsg: "Unexpected error has happened while clustering",
          };
        case DatasetPendingRequestType.JOIN_SUGGESTION:
          return {
            title: "Join Suggestion",
            pendingMsg: "Compuing join suggestions...",
            errorMsg:
              "Unexpected error has happened while retreving join suggestions",
          };
        case DatasetPendingRequestType.OUTLIER:
          return {
            title: "Variable Outliers",
            pendingMsg: "Computing variable outliers...",
            resolvedMsg:
              "Variable outliers has been processed. Would you like to apply the changes to the feature list?",
            errorMsg:
              "Unexpected error has happened while calculating variable outliers",
          };
        default:
          return {
            title: "",
          };
      }
    },
  },

  methods: {
    clearData() {
      if (this.requestData) {
        datasetActions.removePendingRequest(this.$store, this.requestData.id);
      }
    },

    close() {
      if (
        this.requestData &&
        this.requestData.status !== DatasetPendingRequestStatus.PENDING
      ) {
        datasetActions.updatePendingRequestStatus(this.$store, {
          id: this.requestData.id,
          status:
            this.requestData.status === DatasetPendingRequestStatus.ERROR
              ? DatasetPendingRequestStatus.ERROR_REVIEWED
              : DatasetPendingRequestStatus.REVIEWED,
        });
      }
      appActions.closeStatusPanel(this.$store);
    },

    applyChange() {
      switch (this.statusType) {
        case DatasetPendingRequestType.VARIABLE_RANKING:
          this.applyVariableRankingChange();
          break;
        case DatasetPendingRequestType.GEOCODING:
          this.applyGeocodingChange();
          break;
        case DatasetPendingRequestType.OUTLIER:
          this.applyOutlierChange();
          break;
        case DatasetPendingRequestType.CLUSTERING:
          this.applyClusteringChange();
          break;
        default:
      }

      // Log changes
      const status = STATUS_USER_EVENT.get(this.statusType);
      appActions.logUserEvent(this.$store, {
        feature: status,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });
    },

    applyVariableRankingChange() {
      const variableRequest = <VariableRankingPendingRequest>this.requestData;
      datasetActions.updateVariableRankings(this.$store, {
        dataset: variableRequest.dataset,
        rankings: variableRequest.rankings,
      });

      // Update the route to know that the training variables have been ranked.
      const varRankedEntry = overlayRouteEntry(this.$route, {
        varRanked: "1",
      });
      this.$router.push(varRankedEntry).catch((err) => console.warn(err));

      this.clearData();
    },

    applyGeocodingChange() {
      const geoRequest = <GeocodingPendingRequest>this.requestData;
      datasetActions
        .fetchGeocodingResults(this.$store, {
          dataset: geoRequest.dataset,
          field: geoRequest.field,
        })
        .then(() => {
          this.clearData();
        });
    },

    applyOutlierChange() {
      const { dataset, variable } = <OutlierPendingRequest>this.requestData;
      datasetActions
        .fetchOutliers(this.$store, { dataset, variable })
        .then(() => {
          this.clearData();
        });
    },

    // Applies clustering changes and refetches update variable summaries
    applyClusteringChange() {
      // fetch the var modes map
      const varModesMap = routeGetters.getDecodedVarModes(this.$store);

      // find any grouped vars that are using this cluster data and update their
      // mode to cluster now that data is available
      datasetGetters
        .getGroupings(this.$store)
        .filter((v) => isClusterType(v.colType))
        .forEach((v) => {
          varModesMap.set(v.key, SummaryMode.Cluster);
        });

      // find any image variables using this cluster data and update their mode
      datasetGetters
        .getVariables(this.$store)
        .filter((v) => v.colType === IMAGE_TYPE)
        .forEach((v) => {
          varModesMap.set(v.key, SummaryMode.Cluster);
        });

      // serialize the modes map into a string and add to the route
      // and update to know that the clustering has been applied.
      const varModesStr = varModesToString(varModesMap);
      const entry = overlayRouteEntry(this.$route, {
        varModes: varModesStr,
        dataMode: DataMode.Cluster,
        clustering: "1",
      });
      this.$router.push(entry).catch((err) => console.warn(err));

      // update variables
      // pull the updated dataset, vars, and summaries
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        this.$store
      );
      const highlight = routeGetters.getDecodedHighlight(this.$store);
      for (const [k, v] of varModesMap) {
        datasetActions.fetchVariableSummary(this.$store, {
          dataset: this.dataset,
          variable: k,
          highlight: highlight,
          filterParams: filterParams,
          include: true,
          dataMode: DataMode.Cluster,
          mode: $enum(SummaryMode).asValueOrDefault(v, SummaryMode.Default),
        });
      }

      this.clearData();
    },
  },
});
</script>

<style scoped>
.status-panel {
  box-shadow: 0 4px 5px 0 rgba(0, 0, 0, 0.14), 0 1px 10px 0 rgba(0, 0, 0, 0.12),
    0 2px 4px -1px rgba(0, 0, 0, 0.2);
  display: block;
  position: fixed;
  right: 0;
  top: 0;
  bottom: 0;
  z-index: 1040; /* To go above the NavBar */
  width: 300px;
  height: 100%;
  background: #fff;
}

.status-panel .heading {
  height: 58px;
  flex-shrink: 0;
  border-bottom: 1px solid #f1f3f4;
  display: flex;
  align-items: center;
  padding-right: 10px;
  padding-left: 10px;
}

.status-panel .content {
  padding: 10px;
}

.status-panel .title {
  margin: 0;
  flex-grow: 1;
}

.status-panel .spinner {
  display: flex;
  justify-content: center;
  padding-top: 10px;
}

.status-panel .close-button {
  cursor: pointer;
}
</style>
