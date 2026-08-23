import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {objectToSnakeCase} from '@/helpers/case'
import TaskLinkModel from '@/models/taskLink'
import type {ITaskLink} from '@/modelTypes/ITaskLink'

// task_links only exists on /api/v2 (see AGENTS.md's API Version Policy), so this
// bypasses AbstractService (which pins to /api/v1 and uses v1's inverted
// PUT-create/POST-update verbs) and calls v2 directly. Mirrors services/projectLink.ts.
export default class TaskLinkService {
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

	async getAll(taskId: ITaskLink['taskId']): Promise<TaskLinkModel[]> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().get(
				apiV2Url(`tasks/${Number(taskId)}/links`),
				{params: {per_page: 200}},
			)
			return (data.items ?? []).map((l: Partial<ITaskLink>) => new TaskLinkModel(l))
		} finally {
			cancel()
		}
	}

	async create(taskId: ITaskLink['taskId'], link: Partial<ITaskLink>): Promise<TaskLinkModel> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().post(
				apiV2Url(`tasks/${Number(taskId)}/links`),
				objectToSnakeCase(link),
			)
			return new TaskLinkModel(data)
		} finally {
			cancel()
		}
	}

	async update(link: ITaskLink): Promise<TaskLinkModel> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().put(
				apiV2Url(`tasks/${Number(link.taskId)}/links/${Number(link.id)}`),
				objectToSnakeCase(link),
			)
			return new TaskLinkModel(data)
		} finally {
			cancel()
		}
	}

	async delete(link: ITaskLink): Promise<void> {
		const cancel = this.setLoading()
		try {
			await AuthenticatedHTTPFactory().delete(
				apiV2Url(`tasks/${Number(link.taskId)}/links/${Number(link.id)}`),
			)
		} finally {
			cancel()
		}
	}

	async attachCustomIconFromLibrary(link: ITaskLink, customIconId: number): Promise<TaskLinkModel> {
		const cancel = this.setLoading()
		try {
			const {data} = await AuthenticatedHTTPFactory().post(
				apiV2Url(`tasks/${Number(link.taskId)}/links/${Number(link.id)}/icon/library/${Number(customIconId)}`),
			)
			return new TaskLinkModel(data)
		} finally {
			cancel()
		}
	}

	async uploadCustomIcon(link: ITaskLink, file: File): Promise<TaskLinkModel> {
		const cancel = this.setLoading()
		try {
			const data = new FormData()
			data.append('icon', file, file.name)
			const {data: response} = await AuthenticatedHTTPFactory().post(
				apiV2Url(`tasks/${Number(link.taskId)}/links/${Number(link.id)}/icon`),
				data,
			)
			return new TaskLinkModel(response)
		} finally {
			cancel()
		}
	}

	/**
	 * Fetches a link's custom icon as an object URL. The download route requires auth,
	 * so a plain <img src> can't be used — mirrors AbstractService.getBlobUrl.
	 */
	async getCustomIconBlobUrl(link: ITaskLink): Promise<string> {
		const response = await AuthenticatedHTTPFactory()({
			url: apiV2Url(`tasks/${Number(link.taskId)}/links/${Number(link.id)}/icon`),
			method: 'GET',
			responseType: 'blob',
		})
		return window.URL.createObjectURL(response.data)
	}
}
