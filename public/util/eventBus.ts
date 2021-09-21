import Vue from "vue";
/**
 * Creating Vue plugin for global event bus
 */
export default function EventBusPlugin(vue: typeof Vue): void {
  vue.prototype.$eventBus = new Vue();
}
/**
 * Vue type augmentation
 */
declare module "vue/types/vue" {
  interface Vue {
    $eventBus: Vue;
  }
}
