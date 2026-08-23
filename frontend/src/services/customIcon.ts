import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import CustomIconModel from '@/models/customIcon'
import type {ICustomIcon} from '@/modelTypes/ICustomIcon'

// custom-icons only exists on /api/v2 (see AGENTS.md's API Version Policy), so this
// bypasses AbstractService and calls v2 directly, same pattern as services/projectLink.ts.
export default class CustomIconService {
	loading = false

	private setLoading() {
		const timeout = setTimeout(() => {
			this.loading = true
		}, 100)
		return () => {
			clearTimeout(timeout)
			this.loading = false
		}
	}

	async search(query: string): Promise<CustomIconModel[]> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().get(
				apiV2Url('custom-icons'),
				{params: {q: query, per_page: 30}},
			)
			return (data.items ?? []).map((i: Partial<ICustomIcon>) => new CustomIconModel(i))
		} finally {
			cancel()
		}
	}

	async upload(name: string, file: File): Promise<CustomIconModel> {
		const cancel = this.setLoading()
		try {
			const data = new FormData()
			data.append('name', name)
			data.append('file', file, file.name)
			const {data: response} = await AuthenticatedHTTPFactory().post(
				apiV2Url('custom-icons'),
				data,
			)
			return new CustomIconModel(response)
		} finally {
			cancel()
		}
	}

	async delete(icon: ICustomIcon): Promise<void> {
		const cancel = this.setLoading()
		try {
			await AuthenticatedHTTPFactory().delete(apiV2Url(`custom-icons/${Number(icon.id)}`))
		} finally {
			cancel()
		}
	}

	/**
	 * Fetches a library icon's image as an object URL. The download route requires
	 * auth, so a plain <img src> can't be used — mirrors AbstractService.getBlobUrl.
	 */
	async getBlobUrl(icon: Pick<ICustomIcon, 'id'>): Promise<string> {
		const response = await AuthenticatedHTTPFactory()({
			url: apiV2Url(`custom-icons/${Number(icon.id)}/icon`),
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(response.data)
	}
}
