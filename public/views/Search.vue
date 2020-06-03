<template>
  <div class="container-fluid d-flex flex-column h-100 search-view">
    <div class="row flex-0-nav"></div>

    <div class="row align-items-center justify-content-center bg-white">
      <div class="col-12 col-md-10">
        <h5 class="header-label">Select a Dataset</h5>
      </div>
    </div>
    <div class="row">
      <file-uploader-status
        class="file-uploader-status col-12"
        :status="uploadStatus"
        :filename="uploadData.filename"
        :datasetID="uploadData.datasetID"
      />
    </div>
    <div class="row flex-2 align-items-center justify-content-center">
      <div class="col-12 col-md-6">
        <div class="d-flex">
          <search-bar class="search-search-bar"></search-bar>
          <file-uploader
            class="file-uploader"
            @uploadstart="onUploadStart"
            @uploadfinish="onUploadFinish"
          ></file-uploader>
        </div>
      </div>
    </div>

    <section class="row flex-10">
      <b-tabs
        align="center"
        class="search-content"
        content-class="search-content-tab"
      >
        <b-tab :title="'Models (' + nbSearchModels + ')'" active>
          <model-search-results
            class="search-search-results"
            :is-pending="isPending"
          >
          </model-search-results>
        </b-tab>

        <b-tab :title="'Datasets (' + nbSearchDatasets + ')'">
          <dataset-search-results
            class="search-search-results"
            :is-pending="isPending"
          >
          </dataset-search-results>
        </b-tab>
      </b-tabs>
    </section>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import FileUploader from "../components/FileUploader";
import FileUploaderStatus from "../components/FileUploaderStatus";
import DatasetPreviewCard from "../components/DatasetPreviewCard";
import SearchBar from "../components/SearchBar";
import DatasetSearchResults from "../components/DatasetSearchResults";
import ModelSearchResults from "../components/ModelSearchResults";
import { Dataset } from "../store/dataset/index";
import { Model } from "../store/model/index";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as modelGetters } from "../store/model/module";
import { SEARCH_ROUTE, JOIN_DATASETS_ROUTE } from "../store/route/index";

export default Vue.extend({
  name: "search-view",

  components: {
    SearchBar,
    DatasetSearchResults,
    ModelSearchResults,
    DatasetPreviewCard,
    FileUploader,
    FileUploaderStatus
  },

  data() {
    return {
      isPending: false,
      uploadData: {},
      uploadStatus: ""
    };
  },

  computed: {
    terms(): string {
      return routeGetters.getRouteTerms(this.$store);
    },

    nbSearchDatasets(): number {
      return datasetGetters.getCountOfFilteredDatasets(this.$store);
    },
    nbSearchModels(): number {
      return modelGetters.getCountOfModels(this.$store);
    }
  },

  beforeMount() {
    this.fetch();
  },

  watch: {
    terms() {
      this.fetch();
    }
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
    onUploadFinish(err) {
      this.uploadStatus = err ? "error" : "success";
    }
  }
});
</script>

<style>
.header-label {
  padding: 1rem 0 0.5rem 0;
  font-weight: bold;
}

.search-view .file-uploader {
  flex-shrink: 0;
  margin-left: 20px;
}

.row .file-uploader-status {
  padding: 0;
}

.search-search-bar {
  width: 100%;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
}

.search-container {
  height: 100%;
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

.search-content {
  height: 100%;
  width: 100%;
}

.search-content .nav-link {
  padding-left: 1rem;
  padding-right: 1rem;
}

.search-content-tab {
  align-self: center;
  height: calc(100% - 30px); /* 30px ≈ height of the tab nav. */
  overflow: scroll;
  margin-left: auto;
  margin-right: auto;
  max-width: 83.33%; /* Bootstrap ≈ .col-10 */
}
</style>
