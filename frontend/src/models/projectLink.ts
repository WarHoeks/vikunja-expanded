import AbstractModel from '@/models/abstractModel'
import type {IProjectLink} from '@/modelTypes/IProjectLink'
import UserModel from '@/models/user'

export default class ProjectLinkModel extends AbstractModel<IProjectLink> implements IProjectLink {
	id = 0
	projectId = 0
	url = ''
	title = ''
	icon = ''
	customIconId = 0
	createdBy = null

	created: Date
	updated: Date

	constructor(data: Partial<IProjectLink> = {}) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
