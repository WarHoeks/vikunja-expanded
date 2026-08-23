<!-- eslint-disable vue/no-v-html -->
<template>
	<div class="icon-picker">
		<Multiselect
			:model-value="selected"
			:search-results="searchResults"
			label="title"
			:placeholder="$t('project.links.iconPlaceholder')"
			@search="onSearch"
			@select="onSelect"
			@update:model-value="onUpdate"
		>
			<template #searchResult="{option}">
				<span class="icon-picker-result">
					<span
						v-if="option.kind === 'simple'"
						class="icon-picker-svg"
						v-html="svgCache[option.slug]"
					/>
					<img
						v-else
						class="icon-picker-custom-thumb"
						:src="customIconThumbCache[option.id] ?? ''"
						alt=""
					>
					{{ option.title }}
					<span
						v-if="option.kind === 'custom'"
						class="icon-picker-custom-badge"
					>{{ $t('project.links.customIconBadge') }}</span>
				</span>
			</template>
		</Multiselect>

		<div
			v-if="uploadOpen"
			class="icon-picker-upload-form"
		>
			<input
				v-model="uploadName"
				type="text"
				class="input"
				:placeholder="$t('project.links.customIconNamePlaceholder')"
			>
			<input
				ref="fileInputRef"
				type="file"
				accept="image/*"
				@change="onFileChosen"
			>
			<div class="icon-picker-upload-form-actions">
				<XButton
					:disabled="!canUpload"
					:loading="uploading"
					variant="secondary"
					:shadow="false"
					@click="doUpload"
				>
					{{ $t('project.links.uploadButton') }}
				</XButton>
				<XButton
					variant="tertiary"
					:shadow="false"
					@click="uploadOpen = false"
				>
					{{ $t('project.links.cancel') }}
				</XButton>
			</div>
		</div>
		<BaseButton
			v-else
			class="icon-picker-upload-toggle"
			@click="uploadOpen = true"
		>
			{{ $t('project.links.uploadCustomIcon') }}
		</BaseButton>
	</div>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import {ALL_SIMPLE_ICONS, getSimpleIconSvg, type SimpleIcon} from '@/helpers/simpleIcons'
import CustomIconService from '@/services/customIcon'
import {error} from '@/message'

const MAX_RESULTS = 20

interface SimpleIconResult extends SimpleIcon {
	kind: 'simple'
}

interface CustomIconResult {
	kind: 'custom'
	id: number
	title: string
}

type IconResult = SimpleIconResult | CustomIconResult

const props = defineProps<{
	modelValue: string
	customIconId?: number
	customIconName?: string
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'update:customIconId': [value: number],
	'update:customIconName': [value: string],
}>()

const customIconService = new CustomIconService()

const selected = computed<IconResult | null>(() => {
	if (props.customIconId) {
		return {kind: 'custom', id: props.customIconId, title: props.customIconName ?? ''}
	}
	const icon = ALL_SIMPLE_ICONS.find(i => i.slug === props.modelValue)
	return icon ? {kind: 'simple', ...icon} : null
})

const searchResults = ref<IconResult[]>(ALL_SIMPLE_ICONS.slice(0, MAX_RESULTS).map(i => ({kind: 'simple', ...i})))

// Populated lazily as results/selection change; the template reads straight from
// this cache so a v-for row never calls a reactivity-creating composable per render.
const svgCache = reactive<Record<string, string>>({})
const customIconThumbCache = reactive<Record<number, string>>({})

function ensureSvgLoaded(slug: string) {
	if (slug === '' || svgCache[slug] !== undefined) {
		return
	}
	svgCache[slug] = ''
	getSimpleIconSvg(slug).then(svg => {
		svgCache[slug] = svg
	})
}

function ensureCustomThumbLoaded(icon: CustomIconResult) {
	if (customIconThumbCache[icon.id] !== undefined) {
		return
	}
	customIconThumbCache[icon.id] = ''
	customIconService.getBlobUrl({id: icon.id}).then(url => {
		customIconThumbCache[icon.id] = url
	})
}

function loadThumbsFor(results: IconResult[]) {
	for (const r of results) {
		if (r.kind === 'simple') {
			ensureSvgLoaded(r.slug)
		} else {
			ensureCustomThumbLoaded(r)
		}
	}
}

watch(searchResults, loadThumbsFor, {immediate: true})
watch(selected, icon => icon && loadThumbsFor([icon]), {immediate: true})

let searchToken = 0

async function onSearch(query: string) {
	const token = ++searchToken
	const q = query.trim().toLowerCase()

	const simpleResults: IconResult[] = (q === ''
		? ALL_SIMPLE_ICONS.slice(0, MAX_RESULTS)
		: ALL_SIMPLE_ICONS.filter(i => i.title.toLowerCase().includes(q) || i.slug.includes(q)).slice(0, MAX_RESULTS)
	).map(i => ({kind: 'simple', ...i}))

	// Show simple-icons results immediately; the library search is a network call.
	searchResults.value = simpleResults

	try {
		const customResults = await customIconService.search(q)
		if (token !== searchToken) {
			return // a newer search superseded this one
		}
		searchResults.value = [
			...simpleResults,
			...customResults.map(c => ({kind: 'custom', id: c.id, title: c.name} as IconResult)),
		]
	} catch (e) {
		error(e)
	}
}

function selectResult(icon: IconResult) {
	if (icon.kind === 'simple') {
		emit('update:customIconId', 0)
		emit('update:modelValue', icon.slug)
	} else {
		emit('update:modelValue', '')
		emit('update:customIconId', icon.id)
		emit('update:customIconName', icon.title)
	}
}

function onSelect(icon: IconResult) {
	selectResult(icon)
}

function onUpdate(value: IconResult | IconResult[] | null) {
	if (value === null) {
		emit('update:modelValue', '')
		emit('update:customIconId', 0)
	}
}

const uploadOpen = ref(false)
const uploadName = ref('')
const uploading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const pendingFile = ref<File | null>(null)

const canUpload = computed(() => uploadName.value.trim() !== '' && pendingFile.value !== null)

function onFileChosen(e: Event) {
	pendingFile.value = (e.target as HTMLInputElement).files?.[0] ?? null
}

async function doUpload() {
	if (!canUpload.value || !pendingFile.value) {
		return
	}
	uploading.value = true
	try {
		const icon = await customIconService.upload(uploadName.value.trim(), pendingFile.value)
		selectResult({kind: 'custom', id: icon.id, title: icon.name})
		uploadOpen.value = false
		uploadName.value = ''
		pendingFile.value = null
		if (fileInputRef.value) {
			fileInputRef.value.value = ''
		}
	} catch (e) {
		error(e)
	} finally {
		uploading.value = false
	}
}
</script>

<style lang="scss" scoped>
.icon-picker-result {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.icon-picker-svg {
	inline-size: 1.1rem;
	block-size: 1.1rem;
	flex-shrink: 0;

	:deep(svg) {
		inline-size: 100%;
		block-size: 100%;
		fill: currentColor;
	}
}

.icon-picker-custom-thumb {
	inline-size: 1.1rem;
	block-size: 1.1rem;
	flex-shrink: 0;
	object-fit: contain;
}

.icon-picker-custom-badge {
	margin-inline-start: auto;
	font-size: .7rem;
	color: var(--grey-500);
}

.icon-picker-upload-toggle {
	display: inline-block;
	margin-block-start: .5rem;
	font-size: .85rem;
	color: var(--primary);
	cursor: pointer;

	&:hover {
		text-decoration: underline;
	}
}

.icon-picker-upload-form {
	margin-block-start: .5rem;
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.icon-picker-upload-form-actions {
	display: flex;
	gap: .5rem;
}
</style>
