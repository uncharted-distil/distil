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
  <b-button
    class="toggle d-flex align-items-center shadow-none mr-1"
    variant="secondary"
    @click="updateExplore"
  >
    <i :class="buttonClass" />
  </b-button>
</template>

<script lang="ts">
import Vue from "vue";

import { getters as routeGetters } from "../store/route/module";
import { RouteArgs, overlayRouteEntry } from "../util/routes";

export default Vue.extend({
  name: "ToggleExplore",
  props: {
    variable: String as () => string,
  },
  computed: {
    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
    },

    isExplore(): boolean {
      return this.explore.includes(this.variable);
    },

    buttonClass(): string {
      return `fa-sm fas ${this.isExplore ? "fa-eye" : "fa-eye-slash"}`;
    },
  },
  methods: {
    updateExplore(): void {
      const variable = this.variable;
      const args = {} as RouteArgs;
      if (this.isExplore) {
        args.explore = this.explore.filter((v) => v !== variable).join(",");
      } else {
        args.explore = this.explore.concat([variable]).join(",");
      }
      this.updateRoute(args);
    },

    updateRoute(args: RouteArgs) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style scoped>
.toggle {
  height: 22px;
}
</style>
