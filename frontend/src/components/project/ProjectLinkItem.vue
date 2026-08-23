<!-- eslint-disable vue/no-v-html -->
<template>
	<div class="project-link-item">
		<span class="project-link-icon">
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
			class="project-link-title"
		>
			{{ link.title }}
		</a>
		<span class="project-link-actions">
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
import ProjectLinkService from '@/services/projectLink'
import type {IProjectLink} from '@/modelTypes/IProjectLink'
import {error} from '@/message'

const props = defineProps<{
	link: IProjectLink
}>()

defineEmits<{
	'edit': [link: IProjectLink],
	'delete': [link: IProjectLink],
}>()

const svg = useSimpleIconSvg(computed(() => props.link.icon))

const customIconUrl = ref<string>('')
const projectLinkService = new ProjectLinkService()

watch(
	() => [props.link.customIconId, props.link.id] as const,
	async ([customIconId]) => {
		if (customIconId > 0) {
			// Without this catch, a failed fetch (e.g. a stale/broken file
			// reference) left the icon silently blank with no visible error —
			// that's exactly what made this bug so hard to diagnose from a report.
			try {
				customIconUrl.value = await projectLinkService.getCustomIconBlobUrl(props.link)
			} catch (e) {
				error(e)
			}
		}
	},
	{immediate: true},
)
</script>

<style lang="scss" scoped>
.project-link-item {
	display: flex;
	align-items: center;
	gap: .5rem;
	padding: .4rem 0;
}

.project-link-icon {
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

.project-link-title {
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

.project-link-actions {
	display: flex;
	gap: .1rem;
	color: var(--grey-400);
	flex-shrink: 0;
}
</style>
