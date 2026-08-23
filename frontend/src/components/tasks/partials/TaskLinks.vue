<template>
	<div class="task-links">
		<h2 class="task-section-title">
			<span class="icon is-grey">
				<Icon icon="link" />
			</span>
			{{ $t('task.links.title') }}
		</h2>

		<p
			v-if="links.length === 0 && !loading"
			class="task-links-empty is-italic"
		>
			{{ $t('project.links.noLinks') }}
		</p>

		<TaskLinkItem
			v-for="link in links"
			:key="link.id"
			:link="link"
			:edit-enabled="editEnabled"
			@edit="startEdit"
			@delete="confirmDelete"
		/>

		<template v-if="editEnabled && formOpen">
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
					@custom-icon-cleared="pendingCustomIcon = null"
				/>
			</FormField>
			<div class="task-link-form-actions">
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
			v-else-if="editEnabled"
			icon="plus"
			variant="secondary"
			:shadow="false"
			class="mbs-2"
			@click="startCreate"
		>
			{{ $t('project.links.addLink') }}
		</XButton>

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
import {computed, reactive, ref, shallowReactive, watch} from 'vue'

import Icon from '@/components/misc/Icon'
import FormField from '@/components/input/FormField.vue'
import Modal from '@/components/misc/Modal.vue'
import IconPicker from '@/components/project/IconPicker.vue'
import TaskLinkItem from '@/components/tasks/partials/TaskLinkItem.vue'

import TaskLinkService from '@/services/taskLink'
import TaskLinkModel from '@/models/taskLink'
import type {ITaskLink} from '@/modelTypes/ITaskLink'
import type {ITask} from '@/modelTypes/ITask'
import {error} from '@/message'

const props = withDefaults(defineProps<{
	task: ITask
	editEnabled?: boolean
}>(), {
	editEnabled: true,
})

const emit = defineEmits<{
	'has-links': [value: boolean],
}>()

const taskLinkService = shallowReactive(new TaskLinkService())
const loading = computed(() => taskLinkService.loading)
const links = ref<ITaskLink[]>([])

async function loadLinks() {
	if (!props.task?.id) {
		links.value = []
		return
	}
	try {
		links.value = await taskLinkService.getAll(props.task.id)
		emit('has-links', links.value.length > 0)
	} catch (e) {
		error(e)
	}
}

watch(() => props.task?.id, loadLinks, {immediate: true})

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

function startEdit(link: ITaskLink) {
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
	if (!canSave.value || !props.task?.id) {
		return
	}

	saving.value = true
	try {
		let saved: ITaskLink
		if (editingId.value !== null) {
			const existing = links.value.find(l => l.id === editingId.value)
			saved = await taskLinkService.update(new TaskLinkModel({
				...existing,
				title: form.title,
				url: form.url,
				icon: form.icon,
			}))
		} else {
			saved = await taskLinkService.create(props.task.id, {
				title: form.title,
				url: form.url,
				icon: form.icon,
			})
		}

		if (pendingCustomIcon.value) {
			saved = await taskLinkService.uploadCustomIcon(saved, pendingCustomIcon.value)
		}

		if (editingId.value !== null) {
			const index = links.value.findIndex(l => l.id === editingId.value)
			if (index !== -1) {
				links.value.splice(index, 1, saved)
			}
		} else {
			links.value.push(saved)
		}

		emit('has-links', links.value.length > 0)
		closeForm()
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}

const linkToDelete = ref<ITaskLink | null>(null)

function confirmDelete(link: ITaskLink) {
	linkToDelete.value = link
}

async function doDelete() {
	if (!linkToDelete.value) {
		return
	}
	try {
		await taskLinkService.delete(linkToDelete.value)
		links.value = links.value.filter(l => l.id !== linkToDelete.value!.id)
		emit('has-links', links.value.length > 0)
	} catch (e) {
		error(e)
	} finally {
		linkToDelete.value = null
	}
}

defineExpose({
	startCreate,
})
</script>

<style lang="scss" scoped>
.task-links-empty {
	color: var(--grey-500);
	font-size: .85rem;
}

.task-link-form-actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .5rem;
}
</style>
