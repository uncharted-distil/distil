<template lang="">
  <div> <b-form> <b-input-group append=".csv"> <b-form-input v-model="fileName"
  :state="validation"></b-form-input> </b-input-group> <b-form-invalid-feedback
  :state="validation"> Filename cannot be empty or contain invalid characters.
  </b-form-invalid-feedback> <b-button class="mt-2" variant="dark"
  @click="onExportClick"> <i class="fa fa-floppy-o" /> Export </b-button>
  </b-form> </div>
</template>
<script>
import { getters as routeGetters } from "../../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import { EventList } from "../../util/events";

export default {
  name: "ExportPane",
  data() {
    return {
      fileName: routeGetters.getRouteDataset(this.$store) || "",
    };
  },
  methods: {
    onExportClick() {
      this.$emit(EventList.VARIABLES.EXPORT_EVENT, {
        filename: this.fileName,
      });
    },
  },
  computed: {
    validation() {
      const fileNameRegex = /[<>:"/\\|?*\u0000-\u001F]/g;
      return this.fileName.length > 0 && !fileNameRegex.test(this.fileName);
    },
  },
};
</script>
<style lang=""></style>
