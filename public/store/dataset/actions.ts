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
  SummaryMode
} from "./index";
import { mutations, getters } from "./module";
import { DistilState } from "../store";
import { Highlight } from "../dataset/index";
import { FilterParams } from "../../util/filters";
import {
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars
} from "../../util/data";
import { addHighlightToFilterParams } from "../../util/highlights";
import { loadImage } from "../../util/image";
import {
  getVarType,
  IMAGE_TYPE,
  GEOCODED_LON_PREFIX,
  GEOCODED_LAT_PREFIX,
  GEOCOORDINATE_TYPE
} from "../../util/types";

import { DATASET_UPLOAD, PREDICTION_UPLOAD } from "../../util/uploads";

// fetches variables and add dataset name to each variable
function getVariables(dataset: string): Promise<Variable[]> {
  return axios.get(`/distil/variables/${dataset}`).then(response => {
    // extend variable with datasetName and isColTypeReviewed property to track type reviewed state in the client state
    return response.data.variables.map(variable => ({
      ...variable,
      datasetName: dataset,
      isColTypeReviewed: false
    }));
  });
}

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {
  // fetches a dataset description.
  fetchDataset(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<void> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    return axios
      .get(`/distil/datasets/${args.dataset}`)
      .then(response => {
        mutations.setDataset(context, response.data.dataset);
      })
      .catch(error => {
        console.error(error);
        mutations.setDatasets(context, []);
      });
  },

  // searches dataset descriptions and column names for supplied terms
  searchDatasets(context: DatasetContext, terms: string): Promise<void> {
    const params = !_.isEmpty(terms) ? `?search=${terms}` : "";
    return axios
      .get(`/distil/datasets${params}`)
      .then(response => {
        mutations.setDatasets(context, response.data.datasets);
      })
      .catch(error => {
        console.error(error);
        mutations.setDatasets(context, []);
      });
  },

  // fetches all variables for a single dataset.
  fetchVariables(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<void> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    return getVariables(args.dataset)
      .then(variables => {
        mutations.setVariables(context, variables);
      })
      .catch(error => {
        console.error(error);
        mutations.setVariables(context, []);
      });
  },

  // fetches all variables for a two datasets.
  fetchJoinDatasetsVariables(
    context: DatasetContext,
    args: { datasets: string[] }
  ): Promise<void> {
    if (!args.datasets) {
      console.warn("`datasets` argument is missing");
      return null;
    }
    return Promise.all([
      getVariables(args.datasets[0]),
      getVariables(args.datasets[1])
    ])
      .then(res => {
        const varsA = res[0];
        const varsB = res[1];
        mutations.setVariables(context, varsA.concat(varsB));
      })
      .catch(error => {
        console.error(error);
        mutations.setVariables(context, []);
      });
  },

  geocodeVariable(
    context: DatasetContext,
    args: { dataset: string; field: string }
  ): Promise<any> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.field) {
      console.warn("`field` argument is missing");
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
    return axios
      .post(`/distil/geocode/${args.dataset}/${args.field}`, {})
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

  fetchJoinSuggestions(
    context: DatasetContext,
    args: { dataset: string; searchQuery: string }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
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

    // Hack: force to include datamart.upload.fc0ceee28cb74bad83e4f8872979b111 to the result since that data set does not appear on the suggestion list.
    /*
		return axios.get(`/distil/datasets/${args.dataset}`)
			.then(res => {
				const dataset = res.data.dataset;
				const search = dataset.summaryML || dataset.summary || '';
				return Promise.all([
					// axios.get(`/distil/join-suggestions/${args.dataset}`, { params: { search } }).catch(e => ({data: undefined})),
					axios.get(`/distil/datasets`, { params: { search: 'employment' } }),
				]);
			})
			.then((response) => {
				// const suggestions = (response[0].data && response[0].data.datasets) || [];
				const employmentData = ((response[0].data && response[0].data.datasets) || []).filter(dataset =>
					dataset.id === 'datamart.upload.fc0ceee28cb74bad83e4f8872979b111' ||
					dataset.id === 'world_bank_2018');
				mutations.updatePendingRequests(context, { ...request, status: DatasetPendingRequestStatus.RESOLVED, suggestions: [...employmentData] });
			}).catch(error => {
				mutations.updatePendingRequests(context, { ...request, status: DatasetPendingRequestStatus.ERROR });
				console.error(error);
			});
		*/
    const query = args.searchQuery
      ? `?search=${args.searchQuery.split(" ").join(",")}`
      : "";
    return axios
      .get(`/distil/join-suggestions/${args.dataset + query}`)
      .then(response => {
        const suggestions = (response.data && response.data.datasets) || [];
        mutations.updatePendingRequests(context, {
          ...request,
          status: DatasetPendingRequestStatus.RESOLVED,
          suggestions
        });
      })
      .catch(error => {
        mutations.updatePendingRequests(context, {
          ...request,
          status: DatasetPendingRequestStatus.ERROR
        });
        console.error(error);
      });
  },

  // Sends a request to the server to generate cluaster for all data that is a valid target for clustering.
  fetchClusters(
    context: DatasetContext,
    args: { dataset: string }
  ): Promise<any> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    const update: ClusteringPendingRequest = {
      id: _.uniqueId(),
      dataset: args.dataset,
      type: DatasetPendingRequestType.CLUSTERING,
      status: DatasetPendingRequestStatus.PENDING
    };
    mutations.updatePendingRequests(context, update);

    // Find grouped fields that have clusters defined against them and request that they
    // cluster.
    const promises = getters
      .getVariables(context)
      .filter(v => (v.grouping && v.grouping.properties.clusterCol) || (v.colType === "image"))
      .map(v => {
        if (v.grouping && v.grouping.properties.clusterCol) { axios.post(`/distil/cluster/${args.dataset}/${v.grouping.idCol}`, {}); }
        else if (v.colType === "image") {axios.post(`/distil/cluster/${args.dataset}/${v.colName}`, {});}
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

  uploadDataFile(
    context: DatasetContext,
    args: { datasetID: string; file: File; type: string; solutionId?: string }
  ): any {
    if (!args.datasetID) {
      console.warn("`datasetID` argument is missing");
      return null;
    }
    if (!args.file) {
      console.warn("`file` argument is missing");
      return null;
    }
    if (!args.type) {
      console.warn("`type` argument is missing");
      return null;
    }
    const data = new FormData();
    data.append("file", args.file);

    switch (args.type) {
      case PREDICTION_UPLOAD:
        if (!args.solutionId) {
          console.warn("`solutionId` argument is missing");
          return null;
        }
        return axios
          .post(`/distil/predict/${args.datasetID}/tabular/${args.solutionId}`, data, {
            headers: { "Content-Type": "multipart/form-data" }
          })
          .then(response => {
            return response;
          });
      case DATASET_UPLOAD:
        return axios
          .post(`/distil/upload/${args.datasetID}?type=table`, data, {
            headers: { "Content-Type": "multipart/form-data" }
          })
          .then(response => {
            return actions.importDataset(context, {
              datasetID: args.datasetID,
              source: "augmented",
              provenance: "local",
              terms: args.datasetID,
              originalDataset: null,
              joinedDataset: null
            });
          });
      default:
        console.log("unknown upload type");
    }
  },

  importDataset(
    context: DatasetContext,
    args: {
      datasetID: string;
      source: string;
      provenance: string;
      terms: string;
      originalDataset: Dataset;
      joinedDataset: Dataset;
    }
  ): Promise<void> {
    if (!args.datasetID) {
      console.warn("`datasetID` argument is missing");
      return null;
    }
    if (!args.source) {
      console.warn("`source` argument is missing");
      return null;
    }

    let postParams = {};
    if (args.originalDataset !== null) {
      postParams = {
        originalDataset: args.originalDataset,
        joinedDataset: args.joinedDataset
      };
    }

    return axios
      .post(
        `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
        postParams
      )
      .then(response => {
        return actions.searchDatasets(context, args.terms);
      });
  },

  importJoinDataset(
    context: DatasetContext,
    args: {
      datasetID: string;
      source: string;
      provenance: string;
      searchResults: DatasetOrigin[];
    }
  ): Promise<any> {
    if (!args.datasetID && args.datasetID.length > 0) {
      console.warn("`datasetID` argument is missing");
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
    return axios
      .post(
        `/distil/import/${args.datasetID}/${args.source}/${args.provenance}`,
        {
          searchResults: args.searchResults
        }
      )
      .then(response => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.RESOLVED
        });
        return response && response.data;
      })
      .catch(error => {
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.ERROR
        });
        console.error(error);
      });
  },

  composeVariables(
    context: DatasetContext,
    args: { dataset: string; key: string; vars: string[] }
  ): Promise<void> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.key) {
      console.warn("`key` argument is missing");
      return null;
    }
    if (!args.vars) {
      console.warn("`vars` argument is missing");
      return null;
    }
    return axios.post(`/distil/compose/${args.dataset}`, {
      varName: args.key,
      variables: args.vars
    });
  },

  deleteVariable(
    context: DatasetContext,
    args: { dataset: string; key: string }
  ): Promise<any> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.key) {
      console.warn("`key` argument is missing");
      return null;
    }
    return axios
      .post(`/distil/delete/${args.dataset}/${args.key}`, {})
      .then(() => {
        // update dataset
        return Promise.all([
          actions.fetchDataset(context, {
            dataset: args.dataset
          }),
          actions.fetchVariables(context, {
            dataset: args.dataset
          })
        ]).then(() => {
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
        });
      })
      .catch(error => {
        console.error(error);
      });
  },

  joinDatasetsPreview(
    context: DatasetContext,
    args: {
      datasetA: Dataset;
      datasetB: Dataset;
      joinAccuracy: number;
      joinSuggestionIndex: number;
    }
  ): Promise<void> {
    if (!args.datasetA) {
      console.warn("`datasetA` argument is missing");
      return null;
    }
    if (!args.datasetB) {
      console.warn("`datasetB` argument is missing");
      return null;
    }

    if (_.isNil(args.joinAccuracy)) {
      console.warn("`joinAccuracy` argument is missing");
      return null;
    }

    const datasetBrevised: Dataset = JSON.parse(JSON.stringify(args.datasetB));

    datasetBrevised.variables = datasetBrevised.variables.map(v => {
      const roledVar = v;
      roledVar.role = ["attribute"];
      return roledVar;
    });

    return axios
      .post(`/distil/join`, {
        accuracy: args.joinAccuracy,
        datasetLeft: args.datasetA,
        datasetRight: datasetBrevised,
        searchResultIndex: args.joinSuggestionIndex
      })
      .then(response => {
        return response.data;
      });
  },

  setGrouping(
    context: DatasetContext,
    args: { dataset: string; grouping: Grouping }
  ): Promise<any> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.grouping) {
      console.warn("`grouping` argument is missing");
      return null;
    }
    return axios
      .post(`/distil/grouping/${args.dataset}`, {
        grouping: args.grouping
      })
      .then(() => {
        // update dataset
        return Promise.all([
          actions.fetchDataset(context, {
            dataset: args.dataset
          }),
          actions.fetchVariables(context, {
            dataset: args.dataset
          })
        ]).then(() => {
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
        });
      })
      .catch(error => {
        console.error(error);
      });
  },

  removeGrouping(
    context: DatasetContext,
    args: { dataset: string; variable: string }
  ): Promise<any> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.variable) {
      console.warn("`grouping` argument is missing");
      return null;
    }
    return axios
      .post(`/distil/remove-grouping/${args.dataset}/${args.variable}`, {})
      .then(() => {
        // update dataset
        return Promise.all([
          actions.fetchDataset(context, {
            dataset: args.dataset
          }),
          actions.fetchVariables(context, {
            dataset: args.dataset
          })
        ]).then(() => {
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
        });
      })
      .catch(error => {
        console.error(error);
      });
  },

  setVariableType(
    context: DatasetContext,
    args: { dataset: string; field: string; type: string }
  ): Promise<any> {
    if (args.type === GEOCOORDINATE_TYPE) {
      mutations.updateVariableType(context, args);
      return;
    }
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.field) {
      console.warn("`field` argument is missing");
      return null;
    }
    if (!args.type) {
      console.warn("`type` argument is missing");
      return null;
    }

    return axios
      .post(`/distil/variables/${args.dataset}`, {
        field: args.field,
        type: args.type
      })
      .then(() => {
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
      })
      .catch(error => {
        const key = args.field;
        const label = args.field;
        const dataset = args.dataset;
        mutations.updateIncludedVariableSummaries(
          context,
          createErrorSummary(key, label, dataset, error)
        );
        mutations.updateExcludedVariableSummaries(
          context,
          createErrorSummary(key, label, dataset, error)
        );
      });
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
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.variables) {
      console.warn("`variables` argument is missing");
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

  fetchVariableSummary(
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
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.variable) {
      console.warn("`variable` argument is missing");
      return null;
    }

    const filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlight
    );
    const mutator = args.include
      ? mutations.updateIncludedVariableSummaries
      : mutations.updateExcludedVariableSummaries;

    return axios
      .post(
        `/distil/variable-summary/${args.dataset}/${
          args.variable
        }/${!args.include}/${args.mode}`,
        filterParams
      )
      .then(response => {
        const summary = response.data.summary;
        return fetchSummaryExemplars(args.dataset, args.variable, summary).then(
          () => {
            mutator(context, summary);
          }
        );
      })
      .catch(error => {
        console.error(error);
        const key = args.variable;
        const label = args.variable;
        const dataset = args.dataset;
        mutator(context, createErrorSummary(key, label, dataset, error));
      });
  },

  fetchVariableRankings(
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
    mutations.updatePendingRequests(context, update);
    return axios
      .get(`/distil/variable-rankings/${args.dataset}/${args.target}`)
      .then(response => {
        // console.log({response, data: response.data, rankings: response.data.rankings});
        mutations.updatePendingRequests(context, {
          ...update,
          status: DatasetPendingRequestStatus.RESOLVED,
          rankings: response.data.rankings
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
    if (!args.urls) {
      console.warn("`url` argument is missing");
      return null;
    }
    const type = getVarType(args.variable);
    return Promise.all(
      args.urls.map(url => {
        if (type === IMAGE_TYPE) {
          return actions.fetchImage(context, {
            dataset: args.dataset,
            source: "seed",
            url: url
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

  fetchImage(
    context: DatasetContext,
    args: { dataset: string; source: string; url: string }
  ) {
    if (!args.url) {
      console.warn("`url` argument is missing");
      return null;
    }
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    return loadImage(`distil/image/${args.dataset}/${args.source}/${args.url}`)
      .then(response => {
        mutations.updateFile(context, { url: args.url, file: response });
      })
      .catch(error => {
        console.error(error);
      });
  },

  fetchTimeseries(
    context: DatasetContext,
    args: {
      dataset: string;
      xColName: string;
      yColName: string;
      timeseriesColName: string;
      timeseriesID: any;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.xColName) {
      console.warn("`xColName` argument is missing");
      return null;
    }
    if (!args.yColName) {
      console.warn("`yColName` argument is missing");
      return null;
    }
    if (!args.timeseriesColName) {
      console.warn("`timeseriesColName` argument is missing");
      return null;
    }
    if (!args.timeseriesID) {
      console.warn("`timeseriesID` argument is missing");
      return null;
    }

    return axios
      .post(
        `distil/timeseries/${args.dataset}/${args.timeseriesColName}/${args.xColName}/${args.yColName}/${args.timeseriesID}/false`,
        {}
      )
      .then(response => {
        mutations.updateTimeseries(context, {
          dataset: args.dataset,
          id: args.timeseriesID,
          timeseries: response.data.timeseries
        });
      })
      .catch(error => {
        console.error(error);
      });
  },

  fetchGraph(context: DatasetContext, args: { dataset: string; url: string }) {
    if (!args.url) {
      console.warn("`url` argument is missing");
      return null;
    }
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    return axios
      .get(`distil/graphs/${args.dataset}/${args.url}`)
      .then(response => {
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
      })
      .catch(error => {
        console.error(error);
      });
  },

  fetchFile(context: DatasetContext, args: { dataset: string; url: string }) {
    if (!args.url) {
      console.warn("`url` argument is missing");
      return null;
    }
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    return axios
      .get(`distil/resource/${args.dataset}/${args.url}`)
      .then(response => {
        mutations.updateFile(context, { url: args.url, file: response.data });
      })
      .catch(error => {
        console.error(error);
      });
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
    if (!args.datasets) {
      console.warn("`datasets` argument is missing");
      return null;
    }
    if (!args.filterParams) {
      console.warn("`filterParams` argument is missing");
      return null;
    }
    return Promise.all(
      args.datasets.map(dataset => {
        const highlight =
          (args.highlight && args.highlight.dataset) === dataset
            ? args.highlight
            : null;
        const filterParams = addHighlightToFilterParams(
          args.filterParams[dataset],
          highlight
        );

        return axios
          .post(`distil/data/${dataset}/false`, filterParams)
          .then(response => {
            mutations.setJoinDatasetsTableData(context, {
              dataset: dataset,
              data: response.data
            });
          })
          .catch(error => {
            console.error(error);
            mutations.setJoinDatasetsTableData(context, {
              dataset: dataset,
              data: createEmptyTableData()
            });
          });
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

  fetchTableData(
    context: DatasetContext,
    args: {
      dataset: string;
      filterParams: FilterParams;
      highlight: Highlight;
      include: boolean;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.filterParams) {
      console.warn("`filterParams` argument is missing");
      return null;
    }

    const mutator = args.include
      ? mutations.setIncludedTableData
      : mutations.setExcludedTableData;

    const filterParams = addHighlightToFilterParams(
      args.filterParams,
      args.highlight
    );

    return axios
      .post(`distil/data/${args.dataset}/${!args.include}`, filterParams)
      .then(response => {
        mutator(context, response.data);
      })
      .catch(error => {
        console.error(error);
        mutator(context, createEmptyTableData());
      });
  },

  fetchTask(
    context: DatasetContext,
    args: { dataset: string; targetName: string }
  ): Promise<AxiosResponse<Task>> {
    return axios.get<Task>(`/distil/task/${args.dataset}/${args.targetName}`);
  }
};
