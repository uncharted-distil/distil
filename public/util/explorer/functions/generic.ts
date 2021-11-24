import _, { isEmpty } from "lodash";
import ExplorerConfig, {
  Action,
  ActionNames,
  ACTION_MAP,
  ExplorerStateNames,
  getConfigFromName,
  getStateFromName,
} from "..";
import { datasetActions } from "../../../store";
import {
  Highlight,
  RowSelection,
  TableColumn,
  TableRow,
  TimeSeries,
  Variable,
  VariableSummary,
} from "../../../store/dataset";
import { getters as datasetGetters } from "../../../store/dataset/module";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";
import { ActionColumnRef, DataExplorerRef } from "../../componentTypes";
import {
  downloadFile,
  hasRole,
  sortVariablesByImportance,
  totalAreaCoverage,
} from "../../data";
import { Dictionary } from "../../dict";
import { EI, EventList } from "../../events";
import { EXCLUDE_FILTER, Filter, INCLUDE_FILTER } from "../../filters";
import { clearHighlight } from "../../highlights";
import { overlayRouteEntry, RouteArgs } from "../../routes";
import { clearRowSelection, getNumIncludedRows } from "../../row";
import { spinnerHTML } from "../../spinner";
import { BaseState } from "../../state/AppStateWrapper";
import {
  DISTIL_ROLES,
  isGeoLocatedType,
  isImageType,
  isMultibandImageType,
  META_TYPES,
} from "../../types";
import {
  filterViews,
  GEO_VIEW,
  GRAPH_VIEW,
  IMAGE_VIEW,
  TABLE_VIEW,
} from "../../view";

export const GENERIC_METHODS = {
  /**
   * When the user request to fetch a different size of data.
   */
  onDataSizeSubmit(dataSize: number): void {
    const self = (this as unknown) as DataExplorerRef;
    self.updateRoute({ dataSize });
    self.state.fetchData();
  },
  /**
   * changes the active action which changes the active panel
   */
  onSetActive(actionName: string): void {
    const self = (this as unknown) as DataExplorerRef;
    if (actionName === self.config.currentPane) return;
    let activePane = "";
    if (actionName !== "") {
      activePane = self.config.actionList.find((a) => a.name === actionName)
        .paneId;
    }
    self.config.currentPane = activePane;
  },
  /**
   * This opens and closes an action panel
   */
  toggleAction(actionName: ActionNames): void {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return;
    }
    const actionColumn = (self.$refs[
      "action-column"
    ] as unknown) as ActionColumnRef;
    actionColumn.toggle(ACTION_MAP.get(actionName).paneId);
  },
  /**
   * changeStatesByName correctly changes the state by enum
   * one should use this function to change the state as it resets and inits the states
   */
  async changeStatesByName(state: ExplorerStateNames) {
    const self = (this as unknown) as DataExplorerRef;
    // reset state
    self.state.resetState();
    // get the new state object
    self.setState(getStateFromName(state));
    // init this is the basic fetches needed to get the information for the state
    await self.state.init();
    // reset config state
    self.config.resetConfig(self);
    // set the config used for action bar, could be used for other configs
    self.setConfig(getConfigFromName(state));
  },
  /**
   * setBusyState changes the data view from spinner if true to data view components
   * busyState can be empty it defaults to 'Busy'
   */
  setBusyState(isBusy: boolean, busyState?: string): void {
    const self = (this as unknown) as DataExplorerRef;
    self.isBusy = isBusy;
    self.busyState = busyState ? busyState : "Busy";
  },
  /**
   * onTileClick is the callback function for when a tile is clicked
   */
  async onTileClick(data: EI.MAP.TileClickData): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    // filter for area of interests
    const filter: Filter = {
      displayName: data.displayName,
      key: data.key,
      maxX: data.bounds[1][1],
      maxY: data.bounds[0][0],
      minX: data.bounds[0][1],
      minY: data.bounds[1][0],
      mode: INCLUDE_FILTER,
      type: data.type,
      set: "",
    };
    // fetch area of interests
    self.state.fetchMapDrillDown(filter);
  },
  /**
   * setState changes the current state of the DataExplorer
   */
  async setState(state: BaseState): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    self.state = state;
    if (self.explorerRouteState !== state.name) {
      self.updateRoute({
        dataExplorerState: state.name,
        toggledActions: "[]",
      } as RouteArgs);
    }
  },
  /**
   * setConfig changes the current config for the DataExplorer. Which controls the action bar
   * each DataExplorer state has their own config
   */
  setConfig(config: ExplorerConfig): void {
    const self = (this as unknown) as DataExplorerRef;
    self.config = config;
    const toggledMap = new Map(
      routeGetters.getToggledActions(store).map((t) => {
        return [t, true];
      })
    );
    self.bindEventHandlers(self.config.eventHandlers);
    // the switch to the new config will trigger a render of new elements
    // if the defaultActions is one of the new elements it will not exist in the dom yet
    // so we toggle the default actions after the next DOM cycle
    self.$nextTick(() => {
      self.config.defaultAction.forEach((actionName) => {
        const action = ACTION_MAP.get(actionName);
        if (!toggledMap.has(action.paneId)) {
          self.toggleAction(actionName);
        }
      });
    });
  },
  /**
   * setIncludedActive sets the table/mosaic/map view to include in the Select view state
   */
  setIncludedActive(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.include = true;
  },
  /**
   * setIncludedActive sets the table/mosaic/map view to exclude in the Select view state
   */
  setExcludedActive(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.include = false;
  },
  /**
   * updateRoute is a general purpose function it applies any routeArgs supplied to the route
   */
  updateRoute(args: RouteArgs): void {
    const self = (this as unknown) as DataExplorerRef;
    const entry = overlayRouteEntry(self.$route, args);
    self.$router.push(entry).catch((err) => console.warn(err));
  },
  /**
   * preSelectTopVariables in order to fetch data we need training variables to display
   this function is used to create pseudo training variables in order to explore the data
   before selecting training features.
   */
  preSelectTopVariables(number = 5): void {
    const self = (this as unknown) as DataExplorerRef;
    // if explore is already filled let's skip
    if (!isEmpty(self.explore)) return;

    // get the top 5 variables
    const top5Variables = [...self.variables]
      .slice(0, number)
      .map((variable) => variable.key)
      .join(",");

    // Update the route with the top 5 variable as training
    self.updateRoute({ explore: top5Variables });
  },
  /**
   * onTabClick this function handles fetching all the data if the geoplot is clicked
   * the geoplot is a strange view which requires all the data at once for the visualization
   * the other views correctly use pagination and a data size limiter
   */
  onTabClick(view: string): void {
    const self = (this as unknown) as DataExplorerRef;
    // if the data size is not set to the max set it then fetch data
    if (view === GEO_VIEW) {
      self.dataLoading = true;
      if (self.numRows !== self.totalNumRows) {
        self.updateRoute({ dataSize: self.totalNumRows });
        self.state.fetchData();
      }
    }
  },
  /**
   * fetchTimeseries is the callback for the fetching timeseries event
   * which will call the back end and retrieve data related to timeseries
   * this is generally called when pagination (or a table sort) occurs and we need to fetch the page of timeseries
   */
  fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent): void {
    const self = (this as unknown) as DataExplorerRef;
    self.state.fetchTimeseries(args);
  },
  /**
   * fetchSummaries used to fetch the summary information for any state
   */
  fetchSummaries(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.state.fetchVariableSummaries();
  },
  /**
   * onMapFinishedLoading callback for when the map has finished loading visual data
   */
  onMapFinishedLoading(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.dataLoading = false;
  },
  /**
   * resetHighlightsOrRow clears any highlights or any row selection
   */
  resetHighlightsOrRow(): void {
    const self = (this as unknown) as DataExplorerRef;
    if (self.isFilteringHighlights) {
      clearHighlight(self.$router);
    } else {
      clearRowSelection(self.$router);
    }
  },
  /**
   * is the user able to navigate away from a cloned dataset
   */
  async isCurrentDatasetSaved(): Promise<boolean> {
    const self = (this as unknown) as DataExplorerRef;

    const datasetString = routeGetters.getRouteDataset(store);
    await datasetActions.fetchDataset(store, { dataset: datasetString });

    const datasets = datasetGetters.getDatasets(store);
    const dataset = datasets.find((d) => d.id === self.dataset);

    return dataset && dataset.immutable === false;
  },
};

export const GENERIC_COMPUTES = {
  /**
   * items returns the table/mosaic/map data during any state
   */
  items(): TableRow[] {
    const self = (this as unknown) as DataExplorerRef;
    return self.state.getData(self.include);
  },
  /**
   * activePane returns the current action pane that is open
   */
  activePane(): string {
    const self = (this as unknown) as DataExplorerRef;
    return self.config.currentPane;
  },
  /**
   * activeActions returns the actions that exist within the config state
   */
  activeActions(): Action[] {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return [];
    }
    return self.availableActions.map((action) => {
      const count = self.variablesPerActions[action.paneId]?.length;
      return count ? { ...action, count } : action;
    });
  },

  /**
   * Variables displayed on the Facet Panel
   */
  activeVariables(): Variable[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.variablesPerActions[self?.config.currentPane] ?? [];
  },
  /**
   * activeViews returns which views should exist table/mosaic/map
   */
  activeViews(): string[] {
    const self = (this as unknown) as DataExplorerRef;
    const vars = self?.variables ?? [];
    return filterViews(vars);
  },

  /**
   * All variables, only used for lex as we need to parse the hidden variables from groupings
   */
  allVariables(): Variable[] {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return [];
    }
    const variables = [...self.state.getLexBarVariables()];
    return sortVariablesByImportance(variables);
  },

  /**
   * Actions available based on the variables meta types
   */
  availableActions(): Action[] {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return [];
    }
    // Remove the inactive MetaTypes
    return self.config.actionList.filter(
      (action) => !self.inactiveMetaTypes.includes(action.paneId)
    );
  },
  /**
   * the current target variable key can be undefined
   */
  targetName(): string | undefined {
    const self = (this as unknown) as DataExplorerRef;
    return self?.target?.key;
  },
  /**
   * returns true if one of the variables is a multibandimage type
   */
  isMultiBandImage(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return self?.allVariables.some((v) => {
      return isMultibandImageType(v.colType);
    });
  },
  /**
   * returns target variable type as string
   */
  targetType(): string {
    const self = (this as unknown) as DataExplorerRef;
    const target = self?.target;
    if (!target) {
      return null;
    }
    const variables = self.variables;
    return variables.find((v) => v.key === target.key)?.colType;
  },
  /**
   * returns the current action that is open in the pane
   */
  currentAction(): string {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return "";
    }
    return (
      self.config.currentPane &&
      self.config.actionList.find((a) => a.paneId === self.config.currentPane)
        .name
    );
  },
  /**
   * dataset name
   */
  dataset(): string {
    const self = (this as unknown) as DataExplorerRef;
    return self.state.dataset();
  },
  /**
   * returns the explore variables being displayed in table/mosaic/map
   */
  explore(): string[] {
    return routeGetters.getExploreVariables(store);
  },
  /**
   * returns the filter string which needs to be decoded
   */
  filters(): string {
    return routeGetters.getRouteFilters(store);
  },
  /**
   * returns true if data != null
   */
  hasData(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.hasData();
  },
  /**
   * task returns the task stored in route
   */
  task(): string {
    return routeGetters.getRouteTask(store) ?? "";
  },
  /**
   * returns true if the secondaryVariables from state is empty
   */
  hasNoSecondaryVariables(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return isEmpty(self?.secondaryVariables);
  },
  /**
   * returns true if the current active action's variables is empty
   */
  hasNoVariables(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return isEmpty(self?.activeVariables);
  },
  /**
   * returns true if timeseries exists within the route
   */
  isTimeseries(): boolean {
    return routeGetters.isTimeseries(store);
  },
  /**
   * returns all the highlights currently being applied
   */
  highlights(): Highlight[] {
    return _.cloneDeep(routeGetters.getDecodedHighlights(store));
  },
  /**
   * returns the all the items within the drill down area
   */
  drillDownBaseline(): TableRow[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getMapDrillDownBaseline(self?.include);
  },
  /**
   * returns all the items that are not filtered out by highlights / filters
   */
  drillDownFiltered(): TableRow[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getMapDrillDownFiltered(self?.include);
  },
  /**
   * returns the timeseries object which contains all the timeseries info
   */
  timeseries(): TimeSeries {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getTimeseries();
  },
  /**
   * returns the highlight string which needs to be decoded
   */
  routeHighlight(): string {
    return routeGetters.getRouteHighlight(store);
  },
  /**
   * returns the meta types used to categorize what actions should be displayed by variable types
   */
  inactiveMetaTypes(): string[] {
    const self = (this as unknown) as DataExplorerRef;
    // Go trough each meta type
    return self?.metaTypes.map((metaType) => {
      // test if some variables types...
      const typeNotInMetaTypes = !self.variablesTypes.some((t) =>
        // ...is in that meta type
        META_TYPES[metaType].includes(t)
      );
      if (typeNotInMetaTypes) return metaType;
    });
  },
  /**
   * data fields define what types are for each column
   * useful for identifying specific columns like images which need special handling
   */
  fields(): Dictionary<TableColumn> {
    const self = (this as unknown) as DataExplorerRef;
    return self.state.getFields(self.include);
  },
  /**
   * returns true if there is some sort of highlight
   */
  isFilteringHighlights(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return self.highlights && self.highlights.length > 0;
  },
  /**
   * returns true if there is some sort of rowSelection
   */
  isFilteringSelection(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return !!self.rowSelection;
  },
  /**
   * number of items that exist within the table data
   */
  numRows(): number {
    const self = (this as unknown) as DataExplorerRef;
    return self.items.length;
  },
  /**
   * returns the current rowSelection
   */
  rowSelection(): RowSelection {
    return routeGetters.getDecodedRowSelection(store);
  },
  /**
   * the number of selected rows
   */
  selectionNumRows(): number {
    const self = (this as unknown) as DataExplorerRef;
    return getNumIncludedRows(self?.rowSelection);
  },

  spinnerHTML,
  /**
   * the current target variable can be undefined
   */
  target(): Variable | undefined {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getTargetVariable();
  },
  /**
   * returns the total number of rows in the dataset (not the amount of data residing in the client currently)
   */
  totalNumRows(): number {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getTotalItems(self?.include);
  },
  /**
   * returns the variables for the current active state
   */
  variables(): Variable[] {
    const self = (this as unknown) as DataExplorerRef;
    const variables = self?.state
      .getVariables()
      .filter((v) => !hasRole(v, DISTIL_ROLES.Meta));
    return sortVariablesByImportance(variables);
  },
  /**
   * returns a map of all the variables given an action
   */
  variablesPerActions(): Record<string, Variable[]> {
    const self = (this as unknown) as DataExplorerRef;
    const variables = {};
    self?.availableActions.forEach((action) => {
      variables[action.paneId] = action.variables(self);
    });
    return variables;
  },
  /**
   * array of variable types as string
   */
  variablesTypes(): string[] {
    const self = (this as unknown) as DataExplorerRef;
    return [...new Set(self?.variables.map((v) => v.colType))] as string[];
  },
  /**
   * enables or disables coloring by facets
   */
  geoVarExists(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    const varSums = self?.summaries ?? [];
    return varSums.some((v) => {
      return v?.type && isGeoLocatedType(v.type);
    });
  },
  /**
   * checks for any type of image existing within the list of variables
   */
  imageVarExists(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    const varSums = self?.allVariables ?? [];
    return varSums.some((v) => {
      return isImageType(v.colType);
    });
  },
  /**
   * returns true if current view is the map
   */
  isGeoView(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return self?.viewComponent === "GeoPlot";
  },
  /**
   * returns current view component
   */
  viewComponent(): string {
    const self = (this as unknown) as DataExplorerRef;
    const viewType = self?.activeViews[self.activeView] as string;
    if (viewType === GEO_VIEW) return "GeoPlot";
    if (viewType === GRAPH_VIEW) return "SelectGraphView";
    if (viewType === IMAGE_VIEW) return "ImageMosaic";
    if (viewType === TABLE_VIEW) return "SelectDataTable";
    // Default is TABLE_VIEW
    return "SelectDataTable";
  },
  /**
   * used to enable certain UI components
   */
  isResultState(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return ExplorerStateNames.RESULT_VIEW === self.explorerRouteState;
  },
  /**
   * used to enable certain UI components
   */
  isSelectState(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return ExplorerStateNames.SELECT_VIEW === self.explorerRouteState;
  },
  /**
   * used to enable certain UI components
   */
  isPredictState(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return ExplorerStateNames.PREDICTION_VIEW === self.explorerRouteState;
  },
  /**
   * used to enable certain UI components
   */
  isLabelState(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    return ExplorerStateNames.LABEL_VIEW === self.explorerRouteState;
  },

  /**
   * baselineMap is used to maintain index order for faster buffer changes
   */
  baselineMap(): Dictionary<number> {
    const self = (this as unknown) as DataExplorerRef;
    const result = {};
    const base = self.baselineItems ?? [];
    base.forEach((item, i) => {
      result[item.d3mIndex] = i;
    });
    return result;
  },
  /**
   * used for map is the baseline
   */
  baselineItems(): TableRow[] {
    const self = (this as unknown) as DataExplorerRef;
    return self.state.getMapBaseline();
  },
  /**
   * returns all summaries
   */
  summaries(): VariableSummary[] {
    const self = (this as unknown) as DataExplorerRef;
    return self.state.getAllVariableSummaries(self.include) ?? [];
  },
  /**
   * available summaries, result summaries, prediction summaries
   */
  secondarySummaries(): VariableSummary[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getSecondaryVariableSummaries(self?.include);
  },
  /**
   * available variables, result variables, prediction variables
   */
  secondaryVariables(): Variable[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.state.getSecondaryVariables();
  },
  /**
   * returns the route state stored in the route params (useful for restoring state after browser navs)
   */
  explorerRouteState(): ExplorerStateNames {
    return routeGetters.getDataExplorerState(store);
  },
  /**
   * returns number of km^2 currently highlighted (used for the map view)
   */
  areaCoverage(): number {
    const self = (this as unknown) as DataExplorerRef;
    return totalAreaCoverage(self?.items, self?.variables);
  },
  /**
   * toggles right side variable pane
   */
  isOutcomeToggled(): boolean {
    const outcome = ACTION_MAP.get(ActionNames.OUTCOME_VARIABLES).paneId;
    return routeGetters.getToggledActions(store).some((a) => a === outcome);
  },
  /**
   * returns true if multiBandImage is in the route params
   */
  isRemoteSensing(): boolean {
    return routeGetters.isMultiBandImage(store);
  },
};

export const GENERIC_EVENT_HANDLERS = {
  /**
   * onExport is called when the user wants to download the newly annotated dataset to csv
   */
  [EventList.EXPLORER.EXPLORER_EXPORT]: async function (
    filename?: string
  ): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    const highlights = routeGetters.getDecodedHighlights(store);
    const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
      store
    );
    const dataMode = routeGetters.getDataMode(store);
    const file = await datasetActions.extractDataset(store, {
      dataset: self.dataset,
      filterParams,
      highlights,
      include: true,
      mode: EXCLUDE_FILTER,
      dataMode,
    });
    downloadFile(file, filename || self.dataset, ".csv");
    return;
  },
};
