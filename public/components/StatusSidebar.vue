<template>
  <div class="status-sidebar">
    <div class="status-icons">
      <div class="status-icon-wrapper" @click="onStatusIconClick(0)">
        <i class="status-icon fa fa-2x fa-info" aria-hidden="true"></i>
        <i
          v-if="isNew(variableRankingStatus)"
          class="new-update-notification fa fa-refresh fa-circle"
        ></i>
        <i
          v-if="isPending(variableRankingStatus)"
          class="new-update-notification fa fa-refresh fa-spin"
        ></i>
      </div>
      <!-- TODO
        * Disabled because the current solution is not responsive enough:
        * https://github.com/uncharted-distil/distil/issues/1815
      <div class="status-icon-wrapper" @click="onStatusIconClick(1)">
        <i
          class="status-icon fa fa-2x fa-location-arrow"
          aria-hidden="true"
        ></i>
        <i
          v-if="isNew(geocodingStatus)"
          class="new-update-notification fa fa-circle"
        ></i>
        <i
          v-if="isPending(geocodingStatus)"
          class="new-update-notification fa fa-refresh fa-spin"
        ></i>
      </div>
      -->
      <div class="status-icon-wrapper" @click="onStatusIconClick(2)">
        <i class="status-icon fa fa-2x fa-share-alt" aria-hidden="true"></i>
        <i
          v-if="isNew(clusterStatus) && !isClustered"
          class="new-update-notification fa fa-circle"
        ></i>
        <i
          v-if="isPending(clusterStatus)"
          class="new-update-notification fa fa-refresh fa-spin"
        ></i>
      </div>
      <div class="status-icon-wrapper" @click="onStatusIconClick(3)">
        <i class="status-icon fa fa-2x fa-table" aria-hidden="true"></i>
        <i
          v-if="isNew(joinSuggestionStatus)"
          class="new-update-notification fa fa-circle"
        ></i>
        <i
          v-if="isPending(joinSuggestionStatus)"
          class="new-update-notification fa fa-refresh fa-spin"
        ></i>
        <i
          v-if="isReviewd(joinSuggestionStatus) && isNew(joinDataImportStatus)"
          class="new-update-notification fa fa-circle"
        ></i>
        <i
          v-if="
            isReviewd(joinSuggestionStatus) && isPending(joinDataImportStatus)
          "
          class="new-update-notification fa fa-refresh fa-spin"
        ></i>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { actions as appActions } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import {
  DatasetPendingRequestType,
  DatasetPendingRequest,
  VariableRankingPendingRequest,
  DatasetPendingRequestStatus,
  GeocodingPendingRequest,
  JoinSuggestionPendingRequest,
  JoinDatasetImportPendingRequest,
  ClusteringPendingRequest,
  DataMode,
} from "../store/dataset/index";
import { Feature, Activity } from "../util/userEvents";
import { Dictionary } from "vue-router/types/router";

const STATUS_TYPES = [
  DatasetPendingRequestType.VARIABLE_RANKING,
  DatasetPendingRequestType.GEOCODING,
  DatasetPendingRequestType.CLUSTERING,
  DatasetPendingRequestType.JOIN_SUGGESTION,
];

export default Vue.extend({
  name: "status-sidebar",
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    pendingRequests(): DatasetPendingRequest[] {
      // pending requests for given dataset
      const updates = datasetGetters
        .getPendingRequests(this.$store)
        .filter((update) => update.dataset === this.dataset);
      return updates;
    },
    isClustered(): boolean {
      return routeGetters.getDataMode(this.$store) === DataMode.Cluster;
    },
    variableRankingRequestData(): VariableRankingPendingRequest {
      return <VariableRankingPendingRequest>(
        this.pendingRequests.find(
          (item) => item.type === DatasetPendingRequestType.VARIABLE_RANKING
        )
      );
    },

    geocodingRequestData(): GeocodingPendingRequest {
      return <GeocodingPendingRequest>(
        this.pendingRequests.find(
          (item) => item.type === DatasetPendingRequestType.GEOCODING
        )
      );
    },

    clusterRequestData(): ClusteringPendingRequest {
      return <ClusteringPendingRequest>(
        this.pendingRequests.find(
          (item) => item.type === DatasetPendingRequestType.CLUSTERING
        )
      );
    },

    joinSuggestionRequestData(): JoinSuggestionPendingRequest {
      return <JoinSuggestionPendingRequest>(
        this.pendingRequests.find(
          (item) => item.type === DatasetPendingRequestType.JOIN_SUGGESTION
        )
      );
    },

    joinDataImportRequestData(): JoinDatasetImportPendingRequest {
      const pendingRequests = datasetGetters.getPendingRequests(this.$store);
      const joinSuggestions = this.joinSuggestionRequestData.suggestions;
      const importRequest = <JoinDatasetImportPendingRequest>(
        pendingRequests.find(
          (item) => item.type === DatasetPendingRequestType.JOIN_DATASET_IMPORT
        )
      );
      const matchingDataset = joinSuggestions.find(
        (dataset) => dataset.id === (importRequest && importRequest.dataset)
      );
      return matchingDataset && importRequest;
    },

    variableRankingStatus(): DatasetPendingRequestStatus {
      return (
        this.variableRankingRequestData &&
        this.variableRankingRequestData.status
      );
    },

    geocodingStatus(): DatasetPendingRequestStatus {
      return this.geocodingRequestData && this.geocodingRequestData.status;
    },

    clusterStatus(): DatasetPendingRequestStatus {
      return this.clusterRequestData && this.clusterRequestData.status;
    },

    joinSuggestionStatus(): DatasetPendingRequestStatus {
      return (
        this.joinSuggestionRequestData && this.joinSuggestionRequestData.status
      );
    },
    joinDataImportStatus(): DatasetPendingRequestStatus {
      return (
        this.joinDataImportRequestData && this.joinDataImportRequestData.status
      );
    },
  },
  methods: {
    isNew(status) {
      return (
        status === DatasetPendingRequestStatus.RESOLVED ||
        status === DatasetPendingRequestStatus.ERROR
      );
    },
    isPending(status) {
      return DatasetPendingRequestStatus.PENDING === status;
    },
    isReviewd(status) {
      return status === DatasetPendingRequestStatus.REVIEWED;
    },
    onStatusIconClick(iconIndex) {
      const statusType = STATUS_TYPES[iconIndex];
      appActions.openStatusPanelWithContentType(this.$store, statusType);
    },
  },
});
</script>

<style scoped>
.status-sidebar {
  background-color: #fff;
  width: 55px;
  border-left: 1px solid #ccc;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.status-sidebar .status-icon-wrapper {
  padding-top: 15px;
  padding-bottom: 15px;
  text-align: center;
  position: relative;
}
.status-sidebar .status-icon {
  height: 30px;
  width: 30px;
  cursor: pointer;
}
.status-sidebar .new-update-notification {
  position: absolute;
  color: var(--red);
  top: 10px;
  right: 10px;
}
</style>
