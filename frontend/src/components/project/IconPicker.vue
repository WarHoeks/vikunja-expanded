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
						class="icon-picker-svg"
						v-html="svgCache[option.slug]"
					/>
					{{ option.title }}
				</span>
			</template>
		</Multiselect>
		<div
			v-if="customFileName"
			class="icon-picker-custom-selected"
		>
			<img
				:src="customFilePreviewUrl"
				alt=""
				class="icon-picker-custom-preview"
			>
			<span>{{ $t('project.links.customIconSelected', {filename: customFileName}) }}</span>
			<BaseButton
				:aria-label="$t('project.links.removeSelectedIcon')"
				@click="clearCustomFile"
			>
				<Icon icon="times" />
			</BaseButton>
		</div>
		<label
			v-else
			class="icon-picker-upload"
		>
			<input
				ref="fileInputRef"
				type="file"
				accept="image/*"
				class="is-hidden"
				@change="onCustomFileChange"
			>
			{{ $t('project.links.uploadCustomIcon') }}
		</label>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, reactive, ref, watch} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import Multiselect from '@/components/input/Multiselect.vue'
import {ALL_SIMPLE_ICONS, getSimpleIconSvg, type SimpleIcon} from '@/helpers/simpleIcons'

const MAX_RESULTS = 30

const props = defineProps<{
	modelValue: string
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'customIconSelected': [file: File],
	'customIconCleared': [],
}>()

const selected = computed(() => ALL_SIMPLE_ICONS.find(i => i.slug === props.modelValue) ?? null)

const searchResults = ref<SimpleIcon[]>(ALL_SIMPLE_ICONS.slice(0, MAX_RESULTS))

// Populated lazily as results/selection change; the template reads straight from
// this cache so a v-for row never calls a reactivity-creating composable per render.
const svgCache = reactive<Record<string, string>>({})

function ensureSvgLoaded(slug: string) {
	if (slug === '' || svgCache[slug] !== undefined) {
		return
	}
	svgCache[slug] = ''
	getSimpleIconSvg(slug).then(svg => {
		svgCache[slug] = svg
	})
}

watch(searchResults, results => results.forEach(i => ensureSvgLoaded(i.slug)), {immediate: true})
watch(selected, icon => icon && ensureSvgLoaded(icon.slug), {immediate: true})

function onSearch(query: string) {
	const q = query.trim().toLowerCase()
	if (q === '') {
		searchResults.value = ALL_SIMPLE_ICONS.slice(0, MAX_RESULTS)
		return
	}
	searchResults.value = ALL_SIMPLE_ICONS
		.filter(i => i.title.toLowerCase().includes(q) || i.slug.includes(q))
		.slice(0, MAX_RESULTS)
}

function onSelect(icon: SimpleIcon) {
	emit('update:modelValue', icon.slug)
}

function onUpdate(value: SimpleIcon | SimpleIcon[] | null) {
	if (value === null) {
		emit('update:modelValue', '')
	}
}

const fileInputRef = ref<HTMLInputElement | null>(null)
const customFileName = ref('')
const customFilePreviewUrl = ref('')

function onCustomFileChange(e: Event) {
	const file = (e.target as HTMLInputElement).files?.[0]
	if (!file) {
		return
	}

	if (customFilePreviewUrl.value) {
		URL.revokeObjectURL(customFilePreviewUrl.value)
	}
	customFileName.value = file.name
	customFilePreviewUrl.value = URL.createObjectURL(file)

	emit('customIconSelected', file)
}

function clearCustomFile() {
	if (customFilePreviewUrl.value) {
		URL.revokeObjectURL(customFilePreviewUrl.value)
	}
	customFileName.value = ''
	customFilePreviewUrl.value = ''
	if (fileInputRef.value) {
		fileInputRef.value.value = ''
	}
	emit('customIconCleared')
}

onBeforeUnmount(() => {
	if (customFilePreviewUrl.value) {
		URL.revokeObjectURL(customFilePreviewUrl.value)
	}
})
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

.icon-picker-upload {
	display: inline-block;
	margin-block-start: .5rem;
	font-size: .85rem;
	color: var(--primary);
	cursor: pointer;

	&:hover {
		text-decoration: underline;
	}
}

.icon-picker-custom-selected {
	display: flex;
	align-items: center;
	gap: .5rem;
	margin-block-start: .5rem;
	font-size: .85rem;
	color: var(--grey-600);
}

.icon-picker-custom-preview {
	inline-size: 1.5rem;
	block-size: 1.5rem;
	object-fit: contain;
	border-radius: 2px;
}
</style>
