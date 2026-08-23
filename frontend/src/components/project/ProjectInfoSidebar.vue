<!-- eslint-disable vue/no-v-html -->
<template>
	<div
		class="project-info-sidebar"
		:style="{inlineSize: `${width}px`}"
	>
		<div class="project-info-sidebar-content">
			<div
				v-if="htmlDescription !== ''"
				class="project-info-description has-text-start"
				v-html="htmlDescription"
			/>

			<h2 class="project-links-title">
				{{ $t('project.links.title') }}
			</h2>

			<p
				v-if="links.length === 0 && !loading"
				class="project-links-empty is-italic"
			>
				{{ $t('project.links.noLinks') }}
			</p>

			<ProjectLinkItem
				v-for="link in links"
				:key="link.id"
				:link="link"
				@edit="startEdit"
				@delete="confirmDelete"
			/>

			<template v-if="formOpen">
				<FormField>
					<input
						v-model="form.title"
						type="text"
						class="input"
						:placeholder="$t('project.links.titlePlaceholder')"
					>
				</FormField>
				<FormField>
					<input
						v-model="form.url"
						type="url"
						class="input"
						:placeholder="$t('project.links.urlPlaceholder')"
					>
				</FormField>
				<FormField>
					<IconPicker
						v-model="form.icon"
						@custom-icon-selected="onCustomIconSelected"
					/>
				</FormField>
				<div class="project-link-form-actions">
					<XButton
						:disabled="!canSave"
						:loading="saving"
						@click="save"
					>
						{{ $t('project.links.save') }}
					</XButton>
					<XButton
						variant="tertiary"
						@click="closeForm"
					>
						{{ $t('project.links.cancel') }}
					</XButton>
				</div>
			</template>
			<XButton
				v-else
				icon="plus"
				variant="secondary"
				:shadow="false"
				class="mbs-2"
				@click="startCreate"
			>
				{{ $t('project.links.addLink') }}
			</XButton>
		</div>

		<button
			class="project-info-sidebar-resize-handle"
			:aria-label="$t('project.links.resizeHandleLabel')"
			@pointerdown="startResize"
		/>

		<Modal
			:enabled="linkToDelete !== null"
			@close="linkToDelete = null"
			@submit="doDelete"
		>
			<template #header>
				<span>{{ $t('project.links.delete') }}</span>
			</template>
			<template #text>
				<p>{{ linkToDelete?.title }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, reactive, ref, shallowReactive, watch} from 'vue'
import {useStorage} from '@vueuse/core'
import DOMPurify from 'dompurify'

import FormField from '@/components/input/FormField.vue'
import Modal from '@/components/misc/Modal.vue'
import IconPicker from '@/components/project/IconPicker.vue'
import ProjectLinkItem from '@/components/project/ProjectLinkItem.vue'

import ProjectLinkService from '@/services/projectLink'
import ProjectLinkModel from '@/models/projectLink'
import type {IProjectLink} from '@/modelTypes/IProjectLink'
import type {IProject} from '@/modelTypes/IProject'
import {error} from '@/message'

const MIN_WIDTH = 220
const MAX_WIDTH = 600
const DEFAULT_WIDTH = 300

const props = defineProps<{
	project: IProject | null
}>()

// ADD_ATTR re-permits target, which lets a link hand the opened page a live
// window.opener back to this tab. rel must be forced here rather than allowed,
// so a description carrying its own rel can't drop noopener. Same hook as
// views/project/ProjectInfo.vue.
DOMPurify.addHook('afterSanitizeAttributes', node => {
	if (node.hasAttribute('target')) {
		node.setAttribute('rel', 'noopener noreferrer')
	}
})

const htmlDescription = computed(() => {
	const description = props.project?.description || ''
	if (description === '') {
		return ''
	}
	return DOMPurify.sanitize(description, {ADD_ATTR: ['target']})
})

const width = useStorage('projectInfoSidebarWidth', DEFAULT_WIDTH)

let resizeStartX = 0
let resizeStartWidth = 0

function startResize(e: PointerEvent) {
	resizeStartX = e.clientX
	resizeStartWidth = width.value
	window.addEventListener('pointermove', onResize)
	window.addEventListener('pointerup', stopResize, {once: true})
}

function onResize(e: PointerEvent) {
	const delta = e.clientX - resizeStartX
	width.value = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, resizeStartWidth + delta))
}

function stopResize() {
	window.removeEventListener('pointermove', onResize)
}

onBeforeUnmount(stopResize)

const projectLinkService = shallowReactive(new ProjectLinkService())
const loading = computed(() => projectLinkService.loading)
const links = ref<IProjectLink[]>([])

async function loadLinks() {
	if (!props.project?.id) {
		links.value = []
		return
	}
	try {
		links.value = await projectLinkService.getAll(props.project.id)
	} catch (e) {
		error(e)
	}
}

watch(() => props.project?.id, loadLinks, {immediate: true})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const pendingCustomIcon = ref<File | null>(null)
const form = reactive({
	title: '',
	url: '',
	icon: '',
})

const canSave = computed(() => form.title.trim() !== '' && form.url.trim() !== '')

function startCreate() {
	editingId.value = null
	form.title = ''
	form.url = ''
	form.icon = ''
	pendingCustomIcon.value = null
	formOpen.value = true
}

function startEdit(link: IProjectLink) {
	editingId.value = link.id
	form.title = link.title
	form.url = link.url
	form.icon = link.icon
	pendingCustomIcon.value = null
	formOpen.value = true
}

function closeForm() {
	formOpen.value = false
	editingId.value = null
	pendingCustomIcon.value = null
}

function onCustomIconSelected(file: File) {
	pendingCustomIcon.value = file
	form.icon = ''
}

async function save() {
	if (!canSave.value || !props.project?.id) {
		return
	}

	saving.value = true
	try {
		let saved: IProjectLink
		if (editingId.value !== null) {
			const existing = links.value.find(l => l.id === editingId.value)
			saved = await projectLinkService.update(new ProjectLinkModel({
				...existing,
				title: form.title,
				url: form.url,
				icon: form.icon,
			}))
		} else {
			saved = await projectLinkService.create(props.project.id, {
				title: form.title,
				url: form.url,
				icon: form.icon,
			})
		}

		if (pendingCustomIcon.value) {
			saved = await projectLinkService.uploadCustomIcon(saved, pendingCustomIcon.value)
		}

		if (editingId.value !== null) {
			const index = links.value.findIndex(l => l.id === editingId.value)
			if (index !== -1) {
				links.value.splice(index, 1, saved)
			}
		} else {
			links.value.push(saved)
		}

		closeForm()
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

const linkToDelete = ref<IProjectLink | null>(null)

function confirmDelete(link: IProjectLink) {
	linkToDelete.value = link
}

async function doDelete() {
	if (!linkToDelete.value) {
		return
	}
	try {
		await projectLinkService.delete(linkToDelete.value)
		links.value = links.value.filter(l => l.id !== linkToDelete.value!.id)
	} catch (e) {
		error(e)
	} finally {
		linkToDelete.value = null
	}
}
</script>

<style lang="scss" scoped>
.project-info-sidebar {
	position: relative;
	flex-shrink: 0;
	padding-inline-end: .75rem;
	border-inline-end: 1px solid var(--grey-200);

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

.project-info-sidebar-content {
	overflow-y: auto;
	max-block-size: calc(100vh - 12rem);
}

.project-info-description {
	margin-block-end: 1rem;
	font-size: .9rem;
}

.project-links-title {
	font-size: 1rem;
	font-weight: bold;
	margin-block-end: .5rem;
}

.project-links-empty {
	color: var(--grey-500);
	font-size: .85rem;
}

.project-link-form-actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .5rem;
}

.project-info-sidebar-resize-handle {
	position: absolute;
	inset-block: 0;
	inset-inline-end: -3px;
	inline-size: 6px;
	padding: 0;
	border: 0;
	background: transparent;
	cursor: col-resize;
	touch-action: none;

	&:hover,
	&:active {
		background: var(--grey-300);
	}
}
</style>
