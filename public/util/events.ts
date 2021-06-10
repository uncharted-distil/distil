import { Variable, TableRow } from "../store/dataset/index";
/**
 * ALL EVENT RELATED CODE SHOULD BE HERE
 */

// holds all event names (this will help keep things consistent)
export class EventList {
  // occurs when a variable is removed or added to the training set
  static readonly VAR_SET_CHANGE_EVENT = "var-change";
  // occurs when a group of variables is removed or added to the training set
  static readonly VAR_SET_GROUP_CHANGE_EVENT = "group-change";
  /********BASIC EVENTS*************/
  static readonly BASIC = {
    // close event use when something is closing
    CLOSE_EVENT: "close",
    // something has been removed
    REMOVE_EVENT: "removed",
  };
  /********UPLOAD EVENTS*************/
  static readonly UPLOAD = {
    // upload has begun
    START_EVENT: "uploadstart",
    // upload has finished
    FINISHED_EVENT: "uploadfinish",
  };

  /*********MAP EVENTS***************/
  static readonly MAP = {
    // map tile was clicked
    TILE_CLICKED_EVENT: "tile-clicked",
    // the selection tool is being used event
    SELECTION_TOOL_EVENT: "selection-tool-event",
    // changing the map type (basic map, satellite imagery map)
    MAP_TOGGLE_EVENT: "map-toggle",
    // changing tiles on map to cluster state on/off
    CLUSTERING_TOGGLE_EVENT: "clustering-toggle",
    // turn selection tool on/off
    SELECTION_TOOL_TOGGLE_EVENT: "selection-tool-toggle",
    // turn on/off baseline nodes on map
    BASELINE_TOGGLE_EVENT: "baseline-toggle",
  };
  /*********FACET EVENTS*************/
  static readonly FACETS = {
    // Range change applies for numerical, datetime, and timeseries facets
    RANGE_CHANGE_EVENT: "range-change",
    // mostly occurs in categorical type facets
    CLICK_EVENT: "facet-click",
    // categorical selection
    CATEGORICAL_CLICK_EVENT: "categorical-click",
    // numerical facet was clicked
    NUMERICAL_CLICK_EVENT: "numerical-click",
    // this event occurs when the variable facet is paged
    PAGE_EVENT: "page",
    // this event occurs when the variable facet is searched
    SEARCH_EVENT: "search",
  };
  /*********TABLE EVENTS*************/
  static readonly TABLE = {
    COLUMN_CLICKED_EVENT: "col-clicked",
    FETCH_TIMESERIES_EVENT: "fetch-timeseries",
  };

  /*********DATASET EVENTS***********/
  static readonly DATASETS = {
    SAVE_EVENT: "save",
    DELETE_EVENT: "dataset-delete",
  };
  /**********JOIN EVENTS*************/
  static readonly JOIN = {
    // swap datasets event
    SWAP_EVENT: "swap-datasets",
    // join was successful
    SUCCESS_EVENT: "success",
    // join failed
    FAILURE_EVENT: "failure",
  };
}
// these interfaces should probably be within namespaces to be condusive with the EventList
// use this interface for timeseries fetch events
export interface FetchTimeseriesEvent {
  variables: Variable[];
  uniqueTrail: string;
  timeseriesIds: TableRow[];
}

// contains dataset name, target name and a list of variables
export interface GroupChangeParams {
  dataset: string;
  targetName: string;
  variableNames: string[];
}
