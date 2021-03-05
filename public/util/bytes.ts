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

export type Dictionary<T> = { [key: string]: T };

const SUFFIXES = {
  0: "B",
  1: "KB",
  2: "MB",
  3: "GB",
  4: "TB",
  5: "PB",
  6: "EB",
};

export function formatBytes(n: number): string {
  function formatRecursive(size: number, powerOfThousand: number): string {
    if (size > 1024) {
      return formatRecursive(size / 1024, powerOfThousand + 1);
    }
    return `${size.toFixed(2)}${SUFFIXES[powerOfThousand]}`;
  }
  return formatRecursive(n, 0);
}
