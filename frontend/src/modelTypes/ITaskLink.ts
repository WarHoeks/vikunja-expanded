import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface ITaskLink extends IAbstract {
	id: number
	taskId: number
	url: string
	title: string
	createdBy: IUser
	created: Date
	updated: Date
}
