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
  // close event use when something is closing
  static readonly CLOSE_EVENT = "close";
  static readonly BASIC = {};
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
  };
  /*********TABLE EVENTS*************/
  static readonly FETCH_TIMESERIES_EVENT = "fetch-timeseries";
  /*********DATASET EVENTS***********/
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
