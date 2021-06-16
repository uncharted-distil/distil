import {
  datasetGetters,
  requestGetters,
  resultGetters,
  viewActions,
} from "../../store";
import {
  D3M_INDEX_FIELD,
  TableColumn,
  TableRow,
  Variable,
  VariableSummary,
} from "../../store/dataset";
import { getters as routeGetters } from "../../store/route/module";
import store from "../../store/store";
import { getAllVariablesSummaries, getVariableSummariesByState } from "../data";
import { Dictionary } from "../dict";
import { Filter } from "../filters";
import { getSolutionById } from "../solutions";
import {
  getConfidenceSummary,
  getCorrectnessSummary,
  getRankingSummary,
  getResidualSummary,
  getSolutionResultSummary,
  resultSummariesToVariables,
} from "../summaries";

export interface BaseState {
  // gets basic variables
  getVariables(): Variable[];
  // basic data for tables and maps
  getData(include?: boolean): TableRow[];
  // base variable summaries is the standard variables in a view. For select it is the available variables,
  //for result it is the training variables
  getBaseVariableSummaries(include?: boolean): VariableSummary[];
  // this is the trainingVariables, result summaries, prediction summaries
  getSecondaryVariableSummaries(include?: boolean): VariableSummary[];
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
  fetchVariables(): Promise<unknown>;
  fetchData(): Promise<unknown>;
  fetchVariableSummaries(): Promise<unknown>;
  fetchMapBaseline(): Promise<void>;
  fetchMapDrillDown(filter: Filter): Promise<Array<void>>;
}

export class SelectViewState implements BaseState {
  // helper function for getting variable summary dictionary
  getSummaryDictionary(
    include?: boolean
  ): Dictionary<Dictionary<VariableSummary>> {
    return include
      ? datasetGetters.getIncludedVariableSummariesDictionary(store)
      : datasetGetters.getExcludedVariableSummariesDictionary(store);
  }
  getSecondaryVariableSummaries(include?: boolean): VariableSummary[] {
    const varDict = this.getSummaryDictionary(include);
    const variables = routeGetters.getTrainingVariables(store);
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
  // returns all variable summaries based on this.getVariables()
  getBaseVariableSummaries(include: boolean): VariableSummary[] {
    const varDict = this.getSummaryDictionary(include);
    const variables = routeGetters.getAvailableVariables(store);
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
    return viewActions.updateSelectVariables(store);
  }
  fetchMapBaseline(): Promise<void> {
    return viewActions.updateHighlight(store);
  }
}

export class ResultViewState implements BaseState {
  getSecondaryVariableSummaries(): VariableSummary[] {
    const currentSummaries = [];
    const solution = requestGetters.getActiveSolution(store);
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
    return resultGetters.getIncludedResultTableDataItems(store);
  }
  getBaseVariableSummaries(): VariableSummary[] {
    const summaryDictionary = resultGetters.getTrainingSummariesDictionary(
      store
    );
    const variables = this.getVariables();
    const trainingSummaries = getVariableSummariesByState(
      0,
      variables.length,
      variables,
      summaryDictionary,
      true
    );

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
    const solutionID = routeGetters.getRouteSolutionId(store);
    const solution = getSolutionById(
      requestGetters.getRelevantSolutions(store),
      solutionID
    );
    const resultVariables = resultSummariesToVariables(solution?.resultId);
    return datasetGetters.getAllVariables(store).concat(resultVariables);
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
