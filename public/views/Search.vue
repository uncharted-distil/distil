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
  <div class="container-fluid d-flex flex-column h-100 search-view">
    <!-- Spacer for the App.vue <navigation> component. -->
    <div class="row flex-0-nav" />

    <!-- Add dataset modal -->
    <add-dataset
      id="add-dataset"
      @uploadstart="onUploadStart"
      @uploadfinish="onUploadFinish"
    />
    <div class="row justify-content-center">
      <import-status
        class="file-uploader-status col-12"
        :status="uploadStatus"
        :import-response="importResponse"
        :name="uploadData.name"
        :dataset-id="uploadData.datasetID"
        @importfull="onReImportFullDataset"
      />
    </div>

    <!-- Search navigation -->
    <section class="row justify-content-center mt-5">
      <div class="col-12 col-md-11 col-lg-10 col-xl-8">
        <div class="d-flex justify-content-between align-items-end">
          <b-tabs active-nav-item-class="active-search-tab">
            <b-tab @click="onTab(0)">
              <template #title>
                <i class="fa fa-connectdevelop color-black" />
                <b class="color-black">Models</b>
                <span
                  class="badge badge-pill badge-light background-color-grey"
                >
                  {{ nbSearchModels }}
                </span>
              </template>
            </b-tab>
            <b-tab active @click="onTab(1)">
              <template #title>
                <i class="fa fa-table color-black" />
                <b class="color-black">Datasets</b>
                <span
                  class="badge badge-pill badge-light background-color-grey"
                >
                  {{ nbSearchDatasets }}
                </span>
              </template>
            </b-tab>
          </b-tabs>
          <b-button
            v-b-modal.add-dataset
            class="add-new-datasets mb-2"
            variant="secondary"
          >
            <i class="fa fa-plus-circle" /> Add Dataset
          </b-button>
        </div>
      </div>
    </section>

    <!-- Main view. -->
    <section class="row flex-1 justify-content-center">
      <div class="col-12 col-md-11 col-lg-10 col-xl-8 search-content-wrapper">
        <!-- Search bar -->
        <section class="row">
          <div class="col-8">
            <search-bar class="search-search-bar" />
          </div>
          <div class="col-3">
            <div class="input-group search-search-bar">
              <div class="input-group-prepend">
                <div class="input-group-text">Sort By:</div>
              </div>
              <select id="inputState" v-model="sortType" class="form-control">
                <option selected>Name Ascending</option>
                <option>Name Descending</option>
                <option>Features Ascending</option>
                <option>Features Descending</option>
                <option>Imported Ascending</option>
                <option>Imported Descending</option>
              </select>
            </div>
          </div>
        </section>
        <div
          v-if="isPending"
          class="search-content-spinner"
          v-html="spinnerHTML"
        />
        <p v-else-if="isSearchResultsEmpty" class="search-content-empty">
          No {{ tab === "all" ? "datasets or models" : tab }} found for search
        </p>
        <div v-else class="search-content">
          <dataset-preview-table
            v-if="currentTab"
            class="mt-3"
            :datasets="sortedDatasetResults"
            @dataset-delete="onDatasetDeletionClicked"
          />
          <model-preview-table
            v-if="!currentTab"
            class="mt-3"
            :models="sortedModelResults"
            @model-delete="onModelDeletionClicked"
          />
        </div>
        <deletion-modal
          :target="deletionTarget"
          @model-delete="onDeletionConfirmed"
        />
      </div>
    </section>

    <!-- Version of TA2 and TA3 -->
    <footer class="version" v-html="version" />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import AddDataset from "../components/AddDataset.vue";
import DeletionModal from "../components/DeletionModal.vue";
import ImportStatus from "../components/ImportStatus.vue";
import DatasetPreviewTable from "../components/searchComponents/DatasetPreviewTable.vue";
import ModelPreviewTable from "../components/searchComponents/ModelPreviewTable.vue";
import SearchBar from "../components/SearchBar.vue";
import { Dataset } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { actions as modelActions } from "../store/model/module";
import { Model } from "../store/model/index";
import { getters as appGetters } from "../store/app/module";
import { getters as modelGetters } from "../store/model/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";
import { spinnerHTML } from "../util/spinner";

interface SearchResult {
  type: string;
  name: string;
  storageName: string;
  features: number;
}

interface ModelResult extends SearchResult {
  model: Model;
}

interface DatasetResult extends SearchResult {
  dataset: Dataset;
}

export default Vue.extend({
  name: "SearchView",

  components: {
    AddDataset,
    DatasetPreviewTable,
    ImportStatus,
    ModelPreviewTable,
    SearchBar,
    DeletionModal,
  },

  data() {
    return {
      sortType: "Name Ascending",
      isPending: false,
      sorting: {
        asc: true,
        type: "name",
      },
      tab: "datasets",
      uploadData: {
        datasetID: "",
      },
      uploadStatus: "",
      importResponse: {
        dataset: "",
        location: "",
      },
      deletionTarget: "",
      deletionInfo: null,
      deletionInfoModel: null,
      currentTab: 1,
    };
  },

  computed: {
    filteredDatasets(): Dataset[] {
      return datasetGetters.getFilteredDatasets(this.$store);
    },

    filteredDatasetsIds(): Set<string> {
      const ids = this.filteredDatasets.map((dataset) => dataset.id);
      return new Set(ids);
    },

    filteredModels(): Model[] {
      const models = modelGetters.getFilteredModels(this.$store);

      // Only display the models using dataset that the search bar has found.
      return models.filter((model) =>
        this.filteredDatasetsIds.has(model.datasetId)
      );
    },

    nbSearchDatasets(): number {
      return this.filteredDatasets.length ?? 0;
    },

    nbSearchModels(): number {
      return this.filteredModels.length ?? 0;
    },

    /* List of search results to be displayed. */
    searchResults(): (ModelResult | DatasetResult)[] {
      const results = [] as (ModelResult | DatasetResult)[];

      // If tab is either 'models' or 'all' we display the models.
      if (!this.currentTab) {
        const models = this.filteredModels.map((model) => {
          return {
            type: "model",
            name: model.modelName.toUpperCase(),
            storageName: model.modelName.toUpperCase(),
            features: model.variables.length ?? 0,
            model,
          };
        });
        results.push(...models);
      }

      // If tab is either 'datasets' or 'all' we display the datasets.
      if (this.currentTab) {
        const datasets = this.filteredDatasets.map((dataset) => {
          return {
            type: "dataset",
            name: dataset.name.toUpperCase(),
            storageName: dataset.storageName,
            features: dataset.variables.length ?? 0,
            dataset: dataset,
          };
        });
        results.push(...datasets);
      }

      return results;
    },
    sortedDatasetResults(): Dataset[] {
      return (this.sortedResults.filter(
        (d) => d.type === "dataset"
      ) as DatasetResult[]).map((d) => d.dataset);
    },
    sortedModelResults(): Model[] {
      return (this.sortedResults.filter(
        (d) => d.type === "model"
      ) as ModelResult[]).map((d) => d.model);
    },
    /* Sort the results based on the sorting selected. */
    sortedResults(): (ModelResult | DatasetResult)[] {
      return this.searchResults.slice().sort((a, b) => {
        // Sort by recent activity
        // if (this.sorting.type === "recent") {
        // ...
        // }

        // Sort by name
        if (this.sorting.type === "name") {
          return this.sorting.asc
            ? a.name.localeCompare(b.name)
            : b.name.localeCompare(a.name);
        }

        // Sort by features
        if (this.sorting.type === "features") {
          return this.sorting.asc
            ? a.features - b.features
            : b.features - a.features;
        }

        // Sort by import state
        if (this.sorting.type === "imported") {
          // reverse order because we want empty labels to be sorted last not first
          return this.sorting.asc
            ? b.storageName.localeCompare(a.storageName)
            : a.storageName.localeCompare(b.storageName);
        }
      });
    },

    isSearchResultsEmpty(): boolean {
      return _.isEmpty(this.searchResults);
    },

    spinnerHTML(): string {
      return spinnerHTML();
    },

    terms(): string {
      return routeGetters.getRouteTerms(this.$store);
    },

    /* Font Awesome class for the soring dropdown. */
    sortingIcon(): string {
      let type = "amount";
      if (this.sorting.type === "name") {
        type = "alpha";
      }
      if (this.sorting.type === "features") {
        type = "numeric";
      }
      const asc = this.sorting.asc ? "asc" : "desc";
      return `fa fa-sort-${type}-${asc}`;
    },

    /* Dropdown name to be displayed. */
    sortingDisplayName(): string {
      if (this.sorting.type !== "recent") {
        return _.capitalize(this.sorting.type);
      }
      return "Recent Activity";
    },

    // Display the version numer of the app.
    version(): string {
      return appGetters
        .getAllSystemVersions(this.$store)
        .replace(/\n/gi, "<br>");
    },
  },

  watch: {
    terms() {
      this.fetch();
    },
    sortType() {
      this.onSort(this.sortType);
    },
  },

  beforeMount() {
    this.fetch();
    viewActions.clearAllData(this.$store);
  },

  methods: {
    onTab(tab: number) {
      this.currentTab = tab;
    },
    onSort(val: string) {
      switch (val) {
        case "Name Ascending":
          this.sortNameAsc();
          break;
        case "Name Descending":
          this.sortNameDesc();
          break;
        case "Features Ascending":
          this.sortFeaturesAsc();
          break;
        case "Features Descending":
          this.sortFeaturesDesc();
          break;
        case "Imported Ascending":
          this.sortImportedAsc();
          break;
        case "Imported Descending":
          this.sortImportedDesc();
          break;
      }
    },
    fetch() {
      this.isPending = true;
      viewActions.fetchSearchData(this.$store).then(() => {
        this.isPending = false;
      });
    },
    onDatasetDeletionClicked(dataset: Dataset) {
      this.deletionTarget = dataset.name;
      this.deletionInfo = dataset;
    },
    onModelDeletionClicked(model: Model) {
      this.deletionTarget = model.modelName;
      this.deletionInfoModel = model;
    },
    onDeletionConfirmed() {
      const terms = routeGetters.getRouteTerms(this.$store);
      if (this.deletionInfo != null) {
        datasetActions.deleteDataset(this.$store, {
          dataset: this.deletionInfo.id,
          terms: terms,
        });
        this.deletionInfo = null;
      } else if (this.deletionInfoModel != null) {
        modelActions.deleteModel(this.$store, {
          model: this.deletionInfoModel.fittedSolutionId,
          terms: terms,
        });
        this.deletionInfoModel = null;
      }
    },
    onUploadStart(uploadData) {
      this.uploadData = uploadData;
      this.uploadStatus = "started";
    },

    onUploadFinish(err, response) {
      datasetActions.fetchDataset(this.$store, {
        dataset: response.dataset,
      });
      datasetActions.searchDatasets(
        this.$store,
        routeGetters.getRouteTerms(this.$store)
      );
      this.uploadStatus = err ? "error" : "success";
      this.importResponse = response;
    },

    // The dataset will be reimported without sampling.
    async onReImportFullDataset() {
      const path = this.importResponse.location;
      const datasetID = this.uploadData.datasetID;

      // Test that we have the dataset ID and location of the raw file.
      if (_.isEmpty(datasetID) && _.isEmpty(path)) {
        return;
      }

      try {
        this.uploadStatus = "started";
        this.uploadData.datasetID = "";

        this.importResponse = await datasetActions.importFullDataset(
          this.$store,
          {
            datasetID,
            path,
          }
        );
        this.uploadStatus = "success";

        // The dataset as been imported as a new dataset.
        this.uploadData.datasetID = this.importResponse.dataset ?? datasetID;
      } catch (error) {
        this.uploadStatus = "error";
      }
    },

    sortRecentDesc() {
      this.sorting = { asc: false, type: "recent" };
    },
    sortNameAsc() {
      this.sorting = { asc: true, type: "name" };
    },
    sortNameDesc() {
      this.sorting = { asc: false, type: "name" };
    },
    sortFeaturesAsc() {
      this.sorting = { asc: true, type: "features" };
    },
    sortFeaturesDesc() {
      this.sorting = { asc: false, type: "features" };
    },
    sortImportedAsc() {
      this.sorting = { asc: true, type: "imported" };
    },
    sortImportedDesc() {
      this.sorting = { asc: false, type: "imported" };
    },
  },
});
</script>

<style scoped>
.row .file-uploader-status {
  padding: 0;
}

.search-search-bar {
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
  margin-top: 1rem;
  width: 100%;
}

.close-join-button {
  position: absolute;
  top: 4px;
  right: 4px;
  cursor: pointer;
}

.join-datasets-button,
.join-datasets-button i {
  line-height: 32px !important;
  text-align: center;
}

/* Navigation */

.search-nav {
  align-items: center;
  display: flex;
  padding: 1rem;
}

.search-nav > * + * {
  margin-left: 2em;
}
.background-color-grey {
  background-color: #eee;
}
.search-nav-tab {
  background: #eeeeee;
  border-color: transparent;
  border-style: solid;
  border-width: 0 0 3px 0;
  padding: 0.25em 0;
}

.search-nav-tab.active {
  border-bottom-color: var(--blue);
}

.search-nav .add-new-datasets {
  margin-left: auto; /* Align to the right of the navigation. */
}

/* Content */

.search-content-wrapper {
  /* As we use flexbox with .row, the height needs to be define
     here to allow .search-content to be scrollable. */
  height: 95%;
  background-color: white;
}

.search-content {
  height: 95%;
  overflow: none;
}

.search-content .card-result {
  margin-left: 0;
  margin-right: 0;
}
.color-black {
  color: #757575;
}
.search-content-empty,
.search-content-spinner {
  margin-top: 3rem;
  text-align: center;
}

.search-content-empty {
  color: var(--black);
  font-size: 1.2em;
  font-weight: bold;
  line-height: 1.2;
}

/* Version */
.version {
  background-color: rgba(255, 255, 255, 0.8);
  bottom: 0;
  color: var(--gray-500);
  font-size: 0.75rem;
  padding: 1em;
  pointer-events: none;
  position: absolute;
  right: 0;
}
</style>
<style>
.active-search-tab {
  background-color: white !important;
  border-bottom: none !important;
  border-top: 5px solid #424242 !important;
}
</style>
