<template>
  <div class="container-fluid d-flex flex-column h-100 search-view">
    <!-- Spacer for the App.vue <navigation> component. -->
    <div class="row flex-0-nav"></div>

    <!-- Title of the page. -->
    <header class="header row justify-content-center">
      <div class="col-12 col-md-10">
        <h5 class="header-title">
          Select a Model to reuse or a Dataset to create a model
        </h5>
      </div>
    </header>

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

    <!-- Search bar -->
    <section class="row justify-content-center">
      <div class="col-12 col-md-8">
        <search-bar class="search-search-bar"></search-bar>
      </div>
    </section>

    <!-- Search navigation -->
    <section class="row justify-content-center">
      <div class="col-12 col-md-8">
        <nav class="search-nav">
          <button
            class="search-nav-tab"
            @click="tab = 'all'"
            :class="{ active: tab === 'all' }"
          >
            All
            <span class="badge badge-pill badge-danger">{{
              nbSearchModels + nbSearchDatasets
            }}</span>
          </button>
          <button
            class="search-nav-tab"
            @click="tab = 'models'"
            :class="{ active: tab === 'models' }"
          >
            <i class="fa fa-connectdevelop"></i> Models
            <span class="badge badge-pill badge-danger">{{
              nbSearchModels
            }}</span>
          </button>
          <button
            class="search-nav-tab"
            @click="tab = 'datasets'"
            :class="{ active: tab === 'datasets' }"
          >
            <i class="fa fa-table"></i> Datasets
            <span class="badge badge-pill badge-danger">{{
              nbSearchDatasets
            }}</span>
          </button>
          <b-dropdown variant="outline-dark" size="sm">
            <template v-slot:button-content>
              <i :class="sortingIcon"></i> Sort by:
              <strong>{{ sortingDisplayName }}</strong>
            </template>
            <!-- <b-dropdown-item-button @click="sortRecentDesc">
              <i class="fa fa-sort-amount-desc"></i> Recent Activity
            </b-dropdown-item-button>
            <b-dropdown-divider></b-dropdown-divider> -->
            <b-dropdown-item-button @click="sortNameAsc">
              <i class="fa fa-sort-alpha-asc"></i> Name
            </b-dropdown-item-button>
            <b-dropdown-item-button @click="sortNameDesc">
              <i class="fa fa-sort-alpha-desc"></i> Name
            </b-dropdown-item-button>
            <b-dropdown-divider></b-dropdown-divider>
            <b-dropdown-item-button @click="sortFeaturesAsc">
              <i class="fa fa-sort-numeric-asc"></i> Features
            </b-dropdown-item-button>
            <b-dropdown-item-button @click="sortFeaturesDesc">
              <i class="fa fa-sort-numeric-desc"></i> Features
            </b-dropdown-item-button>
          </b-dropdown>
          <b-button
            class="add-new-datasets"
            variant="primary"
            v-b-modal.add-dataset
          >
            <i class="fa fa-plus-circle" /> Add Dataset
          </b-button>
        </nav>
      </div>
    </section>

    <!-- Main view. -->
    <section class="row flex-1 justify-content-center">
      <div class="col-12 col-md-8 search-content-wrapper">
        <div
          v-if="isPending"
          v-html="spinnerHTML"
          class="search-content-spinner"
        ></div>
        <p v-else-if="isSearchResultsEmpty" class="search-content-empty">
          No {{ tab === "all" ? "datasets or models" : tab }} found for search
        </p>
        <div v-else class="search-content">
          <template v-for="result in sortedResults">
            <dataset-preview
              v-if="result.type === 'dataset'"
              :key="result.dataset.id"
              :dataset="result.dataset"
              allow-join
              allow-import
            />
            <model-preview
              v-if="result.type === 'model'"
              :key="result.model.fittedSolutionId"
              :model="result.model"
            />
          </template>
        </div>
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
import DatasetPreview from "../components/DatasetPreview.vue";
import ImportStatus from "../components/ImportStatus.vue";
import ModelPreview from "../components/ModelPreview.vue";
import SearchBar from "../components/SearchBar.vue";
import { Dataset } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { Model } from "../store/model/index";
import { getters as appGetters } from "../store/app/module";
import { getters as modelGetters } from "../store/model/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";
import { spinnerHTML } from "../util/spinner";

export default Vue.extend({
  name: "SearchView",

  components: {
    AddDataset,
    DatasetPreview,
    ImportStatus,
    ModelPreview,
    SearchBar,
  },

  data() {
    return {
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
    };
  },

  computed: {
    filteredDatasets(): Dataset[] {
      return datasetGetters.getFilteredDatasets(this.$store);
    },

    filteredDatasetsIds(): Set<String> {
      const ids = this.filteredDatasets.map((dataset) => dataset.id);
      return new Set(ids);
    },

    filteredModels(): Model[] {
      const models = modelGetters.getModels(this.$store);

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
    searchResults(): any[] {
      const results = [] as any[];

      // If tab is either 'models' or 'all' we display the models.
      if (this.tab !== "datasets") {
        const models = this.filteredModels.map((model) => {
          return { type: "model", model };
        });
        results.push(...models);
      }

      // If tab is either 'datasets' or 'all' we display the datasets.
      if (this.tab !== "models") {
        const datasets = this.filteredDatasets.map((dataset) => {
          return { type: "dataset", dataset };
        });
        results.push(...datasets);
      }

      return results;
    },

    /* Sort the results based on the sorting selected. */
    sortedResults(): any[] {
      return this.searchResults.slice().sort((a, b) => {
        // Sort by recent activity
        // if (this.sorting.type === "recent") {
        // ...
        // }

        // Sort by name
        if (this.sorting.type === "name") {
          a = this.getNameFromResult(a);
          b = this.getNameFromResult(b);
          return this.sorting.asc ? a.localeCompare(b) : b.localeCompare(a);
        }

        // Sort by features
        if (this.sorting.type === "features") {
          a = this.getFeatureFromResult(a);
          b = this.getFeatureFromResult(b);
          return this.sorting.asc ? a - b : b - a;
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
  },

  beforeMount() {
    this.fetch();
  },

  methods: {
    fetch() {
      this.isPending = true;
      viewActions.fetchSearchData(this.$store).then(() => {
        this.isPending = false;
      });
    },

    onUploadStart(uploadData) {
      this.uploadData = uploadData;
      this.uploadStatus = "started";
    },

    onUploadFinish(err, response) {
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

    getNameFromResult(result: any) {
      return result.type === "model"
        ? result.model.modelName.toUpperCase()
        : result.dataset.name.toUpperCase();
    },

    getFeatureFromResult(result: any) {
      return result.type === "model"
        ? result.model.variables.length ?? 0
        : result.dataset.variables.length ?? 0;
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
  },
});
</script>

<style>
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
  height: 100%;
}

.search-content {
  height: 100%;
  overflow: scroll;
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
  bottom: 1em;
  color: var(--gray-500);
  font-size: 0.75rem;
  position: absolute;
  right: 1em;
}
</style>
