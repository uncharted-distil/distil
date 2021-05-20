<template>
  <b-modal id="join-accuracy-modal" size="xl" @ok="updateRoute">
    <form ref="joinAccuracyForm">
      <b-container class="bv-example-row">
        <b-row class="p-2">
          <b-col md><b>Join Relationship</b></b-col>
          <b-col sm><b>Absolute</b></b-col>
          <b-col md><b>Accuracy</b></b-col>
        </b-row>
        <b-row
          v-for="(ad, idx) in accuracyData"
          :key="ad.joinPair.first + ad.joinPair.second"
          class="p-2"
        >
          <b-col md>
            <b>{{ ad.joinPair.first }}->{{ ad.joinPair.second }}</b>
          </b-col>
          <b-col sm
            ><b-form-checkbox
              v-model="ad.absolute"
              :disabled="ad.unitType === 3"
            ></b-form-checkbox
          ></b-col>
          <b-col md>
            <b-tooltip
              v-if="!ad.absolute"
              :target="ad.joinPair.first + ad.joinPair.second"
              placement="right"
            >
              {{ ad.accuracy }}
            </b-tooltip>
            <b-form-input
              v-if="!ad.absolute"
              :id="ad.joinPair.first + ad.joinPair.second"
              v-model="ad.accuracy"
              type="range"
              min="0"
              max="1"
              step="0.1"
              class="mt-1"
            />
            <div v-else class="d-flex">
              <b-form-input type="number" class="d-flex max-width-200" />
              <b-dropdown
                v-if="ad.unitType != 2"
                class="d-flex pl-2"
                variant="outline-secondary"
              >
                <template v-slot:button-content> {{ ad.unit }} </template>
                <b-dropdown-item
                  v-for="unit in getItems(ad.unitType)"
                  :key="unit"
                  @click.stop="setAccuracy(unit, idx)"
                >
                  {{ unit }}
                </b-dropdown-item>
              </b-dropdown>
            </div>
          </b-col>
        </b-row>
      </b-container>
    </form>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { JoinPair } from "../util/data";
import { getters as routeGetters } from "../store/route/module";
import { Variable } from "../store/dataset/index";
import { AccuracyData, UnitTypes } from "../util/data";
import vueSlider from "vue-slider-component";
import { DATE_TIME_TYPE, isGeoLocatedType, isNumericType } from "../util/types";
import { overlayRouteEntry } from "../util/routes";
enum TimeUnits {
  Hours = "Hours",
  Days = "Days",
  Weeks = "Weeks",
  Years = "Years",
}
enum DistanceUnits {
  Meters = "Meters",
  Kilometers = "Kilometers",
}

export default Vue.extend({
  name: "JoinAccuracyModal",
  components: {
    vueSlider,
  },
  data() {
    return {
      accuracyData: [] as AccuracyData[],
      show: true,
    };
  },
  computed: {
    joinPairs(): JoinPair<string>[] {
      return routeGetters.getJoinPairs(this.$store);
    },
    joinDatasets(): string[] {
      return routeGetters.getRouteJoinDatasets(this.$store);
    },
    routeAccuracyData(): AccuracyData[] {
      return routeGetters.getJoinInfo(this.$store);
    },
    variables(): Variable[] {
      if (!this.joinDatasets.length) {
        return [];
      }
      const ds = this.joinDatasets[0];
      return routeGetters.getJoinDatasetsVariables(this.$store).filter((v) => {
        return v.datasetName === ds;
      });
    },
  },
  mounted() {
    if (this.routeAccuracyData != null) {
      this.accuracyData = this.routeAccuracyData;
    }
  },
  watch: {
    joinPairs(cur: JoinPair<string>[], prev: JoinPair<string>[]) {
      if (!cur.length) {
        return;
      }
      if (cur.length < prev.length) {
        const curMap = new Map(
          cur.map((c) => {
            return [c.first + c.second, true];
          })
        );
        this.accuracyData = this.accuracyData.filter((ad) => {
          const key = ad.joinPair.first + ad.joinPair.second;
          return curMap.has(key);
        });
      } else {
        const end = cur.length - 1;
        const unitType = this.getUnitTypes(cur[end]);
        const key = cur[end].first + cur[end].second;
        if (
          !this.accuracyData.some((ad) => {
            return ad.joinPair.first + ad.joinPair.second === key;
          })
        ) {
          this.accuracyData.push({
            joinPair: cur[end],
            absolute: false,
            accuracy: 0.8,
            unitType,
            unit: this.getDefaultUnit(unitType),
          });
        }
      }
    },
  },
  methods: {
    updateRoute() {
      const route = this.$route;
      const entry = overlayRouteEntry(route, {
        joinInfo: JSON.stringify(this.accuracyData),
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    setAccuracy(unit: string, idx: number) {
      this.accuracyData[idx].unit = unit;
    },
    getItems(unitType: UnitTypes): string[] {
      if (unitType === UnitTypes.Time) {
        return Object.values(TimeUnits);
      }
      return Object.values(DistanceUnits);
    },
    getUnitTypes(joinPair: JoinPair<string>): UnitTypes {
      const key = joinPair.first;
      const variable = this.variables.find((v) => {
        return v.key === key;
      });
      if (isGeoLocatedType(variable.colType)) {
        return UnitTypes.Distance;
      }
      if (variable.colType === DATE_TIME_TYPE) {
        return UnitTypes.Time;
      }
      if (isNumericType(variable.colType)) {
        return UnitTypes.None;
      }
      return UnitTypes.Disabled;
    },
    getDefaultUnit(unitType: UnitTypes): string {
      if (unitType === UnitTypes.None || unitType === UnitTypes.Disabled) {
        return "";
      }
      if (unitType === UnitTypes.Time) {
        return TimeUnits.Days;
      }
      if (unitType === UnitTypes.Distance) {
        return DistanceUnits.Meters;
      }
    },
  },
});
</script>

<style scoped>
.max-width-200 {
  max-width: 200px;
}
</style>
