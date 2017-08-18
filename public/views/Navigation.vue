<template>
	<div>
		<div class="row justify-content-center mt-2 mb-2">
			<h4>Distil</h4>
		</div>

		<div class="row justify-content-center mt-2 mb-2">
			<div class="col-md-8"></div>
			<div class="session-header col-md-4">
				<div class="session-id">
					Session ID:
					<span v-if="sessionID===null" class="session-not-ready">
						<i class="fa fa-close"></i>
						Unavailable
					</span>
					<span v-if="sessionID!==null" class="session-ready">
						<i class="fa fa-check"></i>
						{{sessionID}}
					</span>
				</div>
				<div class="session-pipelines">
					<button type="button" class="pipeline-uuid btn btn-primary btn-sm" v-for="uuid in pipelineUUIDs">{{uuid}}</button>
				</div>
			</div>
		</div>

		<div class="row justify-content-center mt-2 mb-2">
			<search-bar class="col-md-6"></search-bar>
		</div>
	</div>
</template>

<script>
import SearchBar from '../components/SearchBar';

export default {
	name: 'navigation',
	components: {
		SearchBar
	},
	mounted() {
		this.$store.dispatch('getPipelineSession');
	},
	computed: {
		sessionID() {
			return this.$store.getters.getPipelineSessionID();
		},
		session() {
			return this.$store.getters.getPipelineSession();
		},
		pipelineUUIDs() {
			return this.$store.getters.getPipelineSessionUUIDs().map(uuid => {
				return uuid.substr(uuid.length - 4);
			});
		}
	}
};
</script>

<style>
.session-header {
	text-align: right;
}
.session-not-ready {
	color: #ff6562;
}
.session-ready {
	color: #00c07f;
}
.pipeline-uuid {
	cursor: pointer;
	margin: 0 2px;
}
</style>
