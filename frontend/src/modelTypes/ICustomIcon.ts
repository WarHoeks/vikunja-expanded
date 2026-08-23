import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface ICustomIcon extends IAbstract {
	id: number
	name: string
	createdBy: IUser
	created: Date
}
