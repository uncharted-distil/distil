<template>
  <v-select
    append-to-body
    :value="value"
    :label="label"
    :calculate-position="withPopper"
    :disabled="disabled"
    :options="options"
    :clearable="false"
    :searchable="false"
    :style="style"
  >
    <!-- Pass on all named slots -->
    <slot v-for="slot in Object.keys($slots)" :name="slot" :slot="slot" />

    <!-- Pass on all scoped slots -->
    <template
      v-for="slot in Object.keys($scopedSlots)"
      :slot="slot"
      slot-scope="scope"
      ><slot :name="slot" v-bind="scope"
    /></template>
    <template #open-indicator="{ attributes }">
      <span v-bind="attributes"><i class="fas fa-caret-down"></i></span>
    </template>
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
    label: { type: String, default: "" },
    fontColor: { type: String, default: "" },
  },
  computed: {
    style(): string {
      if (!this.fontColor.length) {
        return "--dropdown-font-color:#333;";
      }
      return `--dropdown-font-color:${this.fontColor};`;
    },
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
  padding: 0px !important;
  border: none !important;
}
div.vs__actions > svg {
  width: 17px;
}
div.vs__selected-options > span {
  color: var(--dropdown-font-color);
}
div.vs__actions > svg {
  fill: var(--dropdown-font-color);
}
.vs__search {
  padding: 0px !important;
  margin: 0px !important;
}
</style>
