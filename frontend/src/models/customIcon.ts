import AbstractModel from '@/models/abstractModel'
import type {ICustomIcon} from '@/modelTypes/ICustomIcon'
import UserModel from '@/models/user'

export default class CustomIconModel extends AbstractModel<ICustomIcon> implements ICustomIcon {
	id = 0
	name = ''
	createdBy = null

	created: Date

	constructor(data: Partial<ICustomIcon> = {}) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
	}
}
