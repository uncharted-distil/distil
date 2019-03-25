<template>
<div class="notification-sidebar">
    <div class="notification-icons">
        <div class="event-icon">
            <i class="fa fa-2x fa-info" aria-hidden="true"></i>
        </div>
        <div class="event-icon">
            <i class="fa fa-2x fa-plus-square" aria-hidden="true"></i>
        </div>
        <div class="event-icon">
            <i class="fa fa-2x fa-long-arrow-right" aria-hidden="true"></i>
        </div>
    </div>
    <div class="notification-detail">

    </div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';

export default Vue.extend({
	name: 'notification-sidebar',
});

</script>

<style>

.notification-sidebar {
	background-color: #fff;
	width: 55px;
	border-left: 1px solid #ccc;
	height: 100%;
	display: flex;
	flex-direction: column;
}
.notification-sidebar .event-icon {
	padding-top: 15px;
	padding-bottom: 15px;
	text-align: center;
}
.event-icon i {
	height: 30px;
	width: 30px;
	cursor: pointer;
}

</style>
