import router from "../../router/router";
import {
  datasetActions,
  datasetGetters,
  predictionActions,
  predictionGetters,
  requestActions,
  requestGetters,
  resultActions,
  resultGetters,
  viewActions,
} from "../../store";
import {
  D3M_INDEX_FIELD,
  TableColumn,
  TableRow,
  TimeseriesGrouping,
  Variable,
  VariableSummary,
} from "../../store/dataset";
import { getters as routeGetters } from "../../store/route/module";
import store from "../../store/store";
import {
  getAllVariablesSummaries,
  LOW_SHOT_SCORE_COLUMN_PREFIX,
} from "../data";
import { ExplorerStateNames } from "../dataExplorer";
import { Dictionary } from "../dict";
import { EI } from "../events";
import { Filter } from "../filters";
import { overlayRouteEntry } from "../routes";
import { getSolutionById } from "../solutions";
import {
  getConfidenceSummary,
  getCorrectnessSummary,
  getPredictionConfidenceSummary,
  getPredictionRankSummary,
  getPredictionResultSummary,
  getRankingSummary,
  getResidualSummary,
  getSolutionResultSummary,
  resultSummariesToVariables,
  summaryToVariable,
} from "../summaries";
import { DISTIL_ROLES } from "../types";

export interface BaseState {
  name: ExplorerStateNames;
  // gets basic variables
  getVariables(): Variable[];
  // gets secondary variables related to secondary variableSummaries
  getSecondaryVariables(): Variable[];
  // basic data for tables and maps
  getData(include?: boolean): TableRow[];
  // base variable summaries is the standard variables in a view. For select it is the available variables,
  //for result it is the training variables
  getBaseVariableSummaries(include?: boolean): VariableSummary[];
  // this is the availableVariables, result summaries, prediction summaries
  getSecondaryVariableSummaries(include?: boolean): VariableSummary[];
  // allSummaries
  getAllVariableSummaries(include?: boolean): VariableSummary[];
  // target variable, ground truth, prediction variable
  getTargetVariable(): Variable;
  // map baseline data
  getMapBaseline(): TableRow[];
  // drill down baseline for map
  getMapDrillDownBaseline(include?: boolean): TableRow[];
  // get filtered data for map drill down
  getMapDrillDownFiltered(include?: boolean): TableRow[];
  // variables used for lexbar
  getLexBarVariables(): Variable[];
  // gets table data fields
  getFields(include?: boolean): Dictionary<TableColumn>;
  /******Fetch Functions**********/
  init(): Promise<void>;
  fetchVariables(): Promise<unknown>;
  fetchData(): Promise<unknown>;
  fetchVariableSummaries(): Promise<unknown>;
  fetchMapBaseline(): Promise<void>;
  fetchMapDrillDown(filter: Filter): Promise<Array<void>>;
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent);
}

export class SelectViewState implements BaseState {
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent) {
    args.variables.forEach((tsv) => {
      const grouping = tsv.grouping as TimeseriesGrouping;
      datasetActions.fetchTimeseries(store, {
        dataset: routeGetters.getRouteDataset(store),
        variableKey: tsv.key,
        xColName: grouping.xCol,
        yColName: grouping.yCol,
        uniqueTrail: args.uniqueTrail,
        timeseriesIds: args.timeseriesIds.map((item) => {
          return item[tsv.key].value as string;
        }),
      });
    });
  }
  getTargetVariable(): Variable {
    return routeGetters.getTargetVariable(store);
  }
  name = ExplorerStateNames.SELECT_VIEW;

  async init(): Promise<void> {
    await this.fetchVariables();
    await this.fetchMapBaseline();
    this.fetchVariableSummaries();
    return;
  }
  getSecondaryVariables(): Variable[] {
    return routeGetters.getAvailableVariables(store);
  }
  getAllVariableSummaries(include?: boolean): VariableSummary[] {
    const varDict = include
      ? datasetGetters.getIncludedVariableSummariesDictionary(store)
      : datasetGetters.getExcludedVariableSummariesDictionary(store);
    return getAllVariablesSummaries(this.getLexBarVariables(), varDict);
  }
  getSecondaryVariableSummaries(include?: boolean): VariableSummary[] {
    const varDict = include
      ? datasetGetters.getIncludedVariableSummariesDictionary(store)
      : datasetGetters.getExcludedVariableSummariesDictionary(store);
    const variables = routeGetters.getAvailableVariables(store);
    return getAllVariablesSummaries(variables, varDict);
  }
  fetchMapDrillDown(filter: Filter): Promise<Array<void>> {
    return viewActions.updateAreaOfInterest(store, filter);
  }
  // returns select view table fields
  getFields(include: boolean): Dictionary<TableColumn> {
    return include
      ? datasetGetters.getIncludedTableDataFields(store)
      : datasetGetters.getExcludedTableDataFields(store);
  }
  // returns select view variables
  getVariables(): Variable[] {
    return datasetGetters.getVariables(store);
  }
  // returns table data based on include
  getData(include: boolean): TableRow[] {
    const res = include
      ? datasetGetters.getIncludedTableDataItems(store)
      : datasetGetters.getExcludedTableDataItems(store);
    return res ?? [];
  }
  // returns training variables
  getBaseVariableSummaries(include: boolean): VariableSummary[] {
    const varDict = include
      ? datasetGetters.getIncludedVariableSummariesDictionary(store)
      : datasetGetters.getExcludedVariableSummariesDictionary(store);
    const variables = routeGetters.getTrainingVariables(store);
    return getAllVariablesSummaries(variables, varDict);
  }
  // returns select view map baseline
  getMapBaseline(): TableRow[] {
    const bItems = datasetGetters.getBaselineIncludeTableDataItems(store) ?? [];
    return bItems.sort((a, b) => {
      return a[D3M_INDEX_FIELD] - b[D3M_INDEX_FIELD];
    });
  }
  // returns all the tiles within the clicked area
  getMapDrillDownBaseline(include: boolean): TableRow[] {
    return include
      ? datasetGetters.getAreaOfInterestIncludeInnerItems(store)
      : datasetGetters.getAreaOfInterestExcludeInnerItems(store);
  }
  // returns all the tiles matching the current highlight/filter in clicked area
  getMapDrillDownFiltered(include: boolean): TableRow[] {
    return include
      ? datasetGetters.getAreaOfInterestIncludeOuterItems(store)
      : datasetGetters.getAreaOfInterestExcludeOuterItems(store);
  }
  getLexBarVariables(): Variable[] {
    return datasetGetters.getAllVariables(store);
  }
  fetchVariables(): Promise<unknown> {
    return viewActions.fetchSelectTrainingData(store, false);
  }
  fetchData(): Promise<unknown> {
    return viewActions.updateSelectTrainingData(store);
  }
  fetchVariableSummaries(): Promise<unknown> {
    const fetchArgs = {
      dataset: routeGetters.getRouteDataset(store),
      variables: this.getVariables().concat(this.getSecondaryVariables()),
      filterParams: routeGetters.getDecodedSolutionRequestFilterParams(store),
      highlights: routeGetters.getDecodedHighlights(store),
      dataMode: routeGetters.getDataMode(store),
      varModes: routeGetters.getDecodedVarModes(store),
    };

    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, fetchArgs),
      datasetActions.fetchExcludedVariableSummaries(store, fetchArgs),
    ]);
  }
  fetchMapBaseline(): Promise<void> {
    return viewActions.updateHighlight(store);
  }
}

export class ResultViewState implements BaseState {
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent) {
    args.variables.forEach((tsv) => {
      const tsg = tsv.grouping as TimeseriesGrouping;
      resultActions.fetchForecastedTimeseries(store, {
        dataset: routeGetters.getRouteDataset(store),
        variableKey: tsv.key,
        xColName: tsg.xCol,
        yColName: tsg.yCol,
        solutionId: routeGetters.getRouteSolutionId(store),
        uniqueTrail: args.uniqueTrail,
        timeseriesIds: args.timeseriesIds.map((item) => {
          return item[tsv.key].value as string;
        }),
      });
    });
  }
  name = ExplorerStateNames.RESULT_VIEW;
  getTargetVariable(): Variable {
    return routeGetters.getTargetVariable(store);
  }
  async init(): Promise<void> {
    // check if solutionId is not null if not find recent solution and make it the target solutionId
    if (!routeGetters.getRouteSolutionId(store)) {
      await requestActions.fetchSolutions(store, {
        dataset: routeGetters.getRouteDataset(store),
        target: routeGetters.getRouteTargetVariable(store),
      });
      const solutions = requestGetters.getSolutions(store);
      if (solutions && solutions.length) {
        // dont mutate store array
        const sorted = [...solutions].sort((a, b) => {
          return (
            new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
          );
        });
        const entry = overlayRouteEntry(routeGetters.getRoute(store), {
          solutionId: sorted[sorted.length - 1].solutionId,
        });
        router.push(entry).catch((err) => console.warn(err));
      } else {
        console.error("No available solutions");
      }
    }
    await this.fetchVariables();
    await this.fetchMapBaseline();
    return;
  }
  getSecondaryVariables(): Variable[] {
    const solutionId = routeGetters.getRouteSolutionId(store);
    if (!solutionId) {
      return [];
    }
    const solution = getSolutionById(
      requestGetters.getRelevantSolutions(store),
      solutionId
    );
    return resultSummariesToVariables(solution?.resultId);
  }
  getAllVariableSummaries(): VariableSummary[] {
    let res = [];
    const secondVarSums = this.getSecondaryVariableSummaries();
    const baseVarSums = this.getBaseVariableSummaries();
    if (secondVarSums.length) {
      res = res.concat(secondVarSums);
    }
    if (baseVarSums.length) {
      res = res.concat(baseVarSums);
    }
    return res;
  }

  getSecondaryVariableSummaries(): VariableSummary[] {
    const currentSummaries = [];
    const solution = requestGetters.getActiveSolution(store);
    if (!solution?.resultId) {
      return [];
    }
    const predictedSummary = getSolutionResultSummary(solution.resultId);
    if (predictedSummary) {
      currentSummaries.push(predictedSummary);
    }
    const residualSummary = getResidualSummary(solution?.resultId);
    if (residualSummary) {
      currentSummaries.push(residualSummary);
    }
    const correctnessSummary = getCorrectnessSummary(solution?.resultId);
    if (correctnessSummary) {
      currentSummaries.push(correctnessSummary);
    }
    const confidenceSummary = getConfidenceSummary(solution?.resultId);
    if (confidenceSummary) {
      currentSummaries.push(confidenceSummary);
    }
    const rankingSummary = getRankingSummary(solution?.resultId);
    if (rankingSummary) {
      currentSummaries.push(rankingSummary);
    }
    return currentSummaries;
  }
  fetchMapDrillDown(filter: Filter): Promise<Array<void>> {
    return viewActions.updateResultAreaOfInterest(store, filter);
  }
  getVariables(): Variable[] {
    return requestGetters.getActiveSolutionTrainingVariables(store);
  }
  getData(): TableRow[] {
    return resultGetters.getIncludedResultTableDataItems(store) ?? [];
  }
  getBaseVariableSummaries(): VariableSummary[] {
    const summaryDictionary = resultGetters.getTrainingSummariesDictionary(
      store
    );
    const variables = this.getVariables();
    const trainingSummaries = getAllVariablesSummaries(
      variables,
      summaryDictionary
    );
    const target = resultGetters.getTargetSummary(store);
    if (target) {
      return trainingSummaries.concat(target);
    }
    return trainingSummaries;
  }
  getMapBaseline(): TableRow[] {
    return resultGetters.getFullIncludedResultTableDataItems(store);
  }
  getMapDrillDownBaseline(): TableRow[] {
    return resultGetters.getAreaOfInterestInnerDataItems(store);
  }
  getMapDrillDownFiltered(): TableRow[] {
    return resultGetters.getAreaOfInterestOuterDataItems(store);
  }
  getLexBarVariables(): Variable[] {
    const result = [];
    const solutionId = routeGetters.getRouteSolutionId(store);
    if (!solutionId) {
      return [];
    }
    const solution = getSolutionById(
      requestGetters.getRelevantSolutions(store),
      solutionId
    );
    if (!solution?.resultId) {
      return [];
    }
    const target = requestGetters.getActiveSolutionTargetVariable(store);
    if (target) {
      result.push(target);
    }
    const resultVariables = resultSummariesToVariables(solution?.resultId);
    return result.concat(this.getVariables()).concat(resultVariables);
  }
  getFields(): Dictionary<TableColumn> {
    return resultGetters.getIncludedResultTableDataFields(store);
  }
  fetchVariables(): Promise<void> {
    return viewActions.fetchResultsData(store);
  }
  fetchData(): Promise<unknown> {
    return viewActions.updateResultsSolution(store);
  }
  fetchVariableSummaries(): Promise<unknown> {
    return viewActions.updateResultsSummaries(store);
  }
  fetchMapBaseline(): Promise<void> {
    return viewActions.updateResultBaseline(store);
  }
}

export class PredictViewState implements BaseState {
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent) {
    const activePredictions = requestGetters.getActivePredictions(store);
    args.variables.forEach(async (tsv) => {
      const tsg = tsv.grouping as TimeseriesGrouping;
      await predictionActions.fetchForecastedTimeseries(store, {
        truthDataset: routeGetters.getRouteDataset(store),
        forecastDataset: activePredictions.dataset,
        xColName: tsg.xCol,
        yColName: tsg.yCol,
        timeseriesColName: tsg.idCol,
        predictionsId: activePredictions.requestId,
        uniqueTrail: args.uniqueTrail,
        timeseriesIds: args.timeseriesIds.map((item) => {
          return item[tsg.idCol].value as string;
        }),
      });
    });
  }
  name = ExplorerStateNames.PREDICTION_VIEW;
  getTargetVariable(): Variable {
    const activePred = requestGetters.getActivePredictions(store);
    const predSum = getPredictionResultSummary(activePred?.requestId);
    if (!predSum) {
      return null;
    }
    return summaryToVariable(predSum);
  }
  getVariables(): Variable[] {
    return requestGetters.getActivePredictionTrainingVariables(store);
  }
  getSecondaryVariables(): Variable[] {
    const predictionVariables = [] as Variable[];
    const activePred = requestGetters.getActivePredictions(store);
    const rankSum = getPredictionRankSummary(activePred?.resultId);
    const confidenceSum = getPredictionConfidenceSummary(activePred?.resultId);
    const predSum = getPredictionResultSummary(activePred?.requestId);
    if (rankSum) {
      predictionVariables.push(summaryToVariable(rankSum));
    }
    if (confidenceSum) {
      predictionVariables.push(summaryToVariable(confidenceSum));
    }
    if (predSum) {
      predictionVariables.push(summaryToVariable(predSum));
    }
    return predictionVariables;
  }
  getData(): TableRow[] {
    return predictionGetters.getIncludedPredictionTableDataItems(store) ?? [];
  }
  getBaseVariableSummaries(): VariableSummary[] {
    const summaryDictionary = predictionGetters.getTrainingSummariesDictionary(
      store
    );
    return getAllVariablesSummaries(
      this.getVariables(),
      summaryDictionary,
      routeGetters.getRoutePredictionsDataset(store)
    );
  }
  getSecondaryVariableSummaries(): VariableSummary[] {
    const currentSummaries = [];
    const activePred = requestGetters.getActivePredictions(store);
    const rank = getPredictionRankSummary(activePred?.resultId);
    const confidence = getPredictionConfidenceSummary(activePred?.resultId);
    const summary = getPredictionResultSummary(activePred?.requestId);
    if (rank) {
      currentSummaries.push(rank);
    }
    if (confidence) {
      currentSummaries.push(confidence);
    }
    if (summary) {
      currentSummaries.push(summary);
    }
    return currentSummaries;
  }
  getAllVariableSummaries(): VariableSummary[] {
    return this.getBaseVariableSummaries().concat(
      this.getSecondaryVariableSummaries()
    );
  }
  getMapBaseline(): TableRow[] {
    const result = predictionGetters.getBaselinePredictionTableDataItems(store);
    return result?.sort((a, b) => {
      return a.d3mIndex - b.d3mIndex;
    });
  }
  getMapDrillDownBaseline(): TableRow[] {
    return predictionGetters.getAreaOfInterestInnerDataItems(store);
  }
  getMapDrillDownFiltered(): TableRow[] {
    return predictionGetters.getAreaOfInterestOuterDataItems(store);
  }
  getLexBarVariables(): Variable[] {
    return datasetGetters
      .getAllVariables(store)
      .concat(this.getSecondaryVariables());
  }
  getFields(): Dictionary<TableColumn> {
    return predictionGetters.getIncludedPredictionTableDataFields(store);
  }
  async init(): Promise<void> {
    const dataset = routeGetters.getRouteDataset(store);
    const target = routeGetters.getRouteTargetVariable(store);
    const produceRequest = routeGetters.getRouteProduceRequestId(store);
    if (!produceRequest) {
      await requestActions.fetchSolutions(store, { dataset, target });
      const solutions = requestGetters.getSolutions(store);
      let minTime = 0;
      let index = 0;
      if (solutions && solutions.length) {
        solutions.forEach((s, i) => {
          if (s.hasPredictions) {
            const time = new Date(s.timestamp).getTime();
            if (time > minTime) {
              index = i;
              minTime = time;
            }
          }
        });
        if (!minTime) {
          console.error("No Prediction Available");
          return;
        }
        await requestActions.fetchPredictions(store, {
          fittedSolutionId: solutions[index].fittedSolutionId,
        });
        const predictions = requestGetters.getPredictions(store);
        const sorted = [...predictions].sort((a, b) => {
          return (
            new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
          );
        });
        const route = routeGetters.getRoute(store);
        const end = sorted.length - 1;
        const entry = overlayRouteEntry(route, {
          produceRequestId: sorted[end].requestId,
          fittedSolutionId: solutions[index].fittedSolutionId,
          predictionsDataset: sorted[end].dataset,
          solutionId: solutions[index].solutionId,
        });
        router.push(entry).catch((err) => console.warn(err));
      }
    }
    await viewActions.fetchPredictionsData(store);
    datasetActions.fetchClusters(store, { dataset });
    datasetActions.fetchOutliers(store, dataset);
    viewActions.updateBaselinePredictions(store);
  }
  fetchVariables(): Promise<unknown> {
    const dataset = routeGetters.getRouteDataset(store);
    return datasetActions.fetchVariables(store, { dataset });
  }
  fetchData(): Promise<unknown> {
    return viewActions.updatePrediction(store);
  }
  fetchVariableSummaries(): Promise<unknown> {
    return viewActions.updatePredictionTrainingSummaries(store);
  }
  fetchMapBaseline(): Promise<void> {
    return viewActions.updateBaselinePredictions(store) as Promise<void>;
  }
  fetchMapDrillDown(filter: Filter): Promise<void[]> {
    return viewActions.updatePredictionAreaOfInterest(store, filter);
  }
}

export class LabelViewState implements BaseState {
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent) {
    console.error("timeseries is not supported in label view");
  }
  name = ExplorerStateNames.LABEL_VIEW;
  getVariables(): Variable[] {
    const labelScoreName =
      LOW_SHOT_SCORE_COLUMN_PREFIX + routeGetters.getRouteLabel(store);
    return datasetGetters.getVariables(store).filter((v) => {
      return (
        v.distilRole !== DISTIL_ROLES.SystemData || v.key !== labelScoreName
      );
    });
  }
  getSecondaryVariables(): Variable[] {
    const labelScoreName =
      LOW_SHOT_SCORE_COLUMN_PREFIX + routeGetters.getRouteLabel(store);
    return datasetGetters.getVariables(store).filter((v) => {
      return v.key === labelScoreName;
    });
  }
  getData(): TableRow[] {
    return datasetGetters.getIncludedTableDataItems(store) ?? [];
  }
  getBaseVariableSummaries(): VariableSummary[] {
    const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
      store
    );
    const dataset = routeGetters.getRouteDataset(store);
    return getAllVariablesSummaries(
      this.getVariables(),
      summaryDictionary,
      dataset
    );
  }
  getSecondaryVariableSummaries(): VariableSummary[] {
    const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
      store
    );
    const dataset = routeGetters.getRouteDataset(store);
    return getAllVariablesSummaries(
      this.getSecondaryVariables(),
      summaryDictionary,
      dataset
    );
  }
  getAllVariableSummaries(): VariableSummary[] {
    return this.getSecondaryVariableSummaries().concat(
      this.getBaseVariableSummaries()
    );
  }
  getTargetVariable(): Variable {
    const labelName = routeGetters.getRouteLabel(store);
    return datasetGetters.getVariables(store).find((v) => {
      return v.key === labelName;
    });
  }
  getMapBaseline(): TableRow[] {
    const bItems = datasetGetters.getBaselineIncludeTableDataItems(store) ?? [];
    return bItems.sort((a, b) => {
      return a.d3mIndex - b.d3mIndex;
    });
  }
  getMapDrillDownBaseline(): TableRow[] {
    return datasetGetters.getAreaOfInterestIncludeInnerItems(store);
  }
  getMapDrillDownFiltered(): TableRow[] {
    return datasetGetters.getAreaOfInterestIncludeOuterItems(store);
  }
  getLexBarVariables(): Variable[] {
    return datasetGetters.getAllVariables(store);
  }
  getFields(): Dictionary<TableColumn> {
    return datasetGetters.getIncludedTableDataFields(store);
  }
  async init(): Promise<void> {
    await this.fetchData();
    await this.fetchMapBaseline();
    return;
  }
  fetchVariables(): Promise<unknown> {
    const dataset = routeGetters.getRouteDataset(store);
    return datasetActions.fetchVariables(store, {
      dataset,
    });
  }
  async fetchData(): Promise<unknown> {
    await this.fetchVariables();
    await viewActions.updateLabelData(store);
    return;
  }
  fetchVariableSummaries(): Promise<unknown> {
    const dataset = routeGetters.getRouteDataset(store);
    const highlights = routeGetters.getDecodedHighlights(store);
    const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
      store
    );
    const numRows = datasetGetters.getNumberOfRecords(store);
    filterParams.size = numRows;
    const dataMode = routeGetters.getDataMode(store);
    const variables = datasetGetters.getVariables(store);
    const varModes = routeGetters.getDecodedVarModes(store);
    filterParams.variables = variables.map((v) => v.key);
    return datasetActions.fetchIncludedVariableSummaries(store, {
      dataset,
      variables,
      filterParams,
      highlights,
      dataMode,
      varModes,
    });
  }
  fetchMapBaseline(): Promise<void> {
    return viewActions.updateHighlight(store);
  }
  fetchMapDrillDown(filter: Filter): Promise<void[]> {
    return viewActions.updateAreaOfInterest(store, filter);
  }
}
