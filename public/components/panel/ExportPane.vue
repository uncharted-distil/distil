<template lang="">
  <div> <p>Export your current highlighted and filtered data to file</p>
  <b-form> <b-input-group append=".csv"> <b-form-input v-model="fileName"
  :state="validation"></b-form-input> </b-input-group> <b-form-invalid-feedback
  :state="validation"> Filename cannot be empty or contain invalid characters.
  </b-form-invalid-feedback> <b-button class="mt-2" variant="dark"
  @click="onExportClick" :disabled="!validation"><i class="fa fa-floppy-o" />
  Export</b-button> </b-form> </div>
</template>
<script>
import { getters as routeGetters } from "../../store/route/module";
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
      this.$eventBus.$emit(EventList.EXPLORER.EXPLORER_EXPORT, this.fileName);
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
