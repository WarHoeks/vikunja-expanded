import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface IProjectLink extends IAbstract {
	id: number
	projectId: number
	url: string
	title: string
	icon: string
	customIconId: number
	createdBy: IUser
	created: Date
	updated: Date
}
