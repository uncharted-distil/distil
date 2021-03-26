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
  <div class="badge">
    <span
      ><b>{{ content }}</b></span
    >
    <b-button size="sm" @click="onRemove" title="Remove filter">
      <i class="fa fa-times" />
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "badge",
  props: {
    content: String as () => string,
    identifier: Object,
  },
  methods: {
    onRemove() {
      this.$emit("removed", this.identifier);
    },
  },
});
</script>

<style scoped>
.badge {
  --height: 28px;
  background-color: var(--blue);
  border-radius: 2px;
  color: var(--white);
  display: inline-flex;
  height: var(--height);
  margin: 0.2rem;
  overflow: hidden; /* hide issue with button:hover styling */
  position: relative;
}

.active {
  background-color: var(--blue);
}

span {
  line-height: var(--height);
  padding: 0 0.5em;
}

.btn {
  color: #fff;
  background: none;
  border: none;
  border-left: 1px solid var(--white);
  border-radius: 0;
}

.btn:hover {
  color: var(--white);
  background-color: var(--blue);
  border-left: 1px solid var(--white);
}

.active .btn:hover {
  background-color: #3d70d3;
}
</style>
