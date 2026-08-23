import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {objectToSnakeCase} from '@/helpers/case'
import ProjectLinkModel from '@/models/projectLink'
import type {IProjectLink} from '@/modelTypes/IProjectLink'

// project_links only exists on /api/v2 (see AGENTS.md's API Version Policy), so
// this bypasses AbstractService (which pins to /api/v1 and uses v1's inverted
// PUT-create/POST-update verbs) and calls v2 directly, same pattern as the
// bulk-create calls in services/task.ts.
export default class ProjectLinkService {
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

	async getAll(projectId: IProjectLink['projectId']): Promise<ProjectLinkModel[]> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().get(
				apiV2Url(`projects/${Number(projectId)}/links`),
				{params: {per_page: 200}},
			)
			return (data.items ?? []).map((l: Partial<IProjectLink>) => new ProjectLinkModel(l))
		} finally {
			cancel()
		}
	}

	async create(projectId: IProjectLink['projectId'], link: Partial<IProjectLink>): Promise<ProjectLinkModel> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().post(
				apiV2Url(`projects/${Number(projectId)}/links`),
				objectToSnakeCase(link),
			)
			return new ProjectLinkModel(data)
		} finally {
			cancel()
		}
	}

	async update(link: IProjectLink): Promise<ProjectLinkModel> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().put(
				apiV2Url(`projects/${Number(link.projectId)}/links/${Number(link.id)}`),
				objectToSnakeCase(link),
			)
			return new ProjectLinkModel(data)
		} finally {
			cancel()
		}
	}

	async delete(link: IProjectLink): Promise<void> {
		const cancel = this.setLoading()
		try {
			await AuthenticatedHTTPFactory().delete(
				apiV2Url(`projects/${Number(link.projectId)}/links/${Number(link.id)}`),
			)
		} finally {
			cancel()
		}
	}

	async uploadCustomIcon(link: IProjectLink, file: File): Promise<ProjectLinkModel> {
		const cancel = this.setLoading()
		try {
			const data = new FormData()
			data.append('icon', file, file.name)
			const {data: response} = await AuthenticatedHTTPFactory().post(
				apiV2Url(`projects/${Number(link.projectId)}/links/${Number(link.id)}/icon`),
				data,
			)
			return new ProjectLinkModel(response)
		} finally {
			cancel()
		}
	}

	/**
	 * Fetches a link's custom icon as an object URL. The download route requires auth,
	 * so a plain <img src> can't be used — mirrors AbstractService.getBlobUrl.
	 */
	async getCustomIconBlobUrl(link: IProjectLink): Promise<string> {
		const response = await AuthenticatedHTTPFactory()({
			url: apiV2Url(`projects/${Number(link.projectId)}/links/${Number(link.id)}/icon`),
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(response.data)
	}
}
