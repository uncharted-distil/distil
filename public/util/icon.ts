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
