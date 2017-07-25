<template>
	<div>
		<div class="row justify-content-center mt-2 mb-2">
			<h4>Distil</h4>

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
		this.$store.dispatch('startPipelineSession');
	},
	computed: {
		sessionID() {
			return this.$store.getters.getPipelineSessionID();
		}
	}
};
</script>

<style>
.session-id {
	position: absolute;
	text-align: right;
	right: 0;
	padding-right: 4px;
}
.session-not-ready {
	color: #ff6562;
}
.session-ready {
	color: #00c07f;
}
</style>
