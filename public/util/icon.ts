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

import { VueConstructor } from "vue";
import IconBase from "../components/icons/IconBase.vue";
import { ThisTypedComponentOptionsWithArrayProps } from "vue/types/options";

export interface IconBaseProps {
  title?: string;
  width?: number | string;
  height?: number | string;
  iconColor?: string;
}

// render provided vue icon component and return it's html element
export function createIcon(
  IconComponentConstructor: VueConstructor,
  props?: IconBaseProps
) {
  const iconBase = new IconBase({
    propsData: props,
  });
  const icon = new IconComponentConstructor();
  iconBase.$slots.default = [iconBase.$createElement("icon-placeholder")];
  iconBase.$mount();
  icon.$mount(iconBase.$el.querySelector("icon-placeholder"));
  return iconBase.$el;
}
