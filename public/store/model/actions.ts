import _ from "lodash";
import axios from "axios";
import { ActionContext } from "vuex";
import { ModelState, Model } from "./index";
import { mutations } from "./module";
import { DistilState } from "../store";

export type ModelContext = ActionContext<ModelState, DistilState>;

export const actions = {
  // searches model descriptions and column names for supplied terms and writes
  // the results into the store
  async searchModels(context: ModelContext, terms: string): Promise<void> {
    const params = !_.isEmpty(terms) ? `?search=${terms}` : "";
    try {
      const response = await axios.get<Model[]>(`/distil/models${params}`);
      mutations.setFilteredModels(context, response.data);
    } catch (error) {
      console.error(error);
      mutations.setFilteredModels(context, []);
    }
  },

  // fetches the list of saved models and writes the results into the store
  async fetchModels(context: ModelContext): Promise<void> {
    try {
      const response = await axios.get<Model[]>("/distil/models");
      mutations.setModels(context, response.data);
    } catch (error) {
      console.error(error);
      mutations.setModels(context, []);
    }
  }
};
