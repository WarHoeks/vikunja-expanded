<template>
	<div class="task-link-item">
		<a
			:href="link.url"
			target="_blank"
			rel="noopener noreferrer"
			class="task-link-title"
		>
			{{ link.title }}
		</a>
		<span
			v-if="editEnabled"
			class="task-link-actions"
		>
			<BaseButton
				v-tooltip="$t('project.links.edit')"
				:aria-label="$t('project.links.edit')"
				@click="$emit('edit', link)"
			>
				<Icon icon="pen" />
			</BaseButton>
			<BaseButton
				v-tooltip="$t('project.links.delete')"
				:aria-label="$t('project.links.delete')"
				@click="$emit('delete', link)"
			>
				<Icon icon="trash-alt" />
			</BaseButton>
		</span>
	</div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import type {ITaskLink} from '@/modelTypes/ITaskLink'

withDefaults(defineProps<{
	link: ITaskLink
	editEnabled?: boolean
}>(), {
	editEnabled: true,
})

defineEmits<{
	'edit': [link: ITaskLink],
	'delete': [link: ITaskLink],
}>()
</script>

<style lang="scss" scoped>
.task-link-item {
	display: flex;
	align-items: center;
	gap: .5rem;
	padding: .4rem 0;
}

.task-link-title {
	flex: 1;
	min-inline-size: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	color: var(--text);

	&:hover {
		color: var(--primary);
	}
}

.task-link-actions {
	display: flex;
	gap: .1rem;
	color: var(--grey-400);
	flex-shrink: 0;
}
</style>
