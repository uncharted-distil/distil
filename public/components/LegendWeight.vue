<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

<template>
  <div class="legend">
    Importance
    <ol>
      <li title="Low" :style="{ '--color': colorByWeight(0) }" />
      <li title="Medium-low" :style="{ '--color': colorByWeight(0.25) }" />
      <li title="Medium" :style="{ '--color': colorByWeight(0.5) }" />
      <li title="Medium-high" :style="{ '--color': colorByWeight(0.75) }" />
      <li title="High" :style="{ '--color': colorByWeight(1) }" />
    </ol>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { colorByWeight } from "../util/data";

export default Vue.extend({
  name: "legend-weight",
  methods: { colorByWeight },
});
</script>

<style scoped>
.legend {
  display: flex;
  flex-direction: row;
  font-size: 0.9em;
  font-weight: normal;
}

.legend ol {
  display: flex;
  flex-direction: row;
  list-style: none;
  padding: unset;
}

.legend li {
  background-color: var(
    --color /* Get the weight colour from the method used for the results. */
  );
  border: 1px solid var(--gray-500); /* To make the colours pop from a light background. */
  height: 1.5rem;
  margin-left: 0.33rem;
  position: relative; /* for the visible label */
  width: 1.5rem;
}

/* Display a label underneath the first and last one. */
.legend li::after {
  font-size: 0.7em;
  position: absolute;
  text-transform: uppercase;
  top: 100%;
}
.legend li:first-of-type::after,
.legend li:last-of-type::after {
  content: attr(title); /* Use the title as a label. */
}
</style>
