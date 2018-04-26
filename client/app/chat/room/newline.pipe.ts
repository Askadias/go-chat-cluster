import {Pipe, PipeTransform} from '@angular/core';

@Pipe({name: 'newline'})
export class NewlinePipe implements PipeTransform {



  constructor() {
  }

  transform(value: string, args: string[]): any {
    return value.replace(/(?:\r\n|\r|\n)/g, '<br/>');
  }

}
