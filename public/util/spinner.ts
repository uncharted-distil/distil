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

import "../styles/spinner.css";

export function spinnerHTML(): string {
  return [
    '<div class="bounce1"></div>',
    '<div class="bounce2"></div>',
    '<div class="bounce3"></div>',
  ].join("");
}

export function circleSpinnerHTML(): string {
  return '<div class="circle-spinner"></div>';
}
