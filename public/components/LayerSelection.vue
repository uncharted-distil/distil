<template>
  <b-dropdown variant="outline-secondary">
    <template v-slot:button-content>
      <i class="fa fa-clone"></i> Image Layer: <b>{{ band }}</b>
    </template>
    <b-dropdown-item
      v-for="bandInfo in availableBands"
      :key="bandInfo.id"
      @click="setBandCombination(bandInfo.id)"
      >{{ bandInfo.displayName }}
    </b-dropdown-item>
  </b-dropdown>
</template>

<script lang="ts">
import Vue from "vue";
import { BandCombination, BandID } from "../store/dataset";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { overlayRouteEntry } from "../util/routes";

export default Vue.extend({
  name: "layer-selection",

  props: {},

  data() {
    return {};
  },

  computed: {
    // Returns currently selected band ID as a string
    band(): string {
      const bandID = routeGetters.getBandCombinationId(this.$store);
      return this.availableBands.find(b => b.id === bandID)?.displayName;
    },

    // Returns list of band combinations
    availableBands(): BandCombination[] {
      return datasetGetters
        .getMultiBandCombinations(this.$store)
        .sort((a, b) => a.displayName.localeCompare(b.displayName));
    }
  },

  methods: {
    // Writes a new band combination into the route
    setBandCombination(bandID: BandID) {
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        bandCombinationId: bandID
      });
      this.$router.push(entry);
    }
  }
});
</script>

<style></style>