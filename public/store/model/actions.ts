/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

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
  },

  // deletes the specified model and then does a search using the terms to refresh the list
  async deleteModel(
    context: ModelContext,
    payload: { model: string; terms: string }
  ): Promise<void> {
    if (!payload.model) {
      return;
    }
    try {
      // delete dataset
      await axios.post(`/distil/delete-model/${payload.model}`);
      // update current list of models
      await actions.searchModels(context, payload.terms);
    } catch (err) {
      console.error(err);
    }
  },
};
