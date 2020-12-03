<template>
  <nav class="action-column-container">
    <ul class="action-column-nav-bar" role="tablist">
      <li
        v-for="(action, index) in actions"
        :key="index"
        :title="action.name"
        :data-count="action.count"
      >
        <b-button
          role="tab"
          data-toggle="tab"
          :variant="action.name === currentAction ? 'primary' : 'light'"
          @click.stop.prevent="setActive(action.name)"
        >
          <i :class="'fa fa-' + action.icon" />
        </b-button>
      </li>
    </ul>
  </nav>
</template>

<script lang="ts">
import Vue from "vue";

export interface Action {
  name: string;
  icon: string;
  paneId: string;
  count?: number;
}

export default Vue.extend({
  name: "ActionColumn",

  props: {
    actions: { type: Array as () => Action[], default: () => [] },
    currentAction: { type: String, default: "" },
  },

  methods: {
    setActive(actionName: string): void {
      // If the action is currently selected, pass ''
      // to signify it should be unselected.  Otherwise, pass
      // the action's name to select it.
      const name = actionName === this.currentAction ? "" : actionName;
      this.$emit("set-active-pane", name);
    },
  },
});
</script>

<style scoped>
.action-column-container {
  --width: var(--width-action-column, 3rem);
  height: 100%;
  position: relative;
  width: var(--width);
}

.action-column-nav-bar {
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  width: var(--width);
  margin: 0;
  padding: 0;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  border: 1px solid rgba(207, 216, 220, 0.5);
}

.action-column-nav-bar li {
  position: relative;
  display: block;
}

/* Display a count to know the number of variables. */
.action-column-nav-bar li[data-count]::after {
  background-color: var(--color-text-disable);
  border-radius: 50%;
  color: var(--white);
  content: attr(data-count);
  display: block;
  font-size: 0.5rem;
  height: 2em;
  line-height: 2em;
  position: absolute;
  right: 0.5em;
  top: 0.5em;
  text-align: center;
  width: 2em;
}

.action-column-nav-bar button {
  width: var(--width);
  height: var(--width);
}
</style>
