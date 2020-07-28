import _ from "lodash";
import axios, { AxiosResponse } from "axios";
import { Dictionary } from "../../util/dict";
import { ActionContext } from "vuex";
import {
  Dataset,
  DatasetOrigin,
  DatasetState,
  Variable,
  Grouping,
  DatasetPendingRequestType,
  DatasetPendingRequestStatus,
  VariableRankingPendingRequest,
  GeocodingPendingRequest,
  JoinSuggestionPendingRequest,
  JoinDatasetImportPendingRequest,
  Task,
  ClusteringPendingRequest,
  SummaryMode,
  BandCombinations,
  BandID,
  isClusteredGrouping,
  TimeSeriesValue
} from "./index";
import { mutations, getters } from "./module";
import store, { DistilState } from "../store";
import { Highlight } from "../dataset/index";
import { FilterParams } from "../../util/filters";
import {
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars,
  validateArgs
} from "../../util/data";
import { addHighlightToFilterParams } from "../../util/highlights";
import { loadImage } from "../../util/image";
import {
  getVarType,
  IMAGE_TYPE,
  GEOCODED_LON_PREFIX,
  GEOCODED_LAT_PREFIX,
  isRankableVariableType,
  REMOTE_SENSING_TYPE,
  isImageType,
  UNKNOWN_TYPE
} from "../../util/types";
import { getters as routeGetters } from "../route/module";

// fetches variables and add dataset name to each variable
async function getVariables(dataset: string): Promise<Variable[]> {
  const response = await axios.get(`/distil/variables/${dataset}`);
  // extend variable with datasetName and isColTypeReviewed property to track type reviewed state in the client state
  return response.data.variables.map(variable => ({
    ...variable,
    datasetName: dataset,
    isColTypeReviewed: false
  }));
}

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {
  // fetches a dataset description.
  async fetchDataset(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<void> {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    try {
      const response = await axios.get(`/distil/datasets/${args.dataset}`);
      mutations.setDataset(context, response.data.dataset);
    } catch (error) {
      console.error(error);
      mutations.setDatasets(context, []);
    }
  },

  // searches dataset descriptions and column names for supplied terms
  async searchDatasets(context: DatasetContext, terms: string): Promise<void> {
    const params = !_.isEmpty(terms) ? `?search=${terms}` : "";
    try {
      const response = await axios.get(`/distil/datasets${params}`);
      mutations.setDatasets(context, response.data.datasets);
    } catch (error) {
      console.error(error);
      mutations.setDatasets(context, []);
    }
  },

  // fetches all variables for a single dataset.
  async fetchVariables(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<void> {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    try {
      const variables = await getVariables(args.dataset);
      mutations.setVariables(context, variables);
    } catch (error) {
      console.error(error);
      mutations.setVariables(context, []);
    }
  },

  // fetches all variables for a two datasets.
  async fetchJoinDatasetsVariables(
    context: DatasetContext,
    args: { datasets: string[] }
  ): Promise<void> {
    if (!validateArgs(args, ["datasets"])) {
      return null;
    }
    try {
      const res = await Promise.all([
        getVariables(args.datasets[0]),
        getVariables(args.datasets[1])
      ]);
      const varsA = res[0];
      const varsB = res[1];
      mutations.setVariables(context, varsA.concat(varsB));
    } catch (error) {
      console.error(error);
      mutations.setVariables(context, []);
    }
  },

  async geocodeVariable(
    context: DatasetContext,
    args: { dataset: string; field: string }
  ): Promise<any> {
    return null;
    /* TODO
     * Disabled because the current solution is not responsive enough:
     * https://github.com/uncharted-distil/distil/issues/1815
    if (!validateArgs(args, ["dataset", "field"])) {
      return null;
    }
    const update: GeocodingPendingRequest = {
      id: _.uniqueId(),
      dataset: args.dataset,
      type: DatasetPendingRequestType.GEOCODING,
      field: args.field,
      status: DatasetPendingRequestStatus.PENDING
    };
    mutations.updatePendingRequests(context, update);
    try {
      await axios.post(`/distil/geocode/${args.dataset}/${args.field}`, {});
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.RESOLVED
      });
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.ERROR
      });
      console.error(error);
    }
    */
  },

  fetchGeocodingResults(
    context: DatasetContext,
    args: { dataset: string; field: string }
  ) {
    // pull the updated dataset, vars, and summaries
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
    const highlight = context.getters.getDecodedHighlight;

    return Promise.all([
      actions.fetchDataset(context, {
        dataset: args.dataset
      }),
      actions.fetchVariables(context, {
        dataset: args.dataset
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LON_PREFIX + args.field,
        highlight: highlight,
        filterParams: filterParams,
        include: true,
        mode: SummaryMode.Default
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LON_PREFIX + args.field,
        highlight: highlight,
        filterParams: filterParams,
        include: false,
        mode: SummaryMode.Default
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LAT_PREFIX + args.field,
        highlight: highlight,
        filterParams: filterParams,
        include: true,
        mode: SummaryMode.Default
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LAT_PREFIX + args.field,
        highlight: highlight,
        filterParams: filterParams,
        include: false,
        mode: SummaryMode.Default
      })
    ]);
  },

  async fetchJoinSuggestions(
    context: DatasetContext,
    args: { dataset: string; searchQuery: string }
  ) {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    const request: JoinSuggestionPendingRequest = {
      id: _.uniqueId(),
      dataset: args.dataset,
      type: DatasetPendingRequestType.JOIN_SUGGESTION,
      status: DatasetPendingRequestStatus.PENDING,
      suggestions: []
    };
    mutations.updatePendingRequests(context, request);

    const query = args.searchQuery
      ? `?search=${args.searchQuery.split(" ").join(",")}`
      : "";
    try {
      const response = await axios.get(
        `/distil/join-suggestions/${args.dataset + query}`
      );
      const suggestions = (response.data && response.data.datasets) || [];
      mutations.updatePendingRequests(context, {
        ...request,
        status: DatasetPendingRequestStatus.RESOLVED,
        suggestions
      });
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...request,
        status: DatasetPendingRequestStatus.ERROR
      });
      console.error(error);
    }
  },

  // Sends a request to the server to generate cluaster for all data that is a valid target for clustering.
  async fetchClusters(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    const update: ClusteringPendingRequest = {
      id: _.uniqueId(),
      dataset: args.dataset,
      type: DatasetPendingRequestType.CLUSTERING,
      status: DatasetPendingRequestStatus.PENDING
    };

    // Find variables that require cluster requests.  If there are none, then
    // quick exit.
    const clusterVariables = getters
      .getVariables(context)
      .filter(
        v =>
          (v.grouping && isClusteredGrouping(v.grouping)) ||
          isImageType(v.colType)
      );
    if (clusterVariables.length === 0) {
      return Promise.resolve();
    }

    mutations.updatePendingRequests(context, update);

    // Find grouped fields that have clusters defined against them and request that they
    // cluster.
    const promises = clusterVariables.map(v => {
      if (v.grouping && isClusteredGrouping(v.grouping)) {
        return axios.post(
          `/distil/cluster/${args.dataset}/${v.grouping.idCol}`,
          {}
        );
      } else if (isImageType(v.colType)) {
        return axios.post(`/distil/cluster/${args.dataset}/${v.colName}`, {});
      }
      return null;
    });
    Promise.all(promises)
      .then(() => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.RESOLVED
        });
      })
      .catch(error => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.ERROR
        });
        console.error(error);
      });
  },

  async uploadDataFile(
    context: DatasetContext,
    args: {
      datasetID: string;
      file: File;
    }
  ): Promise<void> {
    if (!validateArgs(args, ["datasetID", "file"])) {
      return null;
    }
    const data = new FormData();
    data.append("file", args.file);
    let options = "";
    switch (args.file.type) {
      case "text/csv":
        options = "type=table";
        break;
      case "application/x-zip-compressed":
      case "application/zip":
        options = "type=media&image=jpg";
        break;
      default:
        options = "type=table";
    }
    const uploadResponse = await axios.post(
      `/distil/upload/${args.datasetID}?${options}`,
      data,
      {
        headers: { "Content-Type": "multipart/form-data" }
      }
    );
    return actions.importDataset(context, {
      datasetID: args.datasetID,
      source: "augmented",
      provenance: "local",
      terms: args.datasetID,
      originalDataset: null,
      joinedDataset: null,
      path: uploadResponse.data.location
    });
  },

  async importDataset(
    context: DatasetContext,
    args: {
      datasetID: string;
      source: string;
      provenance: string;
      terms: string;
      originalDataset: Dataset;
      joinedDataset: Dataset;
      path: string;
    }
  ): Promise<void> {
    if (!validateArgs(args, ["datasetID", "source"])) {
      return null;
    }

    let postParams = {};
    if (args.originalDataset !== null) {
      postParams = {
        originalDataset: args.originalDataset,
        joinedDataset: args.joinedDataset
      };
    } else if (args.path !== "") {
      postParams = {
        path: args.path
      };
    }
    await axios.post(
      `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
      postParams
    );
    return actions.searchDatasets(context, args.terms);
  },

  async importJoinDataset(
    context: DatasetContext,
    args: {
      datasetID: string;
      source: string;
      provenance: string;
      searchResults: DatasetOrigin[];
    }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }

    const id = _.uniqueId();
    const update: JoinDatasetImportPendingRequest = {
      id,
      dataset: args.datasetID,
      type: DatasetPendingRequestType.JOIN_DATASET_IMPORT,
      status: DatasetPendingRequestStatus.PENDING
    };
    mutations.updatePendingRequests(context, update);
    try {
      const response = await axios.post(
        `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
        {
          searchResults: args.searchResults
        }
      );
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.RESOLVED
      });
      return response && response.data;
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.ERROR
      });
      console.error(error);
    }
  },

  async deleteVariable(
    context: DatasetContext,
    args: { dataset: string; key: string }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset", "key"])) {
      return null;
    }
    try {
      await axios.post(`/distil/delete/${args.dataset}/${args.key}`, {});
      await Promise.all([
        actions.fetchDataset(context, {
          dataset: args.dataset
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset
        })
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlight = context.getters.getDecodedHighlight as Highlight;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        })
      ]);
    } catch (error) {
      console.error(error);
    }
  },

  async joinDatasetsPreview(
    context: DatasetContext,
    args: {
      datasetA: Dataset;
      datasetB: Dataset;
      joinAccuracy: number;
      joinSuggestionIndex: number;
    }
  ): Promise<void> {
    if (!validateArgs(args, ["datasetA", "datasetB", "joinAccuracy"])) {
      return null;
    }

    const datasetBrevised: Dataset = JSON.parse(JSON.stringify(args.datasetB));

    datasetBrevised.variables = datasetBrevised.variables.map(v => {
      const roledVar = v;
      roledVar.role = ["attribute"];
      return roledVar;
    });

    const response = await axios.post(`/distil/join`, {
      accuracy: args.joinAccuracy,
      datasetLeft: args.datasetA,
      datasetRight: datasetBrevised,
      searchResultIndex: args.joinSuggestionIndex
    });
    return response.data;
  },

  async setGrouping(
    context: DatasetContext,
    args: { dataset: string; grouping: Grouping }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset", "grouping"])) {
      return null;
    }
    try {
      await axios.post(`/distil/grouping/${args.dataset}`, {
        grouping: args.grouping
      });
      await Promise.all([
        actions.fetchDataset(context, {
          dataset: args.dataset
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset
        })
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlight = context.getters.getDecodedHighlight as Highlight;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        })
      ]);
    } catch (error) {
      console.error(error);
    }
  },

  async removeGrouping(
    context: DatasetContext,
    args: { dataset: string; variable: string }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset", "variable"])) {
      return null;
    }
    try {
      await axios.post(
        `/distil/remove-grouping/${args.dataset}/${args.variable}`,
        {}
      );
      await Promise.all([
        actions.fetchDataset(context, {
          dataset: args.dataset
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset
        })
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlight = context.getters.getDecodedHighlight as Highlight;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight,
          varModes: varModes
        })
      ]);
    } catch (error) {
      console.error(error);
    }
  },

  async setVariableType(
    context: DatasetContext,
    args: { dataset: string; field: string; type: string }
  ): Promise<any> {
    if (!validateArgs(args, ["dataset", "field", "type"])) {
      return null;
    }

    try {
      await axios.post(`/distil/variables/${args.dataset}`, {
        field: args.field,
        type: args.type
      });
      mutations.updateVariableType(context, args);
      // update variable summary
      const filterParams =
        context.getters.getDecodedSolutionRequestFilterParams;
      const highlight = context.getters.getDecodedHighlight;
      return Promise.all([
        actions.fetchVariableSummary(context, {
          dataset: args.dataset,
          variable: args.field,
          filterParams: filterParams,
          highlight: highlight,
          include: true,
          mode: SummaryMode.Default
        }),
        actions.fetchVariableSummary(context, {
          dataset: args.dataset,
          variable: args.field,
          filterParams: filterParams,
          highlight: highlight,
          include: false,
          mode: SummaryMode.Default
        })
      ]);
    } catch (error) {
      // const key = args.field;
      // const label = args.field;
      // const dataset = args.dataset;
      // mutations.updateIncludedVariableSummaries(
      //   context,
      //   createErrorSummary(key, label, dataset, error)
      // );
      // mutations.updateExcludedVariableSummaries(
      //   context,
      //   createErrorSummary(key, label, dataset, error)
      // );
      mutations.updateVariableType(context, { ...args, type: UNKNOWN_TYPE });
    }
  },

  reviewVariableType(
    context: DatasetContext,
    args: { dataset: string; field: string; isColTypeReviewed: boolean }
  ) {
    mutations.reviewVariableType(context, args);
  },

  fetchIncludedVariableSummaries(
    context: DatasetContext,
    args: {
      dataset: string;
      variables: Variable[];
      highlight: Highlight;
      filterParams: FilterParams;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    return actions.fetchVariableSummaries(context, {
      dataset: args.dataset,
      variables: args.variables,
      filterParams: args.filterParams,
      highlight: args.highlight,
      include: true,
      varModes: args.varModes
    });
  },

  fetchExcludedVariableSummaries(
    context: DatasetContext,
    args: {
      dataset: string;
      variables: Variable[];
      highlight: Highlight;
      filterParams: FilterParams;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    return actions.fetchVariableSummaries(context, {
      dataset: args.dataset,
      variables: args.variables,
      filterParams: args.filterParams,
      highlight: args.highlight,
      include: false,
      varModes: args.varModes
    });
  },

  fetchVariableSummaries(
    context: DatasetContext,
    args: {
      dataset: string;
      variables: Variable[];
      highlight: Highlight;
      filterParams: FilterParams;
      include: boolean;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    if (!validateArgs(args, ["dataset", "variables"])) {
      return null;
    }

    const mutator = args.include
      ? mutations.updateIncludedVariableSummaries
      : mutations.updateExcludedVariableSummaries;
    const existingSummaries = args.include
      ? context.state.includedSet.variableSummaries
      : context.state.excludedSet.variableSummaries;

    // commit empty place holders, if there is no data
    const promises = [];
    args.variables.forEach(variable => {
      const exists = _.find(existingSummaries, v => {
        return v.dataset === args.dataset && v.key === variable.colName;
      });

      if (!exists) {
        // add placeholder if it doesn't exist
        const key = variable.colName;
        const label = variable.colDisplayName;
        const dataset = args.dataset;
        const desciption = variable.colDescription;
        mutator(context, createPendingSummary(key, label, desciption, dataset));
      }

      // Get the mode or default
      const mode = args.varModes.has(variable.colName)
        ? args.varModes.get(variable.colName)
        : SummaryMode.Default;

      // fetch summary
      promises.push(
        actions.fetchVariableSummary(context, {
          dataset: args.dataset,
          variable: variable.colName,
          filterParams: args.filterParams,
          highlight: args.highlight,
          include: args.include,
          mode: mode
        })
      );
    });
    // fill them in asynchronously
    return Promise.all(promises);
  },

  async fetchVariableSummary(
    context: DatasetContext,
    args: {
      dataset: string;
      variable: string;
      highlight?: Highlight;
      filterParams: FilterParams;
      include: boolean;
      mode: SummaryMode;
    }
  ): Promise<void> {
    if (!validateArgs(args, ["dataset", "variable"])) {
      return null;
    }

    const filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlight
    );
    const mutator = args.include
      ? mutations.updateIncludedVariableSummaries
      : mutations.updateExcludedVariableSummaries;

    try {
      const response = await axios.post(
        `/distil/variable-summary/${args.dataset}/${
          args.variable
        }/${!args.include}/${args.mode}`,
        filterParams
      );
      const summary = response.data.summary;
      await fetchSummaryExemplars(args.dataset, args.variable, summary);
      mutator(context, summary);
    } catch (error) {
      console.error(error);
      const key = args.variable;
      const label = args.variable;
      const dataset = args.dataset;
      mutator(context, createErrorSummary(key, label, dataset, error));
    }
  },

  async fetchVariableRankings(
    context: DatasetContext,
    args: { dataset: string; target: string }
  ) {
    const id = _.uniqueId();
    const update: VariableRankingPendingRequest = {
      id,
      dataset: args.dataset,
      type: DatasetPendingRequestType.VARIABLE_RANKING,
      status: DatasetPendingRequestStatus.PENDING,
      rankings: null,
      target: args.target
    };

    // quick exit if we don't have variables/target that are going to yield ranking
    const target = getters.getVariablesMap(context)[args.target];
    const rankableVariables = getters
      .getVariables(context)
      .filter(
        f => f.colName !== target.colName && isRankableVariableType(f.colType)
      );
    if (
      !isRankableVariableType(target.colType) ||
      rankableVariables.length === 0
    ) {
      return Promise.resolve();
    }

    mutations.updatePendingRequests(context, update);
    try {
      const dataset = args.dataset;

      const response = await axios.get(
        `/distil/variable-rankings/${dataset}/${args.target}`
      );

      const rankings = <Dictionary<number>>response.data;

      // check to see if we got any non-zero rank info back
      const computedRankings = _.filter(rankings, (r, v) => r !== 0).length > 0;

      // check to see if the returned ranks are different than any that we may have previously computed
      const oldRankings = getters.getVariableRankings(context)[args.dataset];

      // If we have valid rankings and they are different than those previously computed we mark
      // as resolved so the user can apply them.  Otherwise we mark as reviewed, so that there is
      // no flag for the user to apply.
      let status = DatasetPendingRequestStatus.REVIEWED;
      if (computedRankings && !_.isEqual(oldRankings, rankings)) {
        // If the request has already been reviewed, we apply the rankings.
        if (routeGetters.getRouteIsTrainingVariablesRanked(store)) {
          mutations.setVariableRankings(context, { dataset, rankings });
          mutations.updateVariableRankings(context, rankings);
        } else {
          status = DatasetPendingRequestStatus.RESOLVED;
        }
      }

      // Update the status.
      mutations.updatePendingRequests(context, { ...update, status, rankings });
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.ERROR
      });
      console.error(error);
    }
  },

  updateVariableRankings(
    context: DatasetContext,
    args: { dataset: string; rankings: Dictionary<number> }
  ) {
    mutations.setVariableRankings(context, {
      dataset: args.dataset,
      rankings: args.rankings
    });
    mutations.updateVariableRankings(context, args.rankings);
  },

  updatePendingRequestStatus(
    context: DatasetContext,
    args: { id: string; status: DatasetPendingRequestStatus }
  ) {
    const update = context.getters.getPendingRequests.find(
      item => item.id === args.id
    );
    if (update) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: args.status
      });
    }
  },

  removePendingRequest(context: DatasetContext, id: string) {
    mutations.removePendingRequest(context, id);
  },

  // update filtered data based on the current filter state
  fetchFiles(
    context: DatasetContext,
    args: { dataset: string; variable: string; urls: string[] }
  ) {
    if (!validateArgs(args, ["dataset", "variable", "urls"])) {
      return null;
    }
    const type = getVarType(args.variable);
    return Promise.all(
      args.urls.map(url => {
        if (type === IMAGE_TYPE) {
          return actions.fetchImage(context, {
            dataset: args.dataset,
            url: url
          });
        }
        if (type === REMOTE_SENSING_TYPE) {
          return actions.fetchMultiBandImage(context, {
            dataset: args.dataset,
            imageId: url,
            bandCombination: BandID.NATURAL_COLORS
          });
        }
        if (type === "graph") {
          return actions.fetchGraph(context, {
            dataset: args.dataset,
            url: url
          });
        }
        return actions.fetchFile(context, {
          dataset: args.dataset,
          url: url
        });
      })
    );
  },

  async fetchImage(
    context: DatasetContext,
    args: { dataset: string; url: string }
  ) {
    if (!validateArgs(args, ["dataset", "url"])) {
      return null;
    }
    try {
      const response = await loadImage(
        `distil/image/${args.dataset}/${args.url}`
      );
      mutations.updateFile(context, { url: args.url, file: response });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchTimeseries(
    context: DatasetContext,
    args: {
      dataset: string;
      xColName: string;
      yColName: string;
      timeseriesColName: string;
      timeseriesId: any;
    }
  ) {
    if (
      !validateArgs(args, [
        "dataset",
        "xColName",
        "yColName",
        "timeseriesColName",
        "timeseriesId"
      ])
    ) {
      return null;
    }

    try {
      const response = await axios.post(
        `distil/timeseries/${encodeURIComponent(
          args.dataset
        )}/${encodeURIComponent(args.timeseriesColName)}/${encodeURIComponent(
          args.xColName
        )}/${encodeURIComponent(args.yColName)}/${encodeURIComponent(
          args.timeseriesId
        )}/false`,
        {}
      );
      mutations.updateTimeseries(context, {
        dataset: args.dataset,
        id: args.timeseriesId,
        timeseries: <TimeSeriesValue[]>response.data.timeseries,
        isDateTime: <boolean>response.data.isDateTime
      });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchMultiBandImage(
    context: DatasetContext,
    args: {
      dataset: string;
      imageId: string;
      bandCombination: string;
    }
  ) {
    if (!validateArgs(args, ["dataset", "imageId", "bandCombination"])) {
      return null;
    }

    try {
      const response = await loadImage(
        `distil/multiband-image/${args.dataset}/${args.imageId}/${args.bandCombination}`
      );
      mutations.updateFile(context, { url: args.imageId, file: response });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchGraph(
    context: DatasetContext,
    args: { dataset: string; url: string }
  ) {
    if (!validateArgs(args, ["dataset", "url"])) {
      return null;
    }
    try {
      const response = await axios.get(
        `distil/graphs/${args.dataset}/${args.url}`
      );
      if (response.data.graphs.length > 0) {
        const graph = response.data.graphs[0];
        const parsed = {
          nodes: graph.nodes.map(n => {
            return {
              id: n.id,
              label: n.label,
              x: n.attributes.attr1,
              y: n.attributes.attr2,
              size: 1,
              color: "#ec5148"
            };
          }),
          edges: graph.edges.map((e, i) => {
            return {
              id: `e${i}`,
              source: e.source,
              target: e.target,
              color: "#aaa"
            };
          })
        };
        mutations.updateFile(context, { url: args.url, file: parsed });
      }
    } catch (error) {
      console.error(error);
    }
  },

  async fetchFile(
    context: DatasetContext,
    args: { dataset: string; url: string }
  ) {
    if (!validateArgs(args, ["dataset", "url"])) {
      return null;
    }
    try {
      const response = await axios.get(
        `distil/resource/${args.dataset}/${args.url}`
      );
      mutations.updateFile(context, { url: args.url, file: response.data });
    } catch (error) {
      console.error(error);
    }
  },

  // update filtered data based on the current filter state
  fetchJoinDatasetsTableData(
    context: DatasetContext,
    args: {
      datasets: string[];
      filterParams: Dictionary<FilterParams>;
      highlight: Highlight;
    }
  ) {
    if (!validateArgs(args, ["dataset", "filterParams"])) {
      return null;
    }
    return Promise.all(
      args.datasets.map(async dataset => {
        const highlight =
          (args.highlight && args.highlight.dataset) === dataset
            ? args.highlight
            : null;
        const filterParams = addHighlightToFilterParams(
          args.filterParams[dataset],
          highlight
        );

        try {
          const response = await axios.post(
            `distil/data/${dataset}/false`,
            filterParams
          );
          mutations.setJoinDatasetsTableData(context, {
            dataset: dataset,
            data: response.data
          });
        } catch (error) {
          console.error(error);
          mutations.setJoinDatasetsTableData(context, {
            dataset: dataset,
            data: createEmptyTableData()
          });
        }
      })
    );
  },

  fetchIncludedTableData(
    context: DatasetContext,
    args: { dataset: string; filterParams: FilterParams; highlight: Highlight }
  ) {
    return actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: args.filterParams,
      highlight: args.highlight,
      include: true
    });
  },

  fetchExcludedTableData(
    context: DatasetContext,
    args: { dataset: string; filterParams: FilterParams; highlight: Highlight }
  ) {
    return actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: args.filterParams,
      highlight: args.highlight,
      include: false
    });
  },

  async fetchTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlight: Highlight;
      include: boolean;
    }
  ) {
    if (!validateArgs(args, ["dataset", "filterParams"])) {
      return null;
    }

    const mutator = args.include
      ? mutations.setIncludedTableData
      : mutations.setExcludedTableData;

    const filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlight
    );

    try {
      const response = await axios.post(
        `distil/data/${args.dataset}/${!args.include}`,
        filterParams
      );
      mutator(context, response.data);
    } catch (error) {
      console.error(error);
      mutator(context, createEmptyTableData());
    }
  },

  fetchTask(
    context: DatasetContext,
    args: { dataset: string; targetName: string; variableNames: string[] }
  ): Promise<AxiosResponse<Task>> {
    const varNamesStr =
      args.variableNames.length > 0 ? args.variableNames.join(",") : null;
    return axios.get<Task>(
      `/distil/task/${args.dataset}/${args.targetName}/${varNamesStr}`
    );
  },

  async fetchMultiBandCombinations(
    context: DatasetContext,
    args: { dataset: string }
  ) {
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }

    try {
      const repsonse = await axios.get<BandCombinations>(
        `distil/multiband-combinations/${args.dataset}`
      );
      const bands = repsonse.data.combinations;
      mutations.updateBands(context, bands);
    } catch (error) {
      console.error(error);
    }
  }
};
