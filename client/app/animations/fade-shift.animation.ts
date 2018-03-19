import {animate, style, transition, trigger} from '@angular/animations';

export const fadeShiftAnimation =
  trigger(
    'fadeShiftAnimation',
    [
      transition(
        ':enter', [
          style({transform: 'translateX(100%)', opacity: 0}),
          animate('100ms', style({transform: 'translateX(0)', opacity: 1}))
        ]
      ),
      transition(
        ':leave', [
          style({transform: 'translateX(0)', opacity: 1}),
          animate('100ms', style({transform: 'translateX(100%)', opacity: 0}))
        ]
      )]
  );
