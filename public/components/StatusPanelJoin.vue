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
  <div class="status-panel-join">
    <div class="status-message">
      <b-alert
        v-if="isImporting && importedDataset"
        :show="showStatusMessage"
        variant="info"
      >
        Importing <strong>{{ importedDataset.name }}...</strong>
      </b-alert>
      <b-alert
        v-else-if="isImportRequestResolved"
        :show="showStatusMessage"
        variant="success"
        dismissible
        @dismissed="reviewImportingRequest"
      >
        Successfully imported <strong>{{ importedDataset.name }}</strong>
      </b-alert>
      <b-alert
        v-else-if="isImportRequestError"
        :show="showStatusMessage"
        variant="danger"
        dismissible
        @dismissed="reviewImportingRequest"
      >
        Unexpected error has occured while importing
        <strong>{{ importedDataset.name }}</strong>
      </b-alert>
    </div>

    <h5 class="suggestion-heading flex-shrink-0">
      Select a dataset to join with:
    </h5>

    <div class="suggestion-list">
      <div v-if="filteredSuggestedItems.length === 0">
        No datasets are found
      </div>

      <div
        v-if="isAttemptingJoin || (isImporting && importedDataset)"
        v-html="spinnerHTML"
      />

      <b-list-group v-else>
        <b-list-group-item
          v-for="item in filteredSuggestedItems"
          :key="item.key"
          href="javascript:void()"
          :class="{ selected: item.selected }"
          :disabled="isImporting"
          @click="selectItem(item)"
        >
          <h6>{{ item.dataset.name }}</h6>
          <div class="description" v-html="item.dataset.summary" />

          <b-list-group>
            <b-list-group-item
              v-for="suggestion in item.suggestionItems"
              :key="suggestion.joinSuggestion.index"
              href="#"
              :class="{ selected: suggestion.selected }"
              :disabled="isImporting"
              @click="selectSuggestion(suggestion)"
            >
              <div class="suggested-columns">
                <b>Suggested Join Columns: </b>
                {{ suggestion.joinSuggestion.joinColumns }}
              </div>
            </b-list-group-item>
          </b-list-group>

          <footer class="item-footer">
            <ul>
              <li>
                <strong>Features:</strong>
                {{ filterVariablesByFeature(item.dataset.variables).length }}
              </li>
              <li>
                <strong>Rows:</strong>
                {{ formatNumber(item.dataset.numRows) }}
              </li>
              <li>
                <strong>Size:</strong>
                {{ formatBytes(item.dataset.numBytes) }}
              </li>
            </ul>
            <div>
              <strong>Top features:</strong>
              <ol>
                <li
                  v-for="(variable, index) in topVariablesNames(
                    item.dataset.variables
                  )"
                  :key="index"
                >
                  {{ variable }}
                </li>
              </ol>
            </div>
          </footer>
        </b-list-group-item>
      </b-list-group>
    </div>

    <div class="join-button-container flex-shrink-0 mt-3">
      <b-input
        v-model="searchQuery"
        placeholder="Refine Suggestions Via Search"
        @keydown.enter.native="refineSuggestedItems"
      />
      <b-button variant="" @click="refineSuggestedItems">
        Refine Join Suggestions
      </b-button>
      <b-button
        :disabled="!isJoinReady || isAttemptingJoin"
        variant="primary"
        @click="join"
      >
        Join
      </b-button>
    </div>

    <b-modal
      v-if="selectedDataset"
      id="join-import-modal"
      ref="import-ask-modal"
      modal-class="join-import-modal"
      title="JoinSuggestionImport"
      @ok="importDataset"
    >
      <p>
        Dataset, <strong>{{ selectedDataset.name }}</strong> is not available in
        the system. Would you like to import the dataset?
      </p>
    </b-modal>

    <b-modal
      v-model="showJoinSuccess"
      modal-class="join-preview-modal"
      cancel-disabled
      hide-header
      hide-footer
    >
      <join-datasets-preview
        :preview-table-data="previewTableData"
        :dataset-a="datasetA"
        :dataset-b="datasetB"
        :joined-column="joinedColumn"
        :search-result-index="searchResultIndex"
        @success="onJoinCommitSuccess"
        @failure="onJoinCommitFailure"
        @close="showJoinSuccess = !showJoinSuccess"
      />
    </b-modal>

    <error-modal
      :show="showJoinFailure"
      title="Join Failed"
      @close="showJoinFailure = !showJoinFailure"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import {
  Dataset,
  JoinSuggestion,
  DatasetPendingRequestType,
  DatasetPendingRequestStatus,
  JoinSuggestionPendingRequest,
  JoinDatasetImportPendingRequest,
} from "../store/dataset/index";
import JoinDatasetsPreview from "../components/JoinDatasetsPreview.vue";
import ErrorModal from "../components/ErrorModal.vue";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { formatBytes } from "../util/bytes";
import { circleSpinnerHTML } from "../util/spinner";
import {
  isDatamartProvenance,
  filterVariablesByFeature,
  topVariablesNames,
} from "../util/data";
import { loadJoinedDataset, loadJoinView } from "../util/join";

interface JoinSuggestionDatasetItem {
  dataset: Dataset;
  key: string;
  isAvailable: boolean; // tell if dataset is available in the system for join. (note. undefined implies that check hasn't made yet)
  selected: boolean;
  suggestionItems: JoinSuggestionItem[];
}

interface JoinSuggestionItem {
  joinSuggestion: JoinSuggestion;
  selected: boolean;
}

interface StatusPanelJoinState {
  suggestionDatasets: JoinSuggestionDatasetItem[];
  showStatusMessage: boolean;
  filterString: string;
  isAttemptingJoin: boolean;
  showJoinFailure: boolean;
  showJoinSuccess: boolean;
  previewTableData: any;
  datasetA: Dataset;
  datasetB: Dataset;
  datasetAColumn: any;
  datasetBColumn: any;
  searchQuery: string;
  searchResultIndex: number;
}

export default Vue.extend({
  name: "StatusPanelJoin",

  components: {
    ErrorModal,
    JoinDatasetsPreview,
  },

  data(): StatusPanelJoinState {
    return {
      datasetA: null,
      datasetAColumn: "",
      datasetB: null,
      datasetBColumn: "",
      filterString: "",
      isAttemptingJoin: false,
      previewTableData: null,
      searchResultIndex: null,
      searchQuery: "",
      showJoinFailure: false,
      showJoinSuccess: false,
      showStatusMessage: true,
      suggestionDatasets: [],
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    datasets(): Dataset[] {
      return datasetGetters.getDatasets(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    joinSuggestionRequestData(): JoinSuggestionPendingRequest {
      const request = datasetGetters
        .getPendingRequests(this.$store)
        .find(
          (request) =>
            request.dataset === this.dataset &&
            request.type === DatasetPendingRequestType.JOIN_SUGGESTION
        );
      return request as JoinSuggestionPendingRequest;
    },
    joinSuggestions(): Dataset[] {
      const joinSuggestions =
        this.joinSuggestionRequestData &&
        this.joinSuggestionRequestData.suggestions;

      if (joinSuggestions) {
        return joinSuggestions.filter((s) => s.numRows <= 100000);
      } else {
        return joinSuggestions || [];
      }
    },
    joinedColumn(): string {
      const a = this.datasetAColumn ? this.datasetAColumn : "";
      // Note: It looks like joined column name is set to same as left column (a) name
      return a;
    },
    filteredSuggestedItems(): JoinSuggestionDatasetItem[] {
      const filteredItems =
        this.filterString.length > 0 && this.suggestionDatasets.length > 0
          ? this.suggestionDatasets.filter(
              (item) =>
                item.dataset.name
                  .toLowerCase()
                  .indexOf(this.filterString.toLowerCase()) > -1 ||
                item.dataset.description
                  .toLowerCase()
                  .indexOf(this.filterString.toLowerCase()) > -1
            )
          : this.suggestionDatasets;
      return filteredItems.sort((a, b) =>
        a.dataset.name.localeCompare(b.dataset.name)
      );
    },
    joinDatasetImportRequestData(): JoinDatasetImportPendingRequest {
      // get importing request for a dataset that is in the suggestion list.
      const request = datasetGetters
        .getPendingRequests(this.$store)
        .find(
          (request) =>
            request.type === DatasetPendingRequestType.JOIN_DATASET_IMPORT
        );
      const isInSuggestionList = Boolean(
        this.joinSuggestions.find(
          (item) => item.id === (request && request.dataset)
        )
      );
      return isInSuggestionList
        ? (request as JoinDatasetImportPendingRequest)
        : undefined;
    },
    isImporting(): boolean {
      const requestStatus =
        this.joinDatasetImportRequestData &&
        this.joinDatasetImportRequestData.status;
      return requestStatus === DatasetPendingRequestStatus.PENDING;
    },
    importedItem(): JoinSuggestionDatasetItem {
      return this.suggestionDatasets.find(
        (item) => item.dataset.id === this.joinDatasetImportRequestData.dataset
      );
    },
    importedDataset(): Dataset {
      return this.importedItem && this.importedItem.dataset;
    },
    isImportRequestResolved(): boolean {
      return (
        this.joinDatasetImportRequestData &&
        this.joinDatasetImportRequestData.status ===
          DatasetPendingRequestStatus.RESOLVED
      );
    },
    isImportRequestError(): boolean {
      return (
        this.joinDatasetImportRequestData &&
        this.joinDatasetImportRequestData.status ===
          DatasetPendingRequestStatus.ERROR
      );
    },
    selectedItem(): JoinSuggestionDatasetItem {
      return this.suggestionDatasets.find((item) => item.selected);
    },
    selectedSuggestion(): JoinSuggestionItem {
      const dataset = this.suggestionDatasets.find(
        (item) => !!item.suggestionItems?.find((js) => js.selected)
      );
      if (dataset) {
        return dataset.suggestionItems.find((js) => js.selected);
      }
      return undefined;
    },
    selectedDataset(): Dataset {
      return this.selectedItem && this.selectedItem.dataset;
    },
    isJoinReady(): boolean {
      return !!this.selectedItem;
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
  },

  created() {
    this.initSuggestionItems();
  },

  beforeDestroy() {
    this.reviewImportingRequest();
  },

  methods: {
    topVariablesNames,
    filterVariablesByFeature,
    initSuggestionItems() {
      const items = this.joinSuggestions || [];
      // resolve join availablity of the importing dataset
      const isImporting = this.isImporting || this.isImportRequestResolved;
      this.suggestionDatasets = items.map((suggestion) => {
        const isImportingDataset =
          suggestion.id ===
          (this.joinDatasetImportRequestData &&
            this.joinDatasetImportRequestData.dataset);
        const isAvailable = isImportingDataset
          ? this.isImportRequestResolved
          : !isDatamartProvenance(suggestion.provenance);
        const selected = isImporting && isImportingDataset;
        const joinSuggestions = suggestion.joinSuggestion?.map((js) => {
          return {
            joinSuggestion: js,
            selected: false,
          };
        });
        return {
          dataset: suggestion,
          // There could be multiple items with same dataset id with different join suggestions.
          // So item key must be a combination of id and the join suggestions to be unique
          key:
            suggestion.id +
            (suggestion.joinSuggestion && suggestion.joinSuggestion[0]
              ? `${suggestion.joinSuggestion[0].baseColumns
                  .concat(suggestion.joinSuggestion[0].joinColumns)
                  .join("-")}`
              : ""),
          isAvailable,
          selected,
          suggestionItems: joinSuggestions,
        };
      });
    },
    refineSuggestedItems() {
      datasetActions.fetchJoinSuggestions(this.$store, {
        dataset: this.dataset,
        searchQuery: this.searchQuery,
      });
    },
    selectItem(item) {
      if (this.isImporting) {
        return;
      }
      if (this.selectedItem) {
        this.selectedItem.selected = false;
      }
      const selectedItem = item;
      selectedItem.selected = true;
    },
    selectSuggestion(suggestion) {
      if (this.isImporting) {
        return;
      }
      if (this.selectedSuggestion) {
        this.selectedSuggestion.selected = false;
      }
      const selectedSuggestion = suggestion;
      selectedSuggestion.selected = true;
    },
    join() {
      const selected = this.selectedItem;
      const currentDataset = _.find(this.datasets, (d) => {
        return d.id === this.dataset;
      });
      if (this.selectedSuggestion?.joinSuggestion?.index) {
        this.previewJoin(
          currentDataset,
          selected.dataset,
          this.selectedSuggestion.joinSuggestion.index
        );
      } else {
        loadJoinView(this.$router, this.dataset, selected.key);
      }
    },
    previewJoin(datasetA, datasetB, joinSuggestionIndex) {
      this.isAttemptingJoin = true;
      const datasetJoinInfo = {
        datasetA,
        datasetB,
        joinAccuracy: 1,
        joinSuggestionIndex: joinSuggestionIndex,
      };

      // dispatch action that triggers request send to server
      datasetActions
        .joinDatasetsPreview(this.$store, datasetJoinInfo)
        .then((tableData) => {
          // sealing the return to prevent slow, unnecessary deep reactivity.
          this.previewTableData = Object.seal(tableData.data);

          // display join preview modal
          this.isAttemptingJoin = false;
          this.showJoinSuccess = true;
          this.datasetA = datasetA;
          this.datasetB = datasetB;
          this.searchResultIndex = joinSuggestionIndex;
        })
        .catch((err) => {
          // display error modal
          this.previewTableData = null;
          this.isAttemptingJoin = false;
          this.showJoinFailure = true;
        });
    },
    importDataset(args: {
      datasetID: string;
      source: string;
      provenance: string;
    }) {
      const { id, provenance, joinSuggestion } = this.selectedDataset;
      const searchResults = joinSuggestion.map((j) => j.datasetOrigin);
      this.showStatusMessage = true;
      if (!this.isImporting) {
        datasetActions
          .importJoinDataset(this.$store, {
            datasetID: id,
            source: "contrib",
            provenance,
            searchResults,
          })
          .then((res) => {
            if (res && res.result === "ingested") {
              this.importedItem.isAvailable = true;
              this.importedDataset.source = "contrib";
            }
          });
      }
    },
    formatBytes(n: number): string {
      return formatBytes(n);
    },
    formatNumber(num: number): string {
      if (num >= 1000000000) {
        return (num / 1000000000).toFixed(1).replace(/\.0$/, "") + "B";
      }
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1).replace(/\.0$/, "") + "M";
      }
      if (num >= 1000) {
        return (num / 1000).toFixed(1).replace(/\.0$/, "") + "K";
      }
      return String(num);
    },
    reviewImportingRequest() {
      const importRequest = this.joinDatasetImportRequestData;
      if (
        importRequest &&
        importRequest.status !== DatasetPendingRequestStatus.PENDING
      ) {
        datasetActions.updatePendingRequestStatus(this.$store, {
          id: importRequest.id,
          status:
            importRequest.status === DatasetPendingRequestStatus.ERROR
              ? DatasetPendingRequestStatus.ERROR_REVIEWED
              : DatasetPendingRequestStatus.REVIEWED,
        });
      }
    },
    onJoinCommitFailure() {
      this.showJoinFailure = true;
      this.showJoinSuccess = false;
    },
    onJoinCommitSuccess(datasetID: string) {
      loadJoinedDataset(this.$router, datasetID, this.target);
    },
  },
});
</script>

<style>
.status-panel-join {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.status-panel-join .suggestion-list {
  overflow: auto;
  overflow-wrap: break-word;
}

.status-panel-join .suggestion-list .suggested-columns {
  font-size: 0.75rem;
}

.status-panel-join .item-footer {
  color: var(--color-text-second);
  display: flex;
  font-size: 0.9em;
  justify-content: space-between;
  margin-top: 1rem;
}

.status-panel-join .item-footer > * {
  flex-basis: 0;
  flex-grow: 2;
  flex-shrink: 1;
}

.status-panel-join .item-footer > *:last-child {
  flex-grow: 3;
}

.status-panel-join .item-footer ul,
.status-panel-join .item-footer ol {
  list-style: none;
  padding: 0;
  margin: 0;
}

.status-panel-join .item-footer ol {
  list-style: decimal inside;
}

.status-panel-join .item-footer ol li {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.status-panel-join .suggestion-search {
  height: 2em;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.status-panel-join .list-group-item.selected {
  background-color: #00c5e114;
}

.status-panel-join .list-group-item .description a:hover {
  color: #007bff;
  text-decoration: inherit;
}

.status-panel-join .status-message {
  min-height: 0;
  flex-shrink: 0;
  margin-top: 5px;
}

.status-panel-join .join-button-container button {
  margin-top: 5px;
  width: 100%;
}
</style>
