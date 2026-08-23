import {ref, watchEffect, type Ref} from 'vue'
import iconsData from 'simple-icons/icons.json'

export interface SimpleIcon {
	title: string
	slug: string
	hex: string
}

export const ALL_SIMPLE_ICONS = iconsData as SimpleIcon[]

// Lazy, per-icon chunks: only icons actually rendered get their SVG fetched, keeping
// the ~3200-icon set out of the main bundle. Keys are the literal relative-path
// strings Vite resolves this glob against, so getSimpleIconSvg() must build the
// same string to look one up.
const svgModules = import.meta.glob<string>(
	'../../node_modules/simple-icons/icons/*.svg',
	{query: '?raw', import: 'default'},
)

const svgCache = new Map<string, Promise<string>>()

export function getSimpleIconSvg(slug: string): Promise<string> {
	if (slug === '') {
		return Promise.resolve('')
	}

	let cached = svgCache.get(slug)
	if (!cached) {
		const key = `../../node_modules/simple-icons/icons/${slug}.svg`
		const loader = svgModules[key]
		cached = loader ? loader() : Promise.resolve('')
		svgCache.set(slug, cached)
	}
	return cached
}

/**
 * Reactively resolves a simple-icons slug to its raw SVG markup, refetching when the slug changes.
 */
export function useSimpleIconSvg(slug: Ref<string> | string): Ref<string> {
	const svg = ref('')
	watchEffect(() => {
		const s = typeof slug === 'string' ? slug : slug.value
		getSimpleIconSvg(s).then(result => {
			svg.value = result
		})
	})
	return svg
}
