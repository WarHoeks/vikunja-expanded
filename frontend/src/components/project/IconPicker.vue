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
		<label class="icon-picker-upload">
			<input
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
import {computed, reactive, ref, watch} from 'vue'

import Multiselect from '@/components/input/Multiselect.vue'
import {ALL_SIMPLE_ICONS, getSimpleIconSvg, type SimpleIcon} from '@/helpers/simpleIcons'

const MAX_RESULTS = 30

const props = defineProps<{
	modelValue: string
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'customIconSelected': [file: File],
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

function onCustomFileChange(e: Event) {
	const file = (e.target as HTMLInputElement).files?.[0]
	if (file) {
		emit('customIconSelected', file)
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
</style>
