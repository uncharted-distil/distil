import { ActionNames, ACTION_MAP } from "..";
import { ActionColumnRef, DataExplorerRef } from "../../componentTypes";

export const GENERIC_METHODS = {
  /* When the user request to fetch a different size of data. */
  onDataSizeSubmit: (dataSize: number): void => {
    const self = this as DataExplorerRef;
    self.updateRoute({ dataSize });
    self.state.fetchData();
  },
  onSetActive: (actionName: string): void => {
    const self = this as DataExplorerRef;
    if (actionName === self.config.currentPane) return;
    let activePane = "";
    if (actionName !== "") {
      activePane = self.config.actionList.find((a) => a.name === actionName)
        .paneId;
    }
    self.config.currentPane = activePane;
  },
  toggleAction: (actionName: ActionNames): void => {
    const self = this as DataExplorerRef;
    if (!self) {
      return;
    }
    const actionColumn = (self.$refs[
      "action-column"
    ] as unknown) as ActionColumnRef;
    actionColumn.toggle(ACTION_MAP.get(actionName).paneId);
  },
};
