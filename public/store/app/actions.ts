import axios from "axios";
import { AppState, StatusPanelContentType } from "./index";
import { DistilState } from "../store";
import { ActionContext } from "vuex";
import { mutations } from "./module";
import { FilterParams } from "../../util/filters";
import { Feature, Activity, SubActivity } from "../../util/userEvents";
import { resultsModule } from "../results/module";

export type AppContext = ActionContext<AppState, DistilState>;

export const actions = {
  async saveModel(
    context: AppContext,
    args: {
      fittedSolutionId: string;
      modelName: string;
      modelDescription: string;
    }
  ) {
    try {
      await axios.post(`/distil/save/${args.fittedSolutionId}/true`, {
        modelName: args.modelName,
        modelDescription: args.modelDescription,
      });
      console.warn(`User saved model for ${args.fittedSolutionId}`);
    } catch (error) {
      // If there's a proxy involved (NGINX) we will end up getting a 502 on a successful export because
      // the server exits.  We need to explicitly check for the condition here so that we don't interpret
      // a success case as a failure.
      if (error.response && error.response.status !== 502) {
        return new Error(error.response.data);
      } else {
        // NOTE: request always fails because we exit on the server
        console.warn(`User saved model for ${args.fittedSolutionId}`);
      }
    }
  },

  async exportSolution(context: AppContext, args: { solutionId: string }) {
    try {
      await axios.get(`/distil/export/${args.solutionId}`);
      console.warn(`User exported solution ${args.solutionId}`);
    } catch (error) {
      // If there's a proxy involved (NGINX) we will end up getting a 502 on a successful export because
      // the server exits.  We need to explicitly check for the condition here so that we don't interpret
      // a success case as a failure.
      if (error.response && error.response.status !== 502) {
        return new Error(error.response.data);
      } else {
        // NOTE: request always fails because we exit on the server
        console.warn(`User exported solution ${args.solutionId}`);
      }
    }
  },

  exportProblem(
    context: AppContext,
    args: {
      dataset: string;
      target: string;
      filterParams: FilterParams;
      meaningful: string;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.target) {
      console.warn("`target` argument is missing");
      return null;
    }
    if (!args.filterParams) {
      console.warn("`filters` argument is missing");
      return null;
    }
    if (!args.meaningful) {
      console.warn("`meaningful` argument is missing");
      return null;
    }
    return axios.post(`/distil/discovery/${args.dataset}/${args.target}`, {
      filterParams: args.filterParams,
      meaningful: args.meaningful,
    });
  },

  async fetchConfig(context: AppContext) {
    try {
      const response = await axios.get(`/distil/config`);
      mutations.setVersionNumber(context, response.data.version);
      mutations.setHelpURL(context, response.data.help);
      mutations.setVersionTimestamp(context, response.data.timestamp);
      mutations.setProblemDataset(context, response.data.dataset);
      mutations.setProblemTarget(context, response.data.target);
      mutations.setProblemMetrics(context, response.data.metrics);
      mutations.setTA2VersionNumber(context, response.data.ta2version);
    } catch (err) {
      console.warn(err);
    }
  },

  openStatusPanelWithContentType(
    context: AppContext,
    contentType: StatusPanelContentType
  ) {
    mutations.openStatusPanel(context);
    mutations.setStatusPanelContentType(context, contentType);
  },

  closeStatusPanel(context: AppContext) {
    mutations.setStatusPanelContentType(context, undefined);
    mutations.closeStatusPanel(context);
  },

  logUserEvent(
    context: AppContext,
    args: {
      feature: Feature;
      activity: Activity;
      subActivity: SubActivity;
      details: any;
    }
  ) {
    return axios.post(`distil/event`, args);
  },
};
