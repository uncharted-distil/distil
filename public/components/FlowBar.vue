<template>
	<div class="flow-container">
		<div class="flow-header">
			<div class="flow-header-left">
				<a class="hover text-success"  v-if="leftText" @click="onLeft">
					<i class="fa fa-long-arrow-left"></i>
					<strong>{{leftText}}</strong>
				</a>
			</div>
			<h4 class="flow-header-center">{{centerText}}</h4>
			<div class="flow-header-right">
				<a class="hover text-success"  v-if="rightText" @click="onRight">
					<strong>{{rightText}}</strong>
					<i class="fa fa-long-arrow-right"></i>
				</a>
			</div>
		</div>
	</div>
</template>

<script>

export default {
	name: 'flow-bar',

	props: {
		rightText: String,
		onRight: Function,
		leftText: String,
		onLeft: Function,
		centerText: String
	}
};

</script>

<style>
.flow-container {
	width: 100%;
}
.flow-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	width: 100%;
}
.flow-header-center {
	color: #333;
	margin: 0.75rem 0;
	width: 60%;
	text-align: center;
}
.flow-header-right {
	text-align: right;
	padding-right: 32px;
}
.flow-header-left {
	text-align: left;
	padding-left: 32px;
}
.flow-header-left,
.flow-header-right {
	width: 20%;
}
</style>
