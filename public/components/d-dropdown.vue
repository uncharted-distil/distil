<template>
  <v-select
    id="type-change-dropdown"
    :value="value"
    append-to-body
    :calculate-position="withPopper"
    :disabled="disabled"
    :options="options"
    :clearable="false"
    :searchable="false"
  >
    <slot name="search" slot="search"></slot>
    <slot name="option" slot="option"></slot>
  </v-select>
</template>

<script lang="ts">
import Vue from "vue";
import { createPopper } from "@popperjs/core";

export default Vue.extend({
  name: "d-drop-down",
  props: {
    disabled: { type: Boolean as () => boolean, default: false },
    options: { type: Array, default: [] }, // this can really receive anything
    value: { type: String, default: "" },
  },
  methods: {
    withPopper(dropdownList, component, { width }) {
      dropdownList.style.width = width;
      const popper = createPopper(component.$refs.toggle, dropdownList, {
        modifiers: [
          {
            name: "offset",
            options: {
              offset: [0, -1],
            },
          },
          {
            name: "toggleClass",
            enabled: true,
            phase: "write",
            fn({ state }) {
              component.$el.classList.toggle(
                "drop-up",
                state.placement === "top"
              );
            },
          },
        ],
      });

      /**
       * To prevent memory leaks Popper needs to be destroyed.
       * If you return function, it will be called just before dropdown is removed from DOM.
       */
      return () => popper.destroy();
    },
  },
});
</script>
<style>
.vs--single.vs--open .vs__selected {
  position: relative !important;
  opacity: 0.4;
}
.vs__actions {
  display: -webkit-box;
  display: -ms-flexbox;
  display: flex;
  -webkit-box-align: center;
  -ms-flex-align: center;
  align-items: center;
  padding: 0px !important;
  margin-right: 5px;
}
.vs__dropdown-toggle {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  padding: 0px !important;
  background: none;
  border: 1px solid rgba(60, 60, 60, 0.26);
  border-radius: 4px;
  white-space: normal;
}
div.vs__actions > svg {
  width: 17px;
}
</style>
