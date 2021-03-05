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

import store from "../store/store";
import { getters as datasetGetters } from "../store/dataset/module";
import _ from "lodash";
import axios from "axios";

// Converts a file into a Base64 string.
export function getBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => {
      let encoded = reader.result.toString().replace(/^data:(.*,)?/, "");
      if (encoded.length % 4 > 0) {
        encoded += "=".repeat(4 - (encoded.length % 4));
      }
      resolve(encoded);
    };
    reader.onerror = (error) => reject(error);
  });
}

// Removes the extension from a filename
export function removeExtension(filename: string): string {
  const name = filename.split(".");
  if (name.length === 1) return name[0];
  return name.slice(0, -1).join(".");
}

// Given a potential dataset name, will compare against those already stored
// in system and return one that is unique by appending `_n` if necessary.
// We make the comparison case insensitive.
export function generateUniqueDatasetName(datasetName: string): string {
  const datasetNames = new Set(
    datasetGetters.getDatasets(store).map((d) => _.toLower(d.id))
  );
  let newName = datasetName;
  let count = 1;
  while (true) {
    if (!datasetNames.has(_.toLower(newName))) {
      return newName;
    }
    newName = `${datasetName}_${count}`;
    count++;
  }
}

// Lookup datasets already available in the $D3MOUTPUTDIR/augmented folder.
export async function getAvailableDatasets() {
  const response = await axios.get("/distil/available");
  return response?.data?.availableDatasets;
}
