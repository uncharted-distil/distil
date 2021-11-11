import {
  datasetActions,
  datasetGetters,
  datasetMutations,
  requestActions,
} from "../../../store";

import store from "../../../store/store";
import { D3M_INDEX_FIELD, VariableSummary } from "../../../store/dataset";
import { getters as routeGetters } from "../../../store/route/module";
import { DataExplorerRef, DataView } from "../../componentTypes";
import {
  emptyFilterParamsObject,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../../filters";
import { cloneFilters } from "../../highlights";
import { bulkRowSelectionUpdate, clearRowSelection } from "../../row";
import { EI, EventList } from "../../events";
import { CATEGORICAL_TYPE, DISTIL_ROLES } from "../../types";
import {
  addOrderBy,
  cloneDatasetUpdateRoute,
  downloadFile,
  hasRole,
  LowShotLabels,
  LOW_SHOT_RANK_COLUMN_PREFIX,
  LOW_SHOT_SCORE_COLUMN_PREFIX,
} from "../../data";
import { LABEL_FEATURE_INSTANCE } from "../../../store/route";
import router from "../../../router/router";
import { ActionNames, ACTION_MAP, ExplorerStateNames } from "..";
import { NavigationGuardNext } from "vue-router";

/**
 * LABEL_COMPUTES contains all of the computes for the label state in the data explorer
 **/
export const LABEL_COMPUTES = {
  /**
   * labelScoreName represents the back end key for the label variable
   */
  labelScoreName(): string {
    const self = (this as unknown) as DataExplorerRef;
    return LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
  },
  /**
   * When DISTIL_ROLES.Label exists within the variable list then this dataset has been used in the label
   * workflow before
   */
  hasLabelRole(): boolean | null {
    const self = (this as unknown) as DataExplorerRef;
    return self?.variables.some((v) => hasRole(v, DISTIL_ROLES.Label));
  },
  /**
   * options is displayed to the user when selecting a pre-existing variable to annotate
   */
  options(): { value: string; text: string }[] {
    const self = (this as unknown) as DataExplorerRef;
    return self?.variables
      .filter((v) => {
        return hasRole(v, DISTIL_ROLES.Label);
      })
      .map((v) => {
        return { value: v.colName, text: v.colName };
      });
  },
  /**
   * modal title for label creation / selection
   */
  labelModalTitle(): string {
    const self = (this as unknown) as DataExplorerRef;
    return self?.isClone ? "Select Label Feature" : "Label Creation";
  },
  /**
   * the label summary that is being annotated
   */
  labelSummary(): VariableSummary {
    const self = (this as unknown) as DataExplorerRef;
    const label = routeGetters.getRouteLabel(store);
    return self?.summaries.find((s) => {
      return s.key === label;
    });
  },
};

export const LABEL_METHODS = {
  /**
   * updateTask sets label variable as the target
   */
  async updateTask(): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    const taskResponse = await datasetActions.fetchTask(store, {
      dataset: self.dataset,
      targetName: self.labelName,
      variableNames: self.variables.map((v) => v.key),
    });
    const training = routeGetters.getDecodedTrainingVariableNames(store);
    const check = training.length;
    const trainingMap = new Map(
      training.map((t) => {
        return [t, true];
      })
    );
    self.variables.forEach((variable) => {
      if (!trainingMap.has(variable.key)) {
        training.push(variable.key);
      }
    });
    if (check === training.length) {
      return;
    }
    self.updateRoute({
      task: taskResponse.data.task.join(","),
      training: training.join(","),
      label: self.labelName,
    });
    return;
  },
  /**
   * onLabelSubmit is called when the user has created a new label to annotate
   * this starts the process of integrating the new variable into the dataset on the backend
   */
  async onLabelSubmit(bvModalEvt): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    if (!self.labelName) {
      bvModalEvt.preventDefault();
      self.labelNameState = false;
      return;
    }
    self.labelNameState = true;

    if (
      self.variables.some((v) => {
        return v.colName === self.labelName;
      })
    ) {
      self.updateRoute({
        label: self.labelName,
      });
      self.setBusyState(true, "Initializing Label View");
      await self.changeStatesByName(ExplorerStateNames.LABEL_VIEW);
      self.setBusyState(false);
      return;
    }
    self.setBusyState(true, "Cloning Dataset");
    const entry = await cloneDatasetUpdateRoute();
    // failed to clone
    if (entry === null) {
      return;
    }
    self.$router.push(entry).catch((err) => console.warn(err));
    self.setBusyState(true, "Adding New Field");
    // add new field
    await datasetActions.addField<string>(store, {
      dataset: self.dataset,
      name: self.labelName,
      fieldType: CATEGORICAL_TYPE,
      defaultValue: LowShotLabels.unlabeled,
      displayName: self.labelName,
      isLabel: true,
    });
    self.setBusyState(true, "Fetching Data");
    // fetch new dataset with the newly added field
    await self.changeStatesByName(ExplorerStateNames.LABEL_VIEW);
    self.setBusyState(false, "Fetching Data");
    // update task based on the current training data
    self.updateTask();
  },

  /**
   * used by b-model in DataExplorer when the user confirms navigation away from a cloned dataset
   */
  onConfirmRouteSave(nextRoute: NavigationGuardNext | null): void {
    const self = (this as unknown) as DataExplorerRef;

    // delete dataset
    const terms = routeGetters.getRouteTerms(store);
    datasetActions.deleteDataset(store, {
      dataset: self.dataset,
      terms: terms,
    });

    // switch route back to search after delete
    if (nextRoute) {
      nextRoute();
    } else {
      self.$router.push("/search");
    }
  },

  /**
   * used by b-model in DataExplorer when the user declines navigation away from a cloned dataset
   */
  onCancelRouteSave(nextRoute: NavigationGuardNext | null): void {
    if (nextRoute) {
      nextRoute(false);
    }
  },

  /**
   * switchToLabelState displays the label modal
   */
  switchToLabelState(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.$bvModal.show(self.labelModalId);
  },
  /**
   * onLabelSaveClick displays the save dataset modal
   */
  onLabelSaveClick(): void {
    const self = (this as unknown) as DataExplorerRef;
    self.$bvModal.show("save-dataset-modal");
  },
};

export const LABEL_EVENT_HANDLERS = {
  /**
   * onAnnotationChanged is called when the user is annotating rows of the data as positive or negative
   * this requires a refetch of data and variable summaries
   */
  [EventList.LABEL.ANNOTATION_EVENT]: async function (
    label: LowShotLabels
  ): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    const rowSelection = routeGetters.getDecodedRowSelection(store);
    const innerData = new Map<number, unknown>();
    const updateData = rowSelection.d3mIndices.map((i) => {
      innerData.set(i, { LowShotLabel: label });
      return {
        index: i.toString(),
        name: self.labelName,
        value: label,
      };
    });
    if (!updateData.length) {
      return;
    }
    const dataset = routeGetters.getRouteDataset(store);
    datasetMutations.updateAreaOfInterestIncludeInner(store, innerData);
    datasetActions.updateDataset(store, {
      dataset: dataset,
      updateData,
    });
    clearRowSelection(self.$router);
    self.updateRoute({
      annotationHasChanged: true,
    });
    await self.state.fetchData();
    if (self.isRemoteSensing) {
      self.state.fetchMapBaseline();
    }
    return;
  },
  /**
   * onToolSelection is called after a map tool selection event
   * this selects all rows inside the quad
   */
  [EventList.MAP.SELECTION_TOOL_EVENT]: async function (
    selection: EI.MAP.SelectionHighlight
  ): Promise<void> {
    const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
      store
    );
    filterParams.size = datasetGetters.getIncludedTableDataNumRows(store);
    // fetch data selected by map tool
    const resp = await datasetActions.fetchTableData(store, {
      dataset: selection.dataset,
      highlights: [selection],
      filterParams: filterParams,
      dataMode: null,
      include: true,
    });
    // find d3mIndex
    const labelIndex = resp.columns.findIndex((c) => {
      return c.key === D3M_INDEX_FIELD;
    });
    // if -1 then something failed
    if (labelIndex === -1) {
      return;
    }
    // map the values
    const indices = resp.values.map((v) => {
      return v[labelIndex].value.toString();
    });
    // update row selection
    const rowSelection = routeGetters.getDecodedRowSelection(store);
    bulkRowSelectionUpdate(router, selection.context, rowSelection, indices);
  },
  /**
   * onSelectAll selects all the items currently on the page
   */
  [EventList.LABEL.SELECT_ALL_EVENT]: function (): void {
    const self = (this as unknown) as DataExplorerRef;
    const dataView = (self.$refs.dataView as unknown) as DataView;
    dataView.selectAll();
  },
  /**
   * open-save-modal-event handler opens the modal that allows the user to save the label dataset
   */
  [EventList.LABEL.OPEN_SAVE_MODAL_EVENT]: function (): void {
    const self = (this as unknown) as DataExplorerRef;
    self.$bvModal.show("save-dataset-modal");
  },
  /**
   * onSaveDataset calls the backend and saves the dataset
   * this removes the clone property for a dataset so if the user tries to label they
   * will have to create a new label
   */
  [EventList.LABEL.SAVE_EVENT]: async function (
    saveName: string,
    retainUnlabeled: boolean
  ): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    self.isBusy = true;
    const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
    const labelRankName = LOW_SHOT_RANK_COLUMN_PREFIX + self.labelName;
    const highlightsClear = [
      {
        context: LABEL_FEATURE_INSTANCE,
        dataset: self.dataset,
        key: self.labelName,
        value: LowShotLabels.unlabeled,
      },
    ]; // exclude unlabeled from data export
    const highlights = retainUnlabeled ? null : highlightsClear;
    let filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
      store
    );
    filterParams = cloneFilters(filterParams);
    if (
      self.allVariables.some((v) => {
        return v.key === labelScoreName;
      })
    ) {
      // delete confidence variable when saving
      await datasetActions.deleteVariable(store, {
        dataset: self.dataset,
        key: labelScoreName,
      });
      await datasetActions.deleteVariable(store, {
        dataset: self.dataset,
        key: labelRankName,
      });
    }
    // clear the unlabeled values when saving
    if (retainUnlabeled) {
      await datasetActions.clearVariable(store, {
        dataset: self.dataset,
        key: self.labelName,
        highlights: highlightsClear,
        filterParams: filterParams,
      });
    }
    const dataMode = routeGetters.getDataMode(store);
    await datasetActions.saveDataset(store, {
      dataset: self.dataset,
      datasetNewName: saveName,
      filterParams,
      highlights,
      include: false,
      mode: INCLUDE_FILTER,
      dataMode,
    });
    self.isBusy = false;
    // CHANGE TO SELECT VIEW AFTER DS IS SAVED IN LABEL VIEW
    self.changeStatesByName(ExplorerStateNames.SELECT_VIEW);
    return;
  },
  /**
   * onExport is called when the user wants to download the newly annotated dataset to csv
   */
  [EventList.LABEL.EXPORT_EVENT]: async function (): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    const highlights = [
      {
        context: LABEL_FEATURE_INSTANCE,
        dataset: self.dataset,
        key: self.labelName,
        value: LowShotLabels.unlabeled,
      },
    ]; // exclude unlabeled from data export
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
    downloadFile(file, self.dataset, ".csv");
    return;
  },
  /**
   * onSearchSimilar calls the backend to start the image query process
   * which ranks unlabelled images based on their similarities to the positive samples
   */
  [EventList.LABEL.APPLY_EVENT]: async function (): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    self.setBusyState(true, "Searching for Similar Images");
    const res = (await requestActions.createQueryRequest(store, {
      datasetId: self.dataset,
      target: self.labelName,
      filters: emptyFilterParamsObject(),
    })) as { success: boolean; error: string };
    if (!res.success) {
      self.$bvToast.toast(res.error, {
        title: "Error",
        autoHideDelay: 5000,
        appendToast: true,
        variant: "danger",
        toaster: "b-toaster-bottom-right",
      });
    }
    const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
    addOrderBy(labelScoreName);
    self.isBusy = false;
    await self.state.fetchData();
    self.state.fetchMapBaseline();
    self.updateRoute({
      annotationHasChanged: false,
    });
    const outcome = ACTION_MAP.get(ActionNames.OUTCOME_VARIABLES);
    const open = routeGetters.getToggledActions(store).some((a) => {
      return a === outcome.paneId;
    });
    // open the outcome variable pane to display the new confidence and ranking
    if (!open) {
      self.toggleAction(ActionNames.OUTCOME_VARIABLES);
    }
    self.setBusyState(false);
  },
};
