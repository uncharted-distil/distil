/* eslint-disable @typescript-eslint/no-namespace */
import { TableRow, Variable } from "../store/dataset/index";
import Vue from "vue";
/**
 * ALL EVENT RELATED CODE SHOULD BE HERE
 */

// holds all event names (this will help keep things consistent)
export namespace EventList {
  /********BASIC EVENTS*************/
  export enum BASIC {
    // close event use when something is closing
    CLOSE_EVENT = "close",
    // something has been removed
    REMOVE_EVENT = "removed",
    // input event
    INPUT_EVENT = "input",
    // click event
    CLICK_EVENT = "click",
    // shift click event
    SHIFT_CLICK_EVENT = "shift-click",
  }
  /********UPLOAD EVENTS*************/
  export enum UPLOAD {
    // upload has begun
    START_EVENT = "uploadstart",
    // upload has finished
    FINISHED_EVENT = "uploadfinish",
  }

  /*********MAP EVENTS***************/
  export enum MAP {
    // map tile was clicked
    TILE_CLICKED_EVENT = "tile-clicked",
    // the selection tool is being used event
    SELECTION_TOOL_EVENT = "selection-tool-event",
    // changing the map type (basic map, satellite imagery map)
    MAP_TOGGLE_EVENT = "map-toggle",
    // changing tiles on map to cluster state on/off
    CLUSTERING_TOGGLE_EVENT = "clustering-toggle",
    // turn selection tool on/off
    SELECTION_TOOL_TOGGLE_EVENT = "selection-tool-toggle",
    // turn on/off baseline nodes on map
    BASELINE_TOGGLE_EVENT = "baseline-toggle",
    // event fires when map has finished loading
    FINISHED_LOADING = "finished-loading",
  }
  /*********FACET EVENTS*************/
  export enum FACETS {
    // Range change applies for numerical, datetime, and timeseries facets
    RANGE_CHANGE_EVENT = "range-change",
    // mostly occurs in categorical type facets
    CLICK_EVENT = "facet-click",
    // categorical selection
    CATEGORICAL_CLICK_EVENT = "categorical-click",
    // numerical facet was clicked
    NUMERICAL_CLICK_EVENT = "numerical-click",
    // this event occurs when the variable facet is paged
    PAGE_EVENT = "page",
    // this event occurs when the variable facet is searched
    SEARCH_EVENT = "search",
  }
  export enum VARIABLES {
    // event is fired when outlier needs to be applied to ds
    APPLY_OUTLIER_EVENT = "apply-outlier-event",
    // event is fired when cluster needs to be applied to grouped variable
    APPLY_CLUSTER_EVENT = "apply-cluster-event",
    // fetch variable rankings
    FETCH_RANK_EVENT = "fetch-variable-rankings",
    // occurs when a variable is removed or added to the training set
    VAR_SET_CHANGE_EVENT = "var-change",
    // occurs when a group of variables is removed or added to the training set
    VAR_SET_GROUP_CHANGE_EVENT = "group-change",
    // occurs when a variable has their type changed
    TYPE_CHANGE = "type-change",
  }
  export enum SUMMARIES {
    FETCH_SUMMARIES_EVENT = "fetch-summaries",
    // fetch specific solution
    FETCH_SUMMARY_SOLUTION = "fetch-summary-solution",
    // fetch specific prediction request
    FETCH_SUMMARY_PREDICTION = "fetch-summary-prediction",
  }
  /*********TABLE EVENTS*************/
  export enum TABLE {
    // table column was click
    COLUMN_CLICKED_EVENT = "col-clicked",
    // fetch timeseries data
    FETCH_TIMESERIES_EVENT = "fetch-timeseries",
    // row selection has occured
    ROW_SELECTION_EVENT = "row-selection",
  }

  /*********DATASET EVENTS***********/
  export enum DATASETS {
    SAVE_EVENT = "save",
    DELETE_EVENT = "dataset-delete",
  }
  /**********JOIN EVENTS*************/
  export enum JOIN {
    // swap datasets event
    SWAP_EVENT = "swap-datasets",
    // join was successful
    SUCCESS_EVENT = "success",
    // join failed
    FAILURE_EVENT = "failure",
    // remove pair from join
    REMOVE_EVENT = "remove-from-join",
    // when variable type has been changed in the join view
    JOIN_TYPE_CHANGE = "join-type-change",
  }
  /***********LEXBAR EVENTS***********/
  export enum LEXBAR {
    // lex query has changed, therefore change filters/highlights
    QUERY_CHANGE_EVENT = "lex-query",
  }
  export enum MODEL {
    // create model based on solutionRequestMsg
    CREATE_EVENT = "create-model",
    // the instance where creating a model failed in the select view
    CREATION_FAILED = "create-model-failed",
    // the instance where creating a model succeeded in the select view
    CREATION_SUCCESS = "create-model-success",
    // save model event typically happens on result screen
    SAVE_EVENT = "save",
    // delete the model
    DELETE_EVENT = "model-delete",
    // apply model
    APPLY_EVENT = "model-apply",
  }
  export enum EXPLORER {
    NAV_EVENT = "nav-event",
    SWITCH_TO_LABELING_EVENT = "label",
    EXPLORER_EXPORT = "explorer-export",
  }
  export enum LABEL {
    ANNOTATION_EVENT = "annotation",
    SELECT_ALL_EVENT = "select-all",
    OPEN_SAVE_MODAL_EVENT = "label-open-save-modal-event",
    SAVE_EVENT = "label-save-dataset-event",
    EXPORT_EVENT = "label-export-dataset-event",
    APPLY_EVENT = "search-similar",
  }
  export enum IMAGE_DRILL_DOWN {
    RESET_IMAGE_EVENT = "image-drill-down-reset-image",
  }
  export enum IMAGES {
    // an event which changes the image drilldown image to one adjacent
    CYCLE_IMAGES = "cycle-images",
  }

  export enum HINTS {
    // select target is being hover over need to display hints across UI
    SELECT_TARGET = "select-target-hint",
    // select training is being hover over need to display hints across UI
    SELECT_TRAINING = "select-training-hint",
  }
}
// expose EventList to vue to be used in the html parts of the components
Vue.prototype.EventList = EventList;
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
  namespace IMAGES {
    enum Side {
      Left = -1,
      Right = 1,
    }
    interface CycleImage {
      side: Side;
      index: number;
    }
  }
}
