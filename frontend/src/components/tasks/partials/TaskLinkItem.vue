<!-- eslint-disable vue/no-v-html -->
<template>
	<div class="task-link-item">
		<span class="task-link-icon">
			<img
				v-if="link.customIconId > 0 && customIconUrl"
				:src="customIconUrl"
				alt=""
			>
			<span
				v-else-if="link.icon"
				v-html="svg"
			/>
			<Icon
				v-else
				icon="link"
			/>
		</span>
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
import {computed, ref, watch} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import {useSimpleIconSvg} from '@/helpers/simpleIcons'
import TaskLinkService from '@/services/taskLink'
import type {ITaskLink} from '@/modelTypes/ITaskLink'

const props = withDefaults(defineProps<{
	link: ITaskLink
	editEnabled?: boolean
}>(), {
	editEnabled: true,
})

defineEmits<{
	'edit': [link: ITaskLink],
	'delete': [link: ITaskLink],
}>()

const svg = useSimpleIconSvg(computed(() => props.link.icon))

const customIconUrl = ref<string>('')
const taskLinkService = new TaskLinkService()

watch(
	() => [props.link.customIconId, props.link.id] as const,
	async ([customIconId]) => {
		if (customIconId > 0) {
			customIconUrl.value = await taskLinkService.getCustomIconBlobUrl(props.link)
		}
	},
	{immediate: true},
)
</script>

<style lang="scss" scoped>
.task-link-item {
	display: flex;
	align-items: center;
	gap: .5rem;
	padding: .4rem 0;
}

.task-link-icon {
	inline-size: 1.2rem;
	block-size: 1.2rem;
	flex-shrink: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--grey-500);

	img {
		inline-size: 100%;
		block-size: 100%;
		object-fit: contain;
	}

	:deep(svg) {
		inline-size: 100%;
		block-size: 100%;
		fill: currentColor;
	}
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
