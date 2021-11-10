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
  <div v-if="isOpen" class="status-panel" :class="{ wider: isWider }">
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
            <b-form-group label="Number of clusters">
              <b-form-select
                v-model="selectedNumberOfClusters"
                :options="availableNumberOfClusters"
              />
            </b-form-group>
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
import { EventList } from "../util/events";
import { DataExplorerRef } from "../util/componentTypes";

const STATUS_USER_EVENT = new Map<DatasetPendingRequestType, Feature>([
  [DatasetPendingRequestType.VARIABLE_RANKING, Feature.RANK_FEATURES],
  [DatasetPendingRequestType.GEOCODING, Feature.GEOCODE_FEATURES],
  [DatasetPendingRequestType.CLUSTERING, Feature.CLUSTER_DATA],
  [DatasetPendingRequestType.JOIN_SUGGESTION, Feature.JOIN_DATASETS],
  [DatasetPendingRequestType.OUTLIER, Feature.OUTLIER_FEATURES],
]);

export default Vue.extend({
  name: "StatusPanel",

  data() {
    return {
      selectedNumberOfClusters: null,
    };
  },
  components: {
    StatusPanelJoin,
  },
  props: {
    dataset: {
      type: String as () => string,
      default: "",
    },
  },
  mounted: function () {
    if (routeGetters.getRouteIsClusterGenerated(this.$store)) {
      this.$eventBus.$emit(EventList.VARIABLES.APPLY_CLUSTER_EVENT);
    }
  },

  computed: {
    statusPanelState(): StatusPanelState {
      return appGetters.getStatusPanelState(this.$store);
    },
    isOpen(): boolean {
      return this.statusPanelState.isOpen;
    },

    /* Display the status-panel wider. */
    isWider(): boolean {
      return this.isResolved && this.statusType === "JOIN_SUGGESTION";
    },

    statusType(): StatusPanelContentType {
      return this.statusPanelState.contentType;
    },
    requestData(): DatasetPendingRequest {
      return datasetGetters
        .getPendingRequests(this.$store)
        .find(
          (request) =>
            request.dataset === this.dataset && request.type === this.statusType
        );
    },
    isPending(): boolean {
      return this.requestData?.status === DatasetPendingRequestStatus.PENDING;
    },
    isResolved(): boolean {
      return (
        this.requestData?.status === DatasetPendingRequestStatus.RESOLVED ||
        this.requestData?.status === DatasetPendingRequestStatus.REVIEWED
      );
    },
    isError(): boolean {
      return (
        this.requestData?.status === DatasetPendingRequestStatus.ERROR ||
        this.requestData?.status === DatasetPendingRequestStatus.ERROR_REVIEWED
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
    availableNumberOfClusters(): Array<Object> {
      const minNumber = 3;
      const maxNumber = 10;

      const returnNumberOfClusterOptions = [];

      for (let i = minNumber; i <= maxNumber; i++) {
        returnNumberOfClusterOptions.push({
          value: i,
          text: i,
        });
      }

      return returnNumberOfClusterOptions;
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
          this.$eventBus.$emit(EventList.VARIABLES.APPLY_OUTLIER_EVENT);
          this.clearData();
          this.close();
          break;
        case DatasetPendingRequestType.CLUSTERING:
          this.$eventBus.$emit(EventList.VARIABLES.APPLY_CLUSTER_EVENT);
          this.clearData();
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
      const { dataset, rankings } = variableRequest;
      datasetActions.updateVariableRankings(this.$store, { dataset, rankings });

      // Update the route to know that the training variables have been ranked.
      const entry = overlayRouteEntry(this.$route, { varRanked: "1" });
      this.$router.push(entry).catch((err) => console.warn(err));

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
  },

  watch: {
    selectedNumberOfClusters() {
      const self = (this.$root.$refs.view as unknown) as DataExplorerRef;
      self.clusterCount = this.selectedNumberOfClusters;
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

.status-panel.wider {
  width: 40vw;
  max-width: 450px;
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
