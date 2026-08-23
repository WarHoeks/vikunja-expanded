import AbstractModel from '@/models/abstractModel'
import type {ITaskLink} from '@/modelTypes/ITaskLink'
import UserModel from '@/models/user'

export default class TaskLinkModel extends AbstractModel<ITaskLink> implements ITaskLink {
	id = 0
	taskId = 0
	url = ''
	title = ''
	createdBy = null

	created: Date
	updated: Date

	constructor(data: Partial<ITaskLink> = {}) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
