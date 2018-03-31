import {Pipe, PipeTransform} from '@angular/core';

@Pipe({
  name: 'skipOwnFilter'
})
export class SkipOwnFilterPipe implements PipeTransform {

  transform(items: string[], myId?: string): any {
    if (!items || !myId) {
      return items;
    }
    return items.filter(userId => userId !== myId);
  }

}
