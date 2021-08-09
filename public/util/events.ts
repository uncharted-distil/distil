/* eslint-disable @typescript-eslint/no-namespace */
import { TableRow, Variable } from "../store/dataset/index";
/**
 * ALL EVENT RELATED CODE SHOULD BE HERE
 */

// holds all event names (this will help keep things consistent)
export class EventList {
  /********BASIC EVENTS*************/
  static readonly BASIC = {
    // close event use when something is closing
    CLOSE_EVENT: "close",
    // something has been removed
    REMOVE_EVENT: "removed",
    // input event
    INPUT_EVENT: "input",
    // click event
    CLICK_EVENT: "click",
    // shift click event
    SHIFT_CLICK_EVENT: "shift-click",
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
  static readonly VARIABLES = {
    // fetch variable rankings
    FETCH_RANK_EVENT: "fetch-variable-rankings",
    // occurs when a variable is removed or added to the training set
    VAR_SET_CHANGE_EVENT: "var-change",
    // occurs when a group of variables is removed or added to the training set
    VAR_SET_GROUP_CHANGE_EVENT: "group-change",
  };
  static readonly SUMMARIES = {
    FETCH_SUMMARIES_EVENT: "fetch-summaries",
  };
  /*********TABLE EVENTS*************/
  static readonly TABLE = {
    // table column was click
    COLUMN_CLICKED_EVENT: "col-clicked",
    // fetch timeseries data
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
    // remove pair from join
    REMOVE_EVENT: "remove-from-join",
  };
  /***********LEXBAR EVENTS***********/
  static readonly LEXBAR = {
    // lex query has changed, therefore change filters/highlights
    QUERY_CHANGE_EVENT: "lex-query",
  };
  static readonly MODEL = {
    // create model based on solutionRequestMsg
    CREATE_EVENT: "create-model",
    // save model event typically happens on result screen
    SAVE_EVENT: "save",
    // delete the model
    DELETE_EVENT: "model-delete",
    // apply model
    APPLY_EVENT: "model-apply",
  };
  static readonly EXPLORER = {
    NAV_EVENT: "nav-event",
    SWITCH_TO_LABELING_EVENT: "label",
  };
}

// contains dataset name, target name and a list of variables
export interface GroupChangeParams {
  dataset: string;
  targetName: string;
  variableNames: string[];
}
/*********EVENT INTERFACES*************/
export declare namespace EI {
  /**
   * MAP INTERFACES
   */
  namespace MAP {
    interface TileClickData {
      bounds: number[][];
      key: string;
      displayName: string;
      type: string;
      callback: (inner: TableRow[], outer: TableRow[]) => void;
    }
    interface SelectionHighlight {
      context: string;
      dataset: string;
      key: string;
      value: {
        minX: number;
        maxX: number;
        minY: number;
        maxY: number;
      };
    }
  }
  /**
   * TIMESERIES INTERFACES
   */
  namespace TIMESERIES {
    // use this interface for timeseries fetch events
    interface FetchTimeseriesEvent {
      variables: Variable[];
      uniqueTrail: string;
      timeseriesIds: TableRow[];
    }
  }
  namespace RESULT {
    interface SaveInfo {
      solutionId: string;
      fittedSolution: string;
      name: string;
      description: string;
    }
  }
}
