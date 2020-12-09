<template>
  <b-modal :id="modalId" size="huge" :title="title">
    <label-header-buttons @button-event="passThrough" />
    <div class="row flex-1 pb-3 h-100">
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Most Similar</h5>
        <component
          :is="viewComponent"
          :instanceName="instanceName"
          :data-items="items"
          :data-fields="dataFields"
          :summaries="summaries"
          includedActive
        />
      </div>
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Randomly Chosen</h5>
        <component
          :is="viewComponent"
          :instanceName="instanceName"
          :data-items="randomItems"
          :data-fields="dataFields"
          :summaries="summaries"
          includedActive
        />
      </div>
    </div>
    <template #modal-footer="{}">
      <!-- Emulate built in modal footer ok and cancel button actions -->
      <b-button size="lg" variant="secondary" @click="onApply">
        <template v-if="isLoading">
          <div v-html="spinnerHTML"></div>
        </template>
        <template v-else> Apply </template>
      </b-button>
    </template>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { Dictionary } from "../../util/dict";
import {
  TableRow,
  VariableSummary,
  TableColumn,
} from "../../store/dataset/index";
import {
  RankedSet,
  ScoreInfo,
  LOW_SHOT_LABEL_COLUMN_NAME,
  LowShotLabels,
  getRandomInt,
} from "../../util/data";
import SelectDataTable from "../SelectDataTable.vue";
import ImageMosaic from "../ImageMosaic.vue";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";
import { circleSpinnerHTML } from "../../util/spinner";

export default Vue.extend({
  name: "label-score-pop-up",
  components: {
    SelectDataTable,
    ImageMosaic,
    LabelHeaderButtons,
  },
  props: {
    data: { type: Array as () => TableRow[], default: () => [] as TableRow[] },
    rankedSet: {
      type: Object as () => RankedSet,
      default: () => {
        return { data: [] as ScoreInfo[] };
      },
    },
    instanceName: {
      type: String as () => string,
      default: "label-score-pop-up",
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: [],
    },
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: {},
    },
    title: { type: String as () => string, default: "Scores" },
    isLoading: { type: Boolean as () => boolean, default: false },
    isRemoteSensing: { type: Boolean as () => boolean, default: false }, // default to false to support every dataset
  },
  data() {
    return { modalId: "score-modal", sampleSize: 10 };
  },
  watch: {
    rankedSet() {
      // on binarySets change show modal
      this.$bvModal.show(this.modalId);
    },
  },
  computed: {
    rankedMap(): Map<number, number> {
      return new Map(
        this.rankedSet?.data.map((d) => {
          return [d.d3mIndex, d.score];
        })
      );
    },
    items(): TableRow[] {
      const ranked: TableRow[] = [];
      this.data?.forEach((d) => {
        if (this.rankedMap.has(d.d3mIndex)) {
          ranked.push(d);
        }
      });
      ranked.sort((a, b) => {
        return this.rankedMap.get(a.d3mIndex) - this.rankedMap.get(b.d3mIndex);
      });
      console.log(this.rankedMap);
      console.table(ranked);
      return ranked;
    },
    randomItems(): TableRow[] {
      const random = this.data?.filter((d) => {
        return (
          !this.rankedMap.has(d.d3mIndex) &&
          d[LOW_SHOT_LABEL_COLUMN_NAME]?.value === LowShotLabels.unlabeled
        );
      });
      if (!random) {
        return null;
      }
      const randomIdx = getRandomInt(random.length - this.sampleSize);
      return random.slice(randomIdx, randomIdx + this.sampleSize);
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    viewComponent(): string {
      if (!this.isRemoteSensing) {
        return "SelectDataTable";
      }
      return "ImageMosaic";
    },
  },
  methods: {
    passThrough(event: string) {
      this.$emit("button-event", event);
    },
    onApply() {
      this.$emit("apply");
    },
  },
});
</script>

<style>
@media (min-width: 992px) {
  .modal .modal-huge {
    max-width: 90% !important;
    width: 90% !important;
  }
}
</style>
