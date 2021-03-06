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
  <div class="container-fluid d-flex flex-column h-100 home-view">
    <div class="row flex-0-nav"></div>
    <div class="row flex-1 align-items-center justify-content-center bg-white">
      <div class="col-12 col-md-10">
        <h5 class="header-label">Recent Activity</h5>
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
      <div class="col-12 col-md-6 justify-content-center">
        <div class="d-flex">
          <search-bar class="home-search-bar"></search-bar>
          <file-uploader
            class="file-uploader"
            @uploadstart="onUploadStart"
            @uploadfinish="onUploadFinish"
          ></file-uploader>
        </div>
      </div>
    </div>
    <div class="row flex-10 justify-content-center pb-3 home-item-container">
      <div class="col-12 col-md-10 d-flex">
        <div class="home-items">
          <recent-datasets :max-datasets="5"></recent-datasets>
          <recent-solutions :max-solutions="5"></recent-solutions>
          <running-solutions :max-solutions="5"></running-solutions>
        </div>
      </div>
    </div>
    <div class="home-version-text text-muted">
      {{ version }}
    </div>
  </div>
</template>

<script lang="ts">
import FileUploader from "../components/FileUploader.vue";
import FileUploaderStatus from "../components/FileUploaderStatus.vue";
import RecentDatasets from "../components/RecentDatasets.vue";
import RecentSolutions from "../components/RecentSolutions.vue";
import RunningSolutions from "../components/RunningSolutions.vue";
import SearchBar from "../components/SearchBar.vue";
import { getters as appGetters } from "../store/app/module";
import { actions as viewActions } from "../store/view/module";
import Vue from "vue";

export default Vue.extend({
  name: "home-view",

  components: {
    FileUploader,
    FileUploaderStatus,
    RecentDatasets,
    RecentSolutions,
    RunningSolutions,
    SearchBar,
  },

  data() {
    return {
      uploadData: {},
      uploadStatus: "",
    };
  },

  computed: {
    version(): string {
      return `version: ${appGetters.getVersionNumber(
        this.$store
      )} at ${appGetters.getVersionTimestamp(this.$store)}`;
    },
  },

  beforeMount() {
    viewActions.fetchHomeData(this.$store);
  },

  methods: {
    onUploadStart(uploadData) {
      this.uploadData = uploadData;
      this.uploadStatus = "started";
    },

    onUploadFinish(err) {
      this.uploadStatus = err ? "error" : "success";
    },
  },
});
</script>

<style>
.header-label {
  padding: 1rem 0 0.5rem 0;
  font-weight: bold;
}

.home-search-bar {
  width: 100%;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
}

.home-items {
  width: 100%;
  overflow: auto;
}

.home-items .card {
  margin-bottom: 1rem;
}

.home-version-text {
  margin: 0 auto;
  font-size: 0.8rem;
}

.home-view .file-uploader {
  flex-shrink: 0;
  margin-left: 20px;
}

.home-view .file-uploader-status {
  padding: 0;
}

.home-item-container {
  overflow: scroll;
}
</style>
