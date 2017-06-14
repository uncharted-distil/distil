<template>
	<div id="data-table">
		<b-table respnosive bordered hover :items="items" :fields="fields">
		</b-table>
	</div>
</template>

<script>

export default {
	name: 'data-table',
	data() {
		return {
			items: [
				{
					isActive: true,
					age: 40,
					name: {
						first: 'Dickerson',
						last: 'Macdonald'
					}
				},
				{
					isActive: false,
					age: 21,
					name: {
						first: 'Larsen',
						last: 'Shaw'
					}

				},
				{
					isActive: false,
					age: 9,
					state: 'success',
					name: {
						first: 'Mitzi',
						last: 'Navarro'
					}
				}
			],
			fields: {
				name: {
					label: 'Person Full name',
					sortable: true
				},
				age: {
					label: 'Person age',
					sortable: true
				},
				isActive: {
					label: 'is Active'
				},
				actions: {
					label: 'Actions'
				}
			},
		};
	}
};
</script>

<style>

</style>
