/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

import axios, { AxiosResponse } from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import {
  createEmptyTableData,
  createErrorSummary,
  createPendingSummary,
  DatasetUpdate,
  fetchSummaryExemplars,
  minimumRouteKey,
  validateArgs,
} from "../../util/data";
import { Dictionary } from "../../util/dict";
import { EXCLUDE_FILTER, FilterParams } from "../../util/filters";
import {
  addHighlightToFilterParams,
  cloneFilters,
  setInvert,
} from "../../util/highlights";
import { loadImage } from "../../util/image";
import {
  GEOCODED_LAT_PREFIX,
  GEOCODED_LON_PREFIX,
  getVarType,
  IMAGE_TYPE,
  isImageType,
  isMultibandImageType,
  isRankableVariableType,
  MULTIBAND_IMAGE_TYPE,
  UNKNOWN_TYPE,
  MultiBandImagePackRequest,
} from "../../util/types";
import { getters as routeGetters } from "../route/module";
import store, { DistilState } from "../store";
import {
  BandCombinations,
  BandID,
  ClonedInfo,
  ClusteringPendingRequest,
  DataMode,
  Dataset,
  DatasetOrigin,
  DatasetPendingRequestStatus,
  DatasetPendingRequestType,
  DatasetState,
  Grouping,
  Highlight,
  isClusteredGrouping,
  JoinDatasetImportPendingRequest,
  JoinSuggestionPendingRequest,
  Metrics,
  OutlierPendingRequest,
  SummaryMode,
  TableData,
  Task,
  Variable,
  VariableRankingPendingRequest,
} from "./index";
import { getters, mutations } from "./module";
import { TimeSeriesUpdate } from "./mutations";

// fetches variables and add dataset name to each variable
async function getVariables(dataset: string): Promise<Variable[]> {
  const response = await axios.get(`/distil/variables/${dataset}`);
  // extend variable with datasetName and isColTypeReviewed property to track type reviewed state in the client state
  return response.data.variables.map((variable) => ({
    ...variable,
    datasetName: dataset,
    isColTypeReviewed: false,
  }));
}

// Return the best variable name of a dataset for outlier detection
function getOutlierVariableName(): string {
  const variables = getters.getVariables(store) ?? [];
  const target = routeGetters.getTargetVariable(store) ?? ({} as Variable);

  /*
    Find a grouping variable, specially a remote-sensing one.
    This is needed in case the remote-sensing images have not
    been prefiturized.
  */
  const groupingVariables = variables.filter((v) => v.grouping);
  const remoteSensingVariable = groupingVariables.find((gv) =>
    isMultibandImageType(gv.colType)
  );

  /*
    The variable name to be sent, is, in order of availability:
      - a remote-sensing variable first,
      - a grouping variable second,
      - or the target variable
  */
  return (
    remoteSensingVariable?.grouping.idCol ??
    groupingVariables[0]?.grouping.idCol ??
    target.key
  );
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

  async deleteDataset(
    context: DatasetContext,
    payload: { dataset: string; terms: string }
  ): Promise<void> {
    if (!payload.dataset) {
      return;
    }
    try {
      // delete dataset
      const response = await axios.post(
        `/distil/delete-dataset/${payload.dataset}`
      );
      // update current list of datasets
      await actions.searchDatasets(context, payload.terms);
    } catch (err) {
      console.error(err);
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
        getVariables(args.datasets[1]),
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
    const highlights = context.getters.getDecodedHighlights;

    return Promise.all([
      actions.fetchDataset(context, {
        dataset: args.dataset,
      }),
      actions.fetchVariables(context, {
        dataset: args.dataset,
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LON_PREFIX + args.field,
        highlights: highlights,
        filterParams: filterParams,
        include: true,
        dataMode: DataMode.Default,
        mode: SummaryMode.Default,
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LON_PREFIX + args.field,
        highlights: highlights,
        filterParams: filterParams,
        include: false,
        dataMode: DataMode.Default,
        mode: SummaryMode.Default,
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LAT_PREFIX + args.field,
        highlights: highlights,
        filterParams: filterParams,
        include: true,
        dataMode: DataMode.Default,
        mode: SummaryMode.Default,
      }),
      actions.fetchVariableSummary(context, {
        dataset: args.dataset,
        variable: GEOCODED_LAT_PREFIX + args.field,
        highlights: highlights,
        filterParams: filterParams,
        include: false,
        dataMode: DataMode.Default,
        mode: SummaryMode.Default,
      }),
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
      suggestions: [],
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
        suggestions,
      });
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...request,
        status: DatasetPendingRequestStatus.ERROR,
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
      status: DatasetPendingRequestStatus.PENDING,
    };

    // Find variables that require cluster requests;
    // If there are none, then quick exit.
    const clusterVariables = getters
      .getVariables(context)
      .filter(
        (v) =>
          (v.grouping && isClusteredGrouping(v.grouping)) ||
          isImageType(v.colType)
      );
    if (clusterVariables.length === 0) {
      return Promise.resolve();
    }

    mutations.updatePendingRequests(context, update);

    // Find grouped fields that have clusters defined against them and request that they
    // cluster.
    const promises = clusterVariables.map((v) => {
      if (v.grouping && isClusteredGrouping(v.grouping)) {
        return axios.post(
          `/distil/cluster/${args.dataset}/${v.grouping.idCol}`,
          {}
        );
      } else if (isImageType(v.colType)) {
        return axios.post(`/distil/cluster/${args.dataset}/${v.key}`, {});
      }
      return null;
    });
    Promise.all(promises)
      .then(() => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.RESOLVED,
        });
      })
      .catch((error) => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.ERROR,
        });
        console.error(error);
      });
  },

  async fetchOutliers(context: DatasetContext, args: { dataset: string }) {
    // Check if the outlier detection has already been applied.
    if (routeGetters.isOutlierApplied(store)) return;

    const { dataset } = args;
    const variableName = getOutlierVariableName();

    // Create the request.
    let status;
    const request: OutlierPendingRequest = {
      id: _.uniqueId(),
      dataset,
      type: DatasetPendingRequestType.OUTLIER,
      status,
    };

    // Set the request status as pending.
    status = DatasetPendingRequestStatus.PENDING;
    mutations.updatePendingRequests(context, { ...request, status });

    // Run the outlier detection
    try {
      await axios.get(`/distil/outlier-detection/${dataset}/${variableName}`);
      status = DatasetPendingRequestStatus.RESOLVED;
    } catch (error) {
      console.error(error);
      status = DatasetPendingRequestStatus.ERROR;
    }

    // Update the pending request status
    mutations.updatePendingRequests(context, { ...request, status });
  },

  async applyOutliers(context: DatasetContext, dataset: string) {
    const variableName = getOutlierVariableName();

    try {
      await axios.get(`/distil/outlier-results/${dataset}/${variableName}`);
    } catch (error) {
      console.error(error);
    }
  },

  async uploadDataFile(
    context: DatasetContext,
    args: {
      datasetID: string;
      file: File;
    }
  ): Promise<any> {
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
        headers: { "Content-Type": "multipart/form-data" },
      }
    );
    return uploadResponse;
  },

  async importDataFile(
    context: DatasetContext,
    args: {
      datasetID: string;
      file: File;
    }
  ): Promise<any> {
    if (!validateArgs(args, ["datasetID", "file"])) {
      return null;
    }
    const uploadResponse = await actions.uploadDataFile(context, {
      datasetID: args.datasetID,
      file: args.file,
    });
    const response = await actions.importDataset(context, {
      datasetID: args.datasetID,
      source: "augmented",
      provenance: "local",
      terms: args.datasetID,
      originalDataset: null,
      joinedDataset: null,
      path: uploadResponse.data.location,
    });

    // Add the location for potential reimport of the dataset.
    response.location = uploadResponse.data.location;
    return response;
  },

  async updateDataset(
    context: DatasetContext,
    args: { dataset: string; updateData: DatasetUpdate[] }
  ) {
    try {
      await axios.post(`/distil/update/${args.dataset}`, {
        updates: args.updateData,
      });
    } catch (error) {
      console.error(error);
    }
  },
  // Re import a dataset without sampling
  async importFullDataset(
    context: DatasetContext,
    args: { datasetID: string; path: string }
  ) {
    return actions.importDataset(context, {
      datasetID: args.datasetID,
      source: "augmented",
      provenance: "local",
      terms: args.datasetID,
      originalDataset: null,
      joinedDataset: null,
      path: args.path,
      nosample: true,
    });
  },

  // Import a Dataset that is available in $D3MOUTPUTDIR/PUBLIC_SUBFOLDER folder
  async importAvailableDataset(
    context: DatasetContext,
    args: { datasetID: string; path: string }
  ) {
    const response = await actions.importDataset(context, {
      datasetID: args.datasetID,
      source: "public",
      provenance: "local",
      terms: args.datasetID,
      originalDataset: null,
      joinedDataset: null,
      path: args.path,
    });

    // Add the location for potential reimport of the dataset.
    response.location = args.path;
    return response;
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
      nosample?: boolean;
    }
  ): Promise<any> {
    if (!validateArgs(args, ["datasetID", "source"])) {
      return null;
    }

    let postParams = {};
    if (args.path !== "") {
      postParams = {
        path: args.path,
        nosample: args.nosample,
        originalDataset: args.originalDataset,
        joinedDataset: args.joinedDataset,
      };
    } else if (args.originalDataset !== null) {
      postParams = {
        originalDataset: args.originalDataset,
        joinedDataset: args.joinedDataset,
      };
    }
    const response = await axios.post(
      `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
      postParams
    );
    await actions.searchDatasets(context, args.terms);
    return response.data;
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
      status: DatasetPendingRequestStatus.PENDING,
    };
    mutations.updatePendingRequests(context, update);
    try {
      const response = await axios.post(
        `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
        {
          searchResults: args.searchResults,
        }
      );
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.RESOLVED,
      });
      return response && response.data;
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.ERROR,
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
          dataset: args.dataset,
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset,
        }),
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlights = context.getters.getDecodedHighlights as Highlight[];
      const dataMode = context.getters.getDataMode as DataMode;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
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
      joinSuggestionIndex?: number;
      datasetAColumn?: string;
      datasetBColumn?: string;
    }
  ): Promise<any> {
    if (!validateArgs(args, ["datasetA", "datasetB", "joinAccuracy"])) {
      return null;
    }

    const datasetBrevised: Dataset = JSON.parse(JSON.stringify(args.datasetB));

    datasetBrevised.variables = datasetBrevised.variables.map((v) => {
      const roledVar = v;
      roledVar.role = ["attribute"];
      return roledVar;
    });

    const response = await axios.post(`/distil/join`, {
      accuracy: args.joinAccuracy,
      datasetLeft: args.datasetA,
      datasetRight: datasetBrevised,
      datasetAColumn: args.datasetAColumn,
      datasetBColumn: args.datasetBColumn,
      searchResultIndex: args.joinSuggestionIndex,
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
        grouping: args.grouping,
      });
      await Promise.all([
        actions.fetchDataset(context, {
          dataset: args.dataset,
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset,
        }),
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlights = context.getters.getDecodedHighlights as Highlight[];
      const dataMode = context.getters.getDataMode as DataMode;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
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
          dataset: args.dataset,
        }),
        actions.fetchVariables(context, {
          dataset: args.dataset,
        }),
      ]);
      mutations.clearVariableSummaries(context);
      const variables = context.getters.getVariables as Variable[];
      const filterParams = context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams;
      const highlights = context.getters.getDecodedHighlights as Highlight[];
      const dataMode = context.getters.getDataMode as DataMode;
      const varModes = context.getters.getDecodedVarModes as Map<
        string,
        SummaryMode
      >;
      return Promise.all([
        actions.fetchIncludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
        actions.fetchExcludedVariableSummaries(context, {
          dataset: args.dataset,
          variables: variables,
          filterParams: filterParams,
          highlights: highlights,
          dataMode: dataMode,
          varModes: varModes,
        }),
      ]);
    } catch (error) {
      console.error(error);
    }
  },

  async updateGrouping(
    context: DatasetContext,
    args: { variable: string; grouping: Grouping }
  ): Promise<any> {
    if (!validateArgs(args, ["variable", "grouping"])) {
      return null;
    }

    const variable = args.variable;
    const grouping = args.grouping;
    const dataset = grouping.dataset;

    try {
      // Remove the existing grouping
      await axios.post(`/distil/remove-grouping/${dataset}/${variable}`, {});

      // Recreate it with an extra idCol
      await actions.setGrouping(context, { dataset, grouping });
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
      const response = await axios.post(`/distil/variables/${args.dataset}`, {
        field: args.field,
        type: args.type,
      });
      const updatedArgs = {
        ...args,
        variables: response.data.variables as Variable[],
      };
      mutations.updateVariableType(context, updatedArgs);
      // update variable summary
      const filterParams =
        context.getters.getDecodedSolutionRequestFilterParams;
      const highlights = context.getters.getDecodedHighlights;
      return Promise.all([
        actions.fetchVariableSummary(context, {
          dataset: args.dataset,
          variable: args.field,
          filterParams: filterParams,
          highlights: highlights,
          include: true,
          dataMode: DataMode.Default,
          mode: SummaryMode.Default,
        }),
        actions.fetchVariableSummary(context, {
          dataset: args.dataset,
          variable: args.field,
          filterParams: filterParams,
          highlights: highlights,
          include: false,
          dataMode: DataMode.Default,
          mode: SummaryMode.Default,
        }),
      ]);
    } catch (error) {
      mutations.updateVariableType(context, {
        ...args,
        type: UNKNOWN_TYPE,
        variables: context.state.variables,
      });
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
      highlights: Highlight[];
      filterParams: FilterParams;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    return actions.fetchVariableSummaries(context, {
      dataset: args.dataset,
      variables: args.variables,
      filterParams: args.filterParams,
      highlights: args.highlights,
      include: true,
      dataMode: args.dataMode,
      varModes: args.varModes,
    });
  },

  fetchExcludedVariableSummaries(
    context: DatasetContext,
    args: {
      dataset: string;
      variables: Variable[];
      highlights: Highlight[];
      filterParams: FilterParams;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    return actions.fetchVariableSummaries(context, {
      dataset: args.dataset,
      variables: args.variables,
      filterParams: args.filterParams,
      highlights: args.highlights,
      include: false,
      dataMode: args.dataMode,
      varModes: args.varModes,
    });
  },

  fetchVariableSummaries(
    context: DatasetContext,
    args: {
      dataset: string;
      variables: Variable[];
      highlights: Highlight[];
      filterParams: FilterParams;
      include: boolean;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ): Promise<void[]> {
    if (!validateArgs(args, ["dataset", "variables"])) {
      return null;
    }
    const mutator = args.include
      ? mutations.updateIncludedVariableSummaries
      : mutations.updateExcludedVariableSummaries;

    const summariesByVariable = args.include
      ? context.state.includedSet.variableSummariesByKey
      : context.state.excludedSet.variableSummariesByKey;
    const routeKey = minimumRouteKey();
    const promises = [];

    args.variables.forEach((variable) => {
      const existingVariableSummary =
        summariesByVariable?.[variable.key]?.[routeKey];

      if (existingVariableSummary) {
        promises.push(existingVariableSummary);
      } else {
        if (summariesByVariable[variable.key]) {
          // if we have any saved state for that variable
          // use that as placeholder due to vue lifecycle
          const tempVariableSummaryKey = Object.keys(
            summariesByVariable[variable.key]
          )[0];
          promises.push(
            summariesByVariable[variable.key][tempVariableSummaryKey]
          );
        } else {
          // add a loading placeholder if nothing exists for that variable
          mutator(
            context,
            createPendingSummary(
              variable.key,
              variable.colDisplayName,
              variable.colDescription,
              args.dataset
            )
          );
        }

        // Get the mode or default
        const mode = args.varModes.has(variable.key)
          ? args.varModes.get(variable.key)
          : SummaryMode.Default;

        // fetch summary
        promises.push(
          actions.fetchVariableSummary(context, {
            dataset: args.dataset,
            variable: variable.key,
            filterParams: args.filterParams,
            highlights: args.highlights,
            include: args.include,
            dataMode: args.dataMode,
            mode: mode,
          })
        );
      }
    });
    // fill them in asynchronously
    return Promise.all(promises);
  },

  async fetchVariableSummary(
    context: DatasetContext,
    args: {
      dataset: string;
      variable: string;
      highlights?: Highlight[];
      filterParams: FilterParams;
      include: boolean;
      dataMode: DataMode;
      mode: SummaryMode;
    }
  ): Promise<void> {
    if (!validateArgs(args, ["dataset", "variable"])) {
      return null;
    }
    let filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlights
    );
    filterParams = setInvert(filterParams, !args.include);
    const decodedVarModes = routeGetters.getDecodedVarModes(store);
    const mutator = args.include
      ? mutations.updateIncludedVariableSummaries
      : mutations.updateExcludedVariableSummaries;
    const varMode = decodedVarModes.get(args.variable)
      ? decodedVarModes.get(args.variable)
      : args.mode;
    const dataModeDefault = routeGetters.getDataMode(store);
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(
        `/distil/variable-summary/${args.dataset}/${args.variable}/${varMode}`,
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
      target: args.target,
    };

    // quick exit if we don't have variables/target that are going to yield ranking
    const target = getters.getVariablesMap(context)[args.target];
    const rankableVariables = getters
      .getVariables(context)
      .filter((f) => f.key !== target.key && isRankableVariableType(f.colType));
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

      const rankings = response.data as Dictionary<number>;

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
        } else {
          status = DatasetPendingRequestStatus.RESOLVED;
        }
      }

      // Update the status.
      mutations.updatePendingRequests(context, { ...update, status, rankings });
    } catch (error) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: DatasetPendingRequestStatus.ERROR,
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
      rankings: args.rankings,
    });
    mutations.updateVariableRankings(context, args.rankings);
  },

  updatePendingRequestStatus(
    context: DatasetContext,
    args: { id: string; status: DatasetPendingRequestStatus }
  ) {
    const update = context.getters.getPendingRequests.find(
      (item) => item.id === args.id
    );
    if (update) {
      mutations.updatePendingRequests(context, {
        ...update,
        status: args.status,
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
      args.urls.map((url) => {
        if (type === IMAGE_TYPE) {
          return actions.fetchImage(context, {
            dataset: args.dataset,
            url: url,
          });
        }
        if (type === MULTIBAND_IMAGE_TYPE) {
          return actions.fetchMultiBandImage(context, {
            dataset: args.dataset,
            imageId: url,
            bandCombination: BandID.NATURAL_COLORS,
            isThumbnail: true,
          });
        }
        if (type === "graph") {
          return actions.fetchGraph(context, {
            dataset: args.dataset,
            url: url,
          });
        }
        return actions.fetchFile(context, {
          dataset: args.dataset,
          url: url,
        });
      })
    );
  },

  async fetchImage(
    context: DatasetContext,
    args: { dataset: string; url: string; isThumbnail?: boolean }
  ) {
    if (!validateArgs(args, ["dataset", "url"])) return;
    try {
      const thumbnail = args.isThumbnail ? "true" : "false";
      const urlRequest = `distil/image/${args.dataset}/${args.url}/${thumbnail}`;
      const response = await loadImage(urlRequest);
      mutations.updateFile(context, { url: args.url, file: response });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchTimeseries(
    context: DatasetContext,
    args: {
      dataset: string;
      variableKey: string;
      xColName: string;
      yColName: string;
      timeseriesIds: string[];
      uniqueTrail?: string;
    }
  ) {
    // format the data
    const timeseriesIDs = args.timeseriesIds.map((seriesID) => ({
      seriesID: seriesID,
      varKey: args.variableKey,
    }));

    try {
      const response = await axios.post<TimeSeriesUpdate[]>(
        `distil/timeseries/${encodeURIComponent(
          args.dataset
        )}/${encodeURIComponent(args.variableKey)}/${encodeURIComponent(
          args.xColName
        )}/${encodeURIComponent(args.yColName)}/false`,
        { timeseries: timeseriesIDs }
      );
      mutations.bulkUpdateTimeseries(context, {
        dataset: args.dataset,
        uniqueTrail: args.uniqueTrail,
        updates: response.data,
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
      isThumbnail: boolean;
      uniqueTrail?: string;
      options?: { gamma: number; gain: number; gainL: number };
    }
  ) {
    if (!validateArgs(args, ["dataset", "imageId", "bandCombination"])) {
      return null;
    }
    const options = !!args.options ? `${JSON.stringify(args.options)}` : "";
    try {
      const response = await loadImage(
        `distil/multiband-image/${args.dataset}/${args.imageId}/${args.bandCombination}/${args.isThumbnail}/${options}`
      );
      const imageUrl = !!args.uniqueTrail
        ? `${args.imageId}/${args.uniqueTrail}`
        : args.imageId;
      mutations.updateFile(context, { url: imageUrl, file: response });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchImagePack(
    context: DatasetContext,
    args: {
      multiBandImagePackRequest: MultiBandImagePackRequest;
      uniqueTrail?: string;
    }
  ) {
    try {
      const response = await axios.post(
        "distil/image-pack",
        args.multiBandImagePackRequest
      );
      let urls = response.data.imageIds;
      if (args.uniqueTrail) {
        urls = response.data.imageIds.map((id) => {
          return `${id}/${args.uniqueTrail}`;
        });
      }
      response.data.errorIds.forEach((id) => {
        console.error(`Error fetching image with ${id}`);
      });
      mutations.bulkUpdateFiles(context, {
        urls: urls,
        files: response.data.images,
      });
    } catch (error) {
      console.error(error);
    }
  },
  async fetchImageAttention(
    context: DatasetContext,
    args: {
      dataset: string;
      resultId: string;
      d3mIndex: number;
      opacity?: number;
    }
  ) {
    try {
      const colorScale = routeGetters.getColorScale(store);
      const response = await loadImage(
        `distil/image-attention/${args.dataset}/${args.resultId}/${
          args.d3mIndex
        }/${args.opacity ?? 100}/${colorScale}`
      );
      mutations.updateFile(context, {
        url: args.resultId + args.d3mIndex,
        file: response,
      });
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
          nodes: graph.nodes.map((n) => {
            return {
              id: n.id,
              label: n.label,
              x: n.attributes.attr1,
              y: n.attributes.attr2,
              size: 1,
              color: "#ec5148",
            };
          }),
          edges: graph.edges.map((e, i) => {
            return {
              id: `e${i}`,
              source: e.source,
              target: e.target,
              color: "#aaa",
            };
          }),
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
      highlights: Dictionary<Highlight[]>;
    }
  ) {
    if (!validateArgs(args, ["datasets", "filterParams"])) {
      return null;
    }
    return Promise.all(
      args.datasets.map(async (dataset) => {
        const highlights = args.highlights[dataset];
        let filterParams = addHighlightToFilterParams(
          args.filterParams[dataset],
          highlights
        );
        filterParams = setInvert(filterParams, false);
        try {
          const response = await axios.post(
            `distil/data/${dataset}/true`,
            filterParams
          );
          mutations.setJoinDatasetsTableData(context, {
            dataset: dataset,
            data: response.data,
          });
        } catch (error) {
          console.error(error);
          mutations.setJoinDatasetsTableData(context, {
            dataset: dataset,
            data: createEmptyTableData(),
          });
        }
      })
    );
  },

  async fetchIncludedTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      dataMode: DataMode;
      orderBy?: string;
    }
  ) {
    const data = await actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: args.filterParams,
      highlights: args.highlights,
      include: true,
      dataMode: args.dataMode,
      orderBy: args.orderBy,
    });
    mutations.setIncludedTableData(context, data);
  },

  async fetchExcludedTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      dataMode: DataMode;
    }
  ) {
    const filterParams = cloneFilters(args.filterParams);
    filterParams.highlights.invert = false;
    const data = await actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: filterParams,
      highlights: args.highlights,
      include: false,
      dataMode: args.dataMode,
    });
    mutations.setExcludedTableData(context, data);
  },
  async fetchBaselineTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      dataMode: DataMode;
    }
  ) {
    const mutator = mutations.setBaselineIncludeTableData;
    const data = await actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: args.filterParams,
      highlights: args.highlights,
      include: true,
      dataMode: args.dataMode,
    });
    mutator(context, data);
  },
  async fetchAreaOfInterestData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      dataMode: DataMode;
      include: boolean;
      mutatorIsInclude: boolean;
      isExclude: boolean;
    }
  ) {
    let mutator = null;
    if (!args.isExclude) {
      mutator = args.mutatorIsInclude
        ? mutations.setAreaOfInterestIncludeInner
        : mutations.setAreaOfInterestIncludeOuter;
    } else {
      mutator = args.mutatorIsInclude
        ? mutations.setAreaOfInterestExcludeInner
        : mutations.setAreaOfInterestExcludeOuter;
    }
    // if is exclude and highlight is null there is nothing to invert
    if (!args.mutatorIsInclude && !args.highlights.length) {
      mutator(context, { values: [] });
      return;
    }
    const data = await actions.fetchTableData(context, {
      dataset: args.dataset,
      filterParams: args.filterParams,
      highlights: args.highlights,
      include: args.include,
      dataMode: args.dataMode,
    });
    mutator(context, data);
  },

  async fetchTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      include: boolean;
      dataMode: DataMode;
      mode?: string;
      orderBy?: string;
    }
  ): Promise<TableData> {
    if (!validateArgs(args, ["dataset", "filterParams"])) {
      return null;
    }
    let filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlights,
      args.mode
    );
    filterParams = setInvert(filterParams, !args.include);
    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(`distil/data/${args.dataset}/false`, {
        ...filterParams,
        orderBy: args.orderBy,
      });
      return response.data;
    } catch (error) {
      console.error(error);
      return createEmptyTableData();
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
  },
  async cloneDataset(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<ClonedInfo> {
    // check for valid dataset
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    try {
      const response = await axios.post(`distil/clone/${args.dataset}`);
      return response.data;
    } catch (error) {
      console.error(error);
      return null;
    }
  },
  async fetchModelingMetrics(context: DatasetContext, args: { task: string }) {
    if (!validateArgs(args, ["task"])) {
      return null;
    }

    try {
      const repsonse = await axios.get<Metrics>(
        `distil/model-metrics/${args.task}`
      );
      const metrics = repsonse.data.metrics;
      mutations.updateMetrics(context, metrics);
    } catch (error) {
      console.error(error);
    }
  },

  updateRowSelectionData(context: DatasetContext): void {
    mutations.updateRowSelectionData(context);
  },
  async addField<T>(
    context: DatasetContext,
    args: {
      dataset: string;
      name: string;
      fieldType: string;
      defaultValue?: T;
      displayName?: string;
    }
  ) {
    // check for valid dataset
    if (!validateArgs(args, ["dataset"])) {
      return null;
    }
    try {
      const response = await axios.post(`distil/add-field/${args.dataset}`, {
        name: args.name,
        fieldType: args.fieldType,
        defaultValue: args.defaultValue.toString(),
        displayName: args.displayName,
      });
      return response.data;
    } catch (error) {
      console.error(error);
      return null;
    }
  },
  async saveDataset(
    context: DatasetContext,
    args: {
      dataset: string;
      datasetNewName: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      include: boolean;
      dataMode: DataMode;
      mode?: string;
    }
  ) {
    if (!validateArgs(args, ["dataset", "filterParams"])) {
      return null;
    }
    let filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlights,
      args.mode
    );
    filterParams = setInvert(filterParams, args.include);
    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(`distil/save-dataset/${args.dataset}`, {
        datasetName: args.datasetNewName,
        ...filterParams,
      });
      return response.data;
    } catch (error) {
      console.error(error);
      return null;
    }
  },
  async extractDataset(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlights: Highlight[];
      include: boolean;
      dataMode: DataMode;
      mode?: string;
    }
  ) {
    if (!validateArgs(args, ["dataset", "filterParams"])) {
      return null;
    }
    let filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlights,
      args.mode
    );
    filterParams = setInvert(filterParams, !args.include);
    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(
        `distil/extract/${args.dataset}`,
        filterParams
      );
      return response.data;
    } catch (error) {
      console.error(error);
      return null;
    }
  },
};
