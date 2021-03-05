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

import { Dictionary } from "./dict";

export function buildLookup(strs: any[]): Dictionary<boolean> {
  const lookup = {};
  strs.forEach((str) => {
    if (str) {
      lookup[str] = true;
      lookup[str.toLowerCase()] = true;
    } else {
      console.error(
        "Ignoring NULL string in look-up parameter list.  This should not happen."
      );
    }
  });
  return lookup;
}
