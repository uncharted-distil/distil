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
    <b-dropdown-divider v-if="displayImageAttention" />
    <b-dropdown-item
      v-if="displayImageAttention"
      :disabled="imageAttentionEnabled"
      @click="enableImageAttention()"
      >Enable Image Attention</b-dropdown-item
    >
    <b-dropdown-item
      v-if="displayImageAttention"
      :disabled="!imageAttentionEnabled"
      @click="enableImageAttention()"
      >Disable Image Attention</b-dropdown-item
    >
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

  props: { hasImageAttention: { type: Boolean, default: false } },

  data() {
    return {
      imageAttentionEnabled: false,
    };
  },

  computed: {
    // Returns currently selected band ID as a string
    band(): string {
      const bandID = routeGetters.getBandCombinationId(this.$store);
      return this.availableBands.find((b) => b.id === bandID)?.displayName;
    },

    // Returns list of band combinations
    availableBands(): BandCombination[] {
      return datasetGetters
        .getMultiBandCombinations(this.$store)
        .slice() // copy so we don't mutate vuex store object
        .sort((a, b) => a.displayName.localeCompare(b.displayName));
    },
    displayImageAttention(): boolean {
      return this.hasImageAttention;
    },
  },

  methods: {
    // Writes a new band combination into the route
    setBandCombination(bandID: BandID) {
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        bandCombinationId: bandID,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    enableImageAttention() {
      this.imageAttentionEnabled = !this.imageAttentionEnabled;
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        imageAttention: this.imageAttentionEnabled,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
  created() {
    this.imageAttentionEnabled = routeGetters.getImageAttention(this.$store);
  },
});
</script>

<style></style>
