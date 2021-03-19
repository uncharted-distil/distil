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
  <b-form-group
    label="Positive Label:"
    label-class="font-weight-bold"
    label-cols="auto"
    label-for="positive-label"
    label-size="sm"
  >
    <b-form-select
      id="positive-label"
      v-model="positiveLabel"
      :options="labels"
      size="sm"
    />
  </b-form-group>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../../store/route/module";
import { overlayRouteEntry } from "../../util/routes";

/**
 * Library to find best available positive label
 * https://github.com/aceakash/string-similarity#findbestmatchmainstring-targetstrings
 */
import { findBestMatch } from "string-similarity";
const findBestRating = (
  mainString: string,
  targetStrings: string[]
): number => {
  return findBestMatch(mainString, targetStrings)?.bestMatch.rating ?? 0;
};

// List of positives and negatives words that could be used in labels
const positives = ["true", "positive", "aff", "1", "yes", "good", "high"];
const negatives = ["false", "negative", "not", "0", "no", "bad", "low"];

/** Dropdown to select a positive label for Binary Classification task */
export default Vue.extend({
  name: "PositiveLabel",

  props: {
    labels: {
      type: Array as () => string[],
      default: null,
    },
  },

  data() {
    return {
      positiveLabel: null as string,
    };
  },

  computed: {
    routePositiveLabel(): string {
      return routeGetters.getPositiveLabel(this.$store);
    },
  },

  watch: {
    // update the route on positive label changes
    positiveLabel(value: string, oldValue: string): void {
      if (value === oldValue) return;
      if (value === this.routePositiveLabel) return;
      const entry = overlayRouteEntry(this.$route, { positiveLabel: value });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },

  beforeMount() {
    // If the positive label is already set in the route, pre-select it,
    // otherwise, find the label that's most likely to be a positive one.
    this.positiveLabel = !!this.routePositiveLabel
      ? this.routePositiveLabel
      : this.findAPositiveLabel(this.labels);
  },

  methods: {
    // Find which labels is most suited to be the positive one
    findAPositiveLabel(labels: string[]): string {
      const ratings = labels.map((label) => {
        return {
          positive: findBestRating(label, positives),
          negative: findBestRating(label, negatives),
        };
      });

      // Default to the first label
      let positiveLabel = labels[0];

      // Select the second label, if the first label...
      if (
        // has a lower or identical positive rating and
        ratings[0].positive <= ratings[1].positive &&
        // has a higher negative rating
        ratings[0].negative > ratings[1].negative
      ) {
        positiveLabel = labels[1];
      }

      return positiveLabel;
    },
  },
});
</script>
