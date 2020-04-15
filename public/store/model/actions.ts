import _ from "lodash";
import axios from "axios";
import { ActionContext } from "vuex";
import { ModelState } from "./index";
import { mutations } from "./module";
import { DistilState } from "../store";

export type ModelContext = ActionContext<ModelState, DistilState>;

export const actions = {
  // searches model descriptions and column names for supplied terms
  async searchModels(context: ModelContext, terms: string): Promise<void> {
    const params = !_.isEmpty(terms) ? `?search=${terms}` : "";
    try {
      const response = await axios.get(`/distil/models${params}`);
      mutations.setModels(context, response.data.models);
    } catch (error) {
      console.error(error);
      mutations.setModels(context, []);
    }
  }
};
