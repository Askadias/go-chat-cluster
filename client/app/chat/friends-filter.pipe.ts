import {Pipe, PipeTransform} from '@angular/core';
import {User} from '../domain/user';

@Pipe({
  name: 'friendsFilter'
})
export class FriendsFilterPipe implements PipeTransform {

  transform(items: User[], filter?: string): any {
    if (!items || !filter) {
      return items;
    }
    return items.filter(item => (item.name.toLowerCase().indexOf(filter.toLowerCase()) !== -1));
  }

}
