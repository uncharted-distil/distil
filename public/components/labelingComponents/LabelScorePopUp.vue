<template>
  <b-modal :id="modalId" size="huge" :title="title">
    <label-header-buttons @button-event="passThrough" />
    <div class="row flex-1 pb-3 h-100">
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Positive</h5>
        <select-data-table
          :instanceName="instanceName"
          :data-items="items.positive"
          :data-fields="dataFields"
          :summaries="summaries"
          includedActive
        />
      </div>
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Negative</h5>
        <select-data-table
          :instanceName="instanceName"
          :data-items="items.negative"
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
import { BinarySets, ScoreInfo } from "../../util/data";
import SelectDataTable from "../SelectDataTable.vue";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";
import { circleSpinnerHTML } from "../../util/spinner";

export default Vue.extend({
  name: "label-score-pop-up",
  components: {
    SelectDataTable,
    LabelHeaderButtons,
  },
  props: {
    data: { type: Array as () => TableRow[], default: () => [] },
    binarySets: {
      type: Object as () => BinarySets,
      default: () => {
        return { positive: [] as ScoreInfo[], negative: [] as ScoreInfo[] };
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
  },
  data() {
    return { modalId: "score-modal" };
  },
  watch: {
    binarySets() {
      // on binarySets change show modal
      this.$bvModal.show(this.modalId);
    },
  },
  computed: {
    positiveIds(): Map<number, number> {
      return new Map(
        this.binarySets?.positive.map((p) => {
          return [p.d3mIndex, p.score];
        })
      );
    },
    negativeIds(): Map<number, number> {
      return new Map(
        this.binarySets?.negative.map((n) => {
          return [n.d3mIndex, n.score];
        })
      );
    },
    items(): { positive: TableRow[]; negative: TableRow[] } {
      const positive: TableRow[] = [];
      const negative: TableRow[] = [];
      this.data?.forEach((d) => {
        if (this.positiveIds.has(d.d3mIndex)) {
          positive.push(d);
        } else if (this.negativeIds.has(d.d3mIndex)) {
          negative.push(d);
        }
      });
      positive.sort((a, b) => {
        return (
          this.positiveIds.get(a.d3mIndex) - this.positiveIds.get(b.d3mIndex)
        );
      });
      negative.sort((a, b) => {
        return (
          this.negativeIds.get(a.d3mIndex) + this.negativeIds.get(b.d3mIndex)
        );
      });
      return { positive, negative };
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
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
