<template>
  <div class="card card-result">
    <div
      class="dataset-header hover card-header"
      variant="dark"
      @click.stop="setActiveDataset()"
      v-bind:class="{
        collapsed: !expanded,
        disabled: isImportReady || importPending,
      }"
    >
      <a class="nav-link">
        <i class="fa fa-table"></i> <b>Dateset Name:</b>
        {{ dataset.name }}
      </a>
      <a class="nav-link"
        ><b>Features:</b>
        {{ filterVariablesByFeature(dataset.variables).length }}</a
      >
      <a class="nav-link"><b>Rows:</b> {{ dataset.numRows }}</a>
      <a class="nav-link"><b>Size:</b> {{ formatBytes(dataset.numBytes) }}</a>
      <a v-if="isImportReady">
        <b-button
          class="dataset-preview-button"
          variant="danger"
          @click.stop="importDataset()"
        >
          <div class="row justify-content-center pl-3 pr-3">
            <i class="fa fa-cloud-download mr-2"></i>
            <b>Import</b>
          </div>
        </b-button></a
      >
      <a class="nav-link import-progress-bar" v-if="importPending">
        <b-progress
          :value="percentComplete"
          variant="outline-secondary"
          striped
          :animated="true"
        ></b-progress>
      </a>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-4">
          <span><b>Top features:</b></span>
          <ul>
            <li :key="variable.name" v-for="variable in topVariables">
              {{ variable.colDisplayName }}
            </li>
          </ul>
        </div>
        <div class="col-8">
          <div v-if="dataset.summaryML.length > 0">
            <span><b>May relate to topics such as:</b></span>
            <p class="small-text">
              {{ dataset.summaryML }}
            </p>
          </div>
          <span><b>Summary:</b></span>
          <p class="small-text">
            {{ dataset.summary || "n/a" }}
          </p>
        </div>
      </div>

      <div class="row mt-1">
        <div v-if="!expanded" class="col-12">
          <b-button
            class="full-width hover"
            variant="outline-secondary"
            @click="toggleExpansion()"
          >
            More Details...
          </b-button>
        </div>
        <div v-if="expanded" class="col-12">
          <span><b>Full Description:</b></span>
          <p v-html="highlightedDescription()" />
          <b-button
            class="full-width hover"
            variant="outline-secondary"
            @click="toggleExpansion()"
          >
            Less Details...
          </b-button>
        </div>
      </div>
    </div>
    <error-modal
      :show="showImportFailure"
      title="Import Failed"
      @close="showImportFailure = !showImportFailure"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ErrorModal from "../components/ErrorModal";
import { createRouteEntry } from "../util/routes";
import { formatBytes } from "../util/bytes";
import {
  sortVariablesByPCARanking,
  isDatamartProvenance,
  filterVariablesByFeature,
} from "../util/data";
import { getters as routeGetters } from "../store/route/module";
import { Dataset, Variable } from "../store/dataset/index";
import { actions as datasetActions } from "../store/dataset/module";
import { SELECT_TARGET_ROUTE } from "../store/route/index";
import localStorage from "store";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
  name: "dataset-preview",

  components: {
    ErrorModal,
  },

  props: {
    dataset: Object as () => Dataset,
    allowImport: Boolean as () => boolean,
  },

  data() {
    return {
      expanded: false,
      importPending: false,
      showImportFailure: false,
    };
  },

  computed: {
    terms(): string {
      return routeGetters.getRouteTerms(this.$store);
    },
    isImportReady(): boolean {
      return (
        this.allowImport &&
        !this.importPending &&
        this.datamartProvenance(this.dataset.provenance)
      );
    },
    topVariables(): Variable[] {
      return sortVariablesByPCARanking(
        filterVariablesByFeature(this.dataset.variables).slice(0)
      ).slice(0, NUM_TOP_FEATURES);
    },
    percentComplete(): number {
      return 100;
    },
  },

  methods: {
    formatBytes(n: number): string {
      return formatBytes(n);
    },
    filterVariablesByFeature(variables: Variable[]): Variable[] {
      return filterVariablesByFeature(variables);
    },
    setActiveDataset() {
      if (this.isImportReady || this.importPending) {
        return;
      }
      const entry = createRouteEntry(SELECT_TARGET_ROUTE, {
        dataset: this.dataset.id,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
      this.addRecentDataset(this.dataset.id);
      appActions.logUserEvent(this.$store, {
        feature: Feature.SELECT_DATASET,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_OPEN,
        details: { dataset: this.dataset.id },
      });
    },
    toggleExpansion() {
      this.expanded = !this.expanded;
    },
    highlightedDescription(): string {
      const terms = this.terms;
      if (_.isEmpty(terms)) {
        return this.dataset.description;
      }
      const split = terms.split(/[ ,]+/); // split on whitespace
      const joined = split.join("|"); // join
      const regex = new RegExp(`(${joined})(?![^<]*>)`, "gm");
      return this.dataset.description.replace(
        regex,
        '<span class="highlight">$1</span>'
      );
    },
    addRecentDataset(dataset: string) {
      const datasets = localStorage.get("recent-datasets") || [];
      if (datasets.indexOf(dataset) === -1) {
        datasets.unshift(dataset);
        localStorage.set("recent-datasets", datasets);
      }
    },
    importDataset() {
      this.importPending = true;
      datasetActions
        .importDataset(this.$store, {
          datasetID: this.dataset.id,
          terms: this.terms,
          source: "contrib",
          provenance: this.dataset.provenance,
          originalDataset: null,
          joinedDataset: null,
          path: "",
        })
        .then(() => {
          this.importPending = false;
        })
        .catch(() => {
          this.showImportFailure = true;
          this.importPending = false;
        });
    },
    datamartProvenance(provenance: string): boolean {
      return isDatamartProvenance(provenance);
    },
  },
});
</script>

<style>
.highlight {
  background-color: #87cefa;
}
.dataset-header {
  display: flex;
  padding: 4px 8px;
  color: white;
  justify-content: space-between;
  border: none;
  border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}
.card-result .card-header {
  background-color: #424242;
}
.card-result .card-header:hover {
  color: #fff;
  background-color: #535353;
}
.dataset-preview-button {
  line-height: 14px !important;
}
.dataset-header:hover {
  text-decoration: underline;
}
.full-width {
  width: 100%;
}
.import-progress-bar {
  position: relative;
  width: 128px;
}
.import-progress-bar .progress {
  height: 22px;
}
</style>
