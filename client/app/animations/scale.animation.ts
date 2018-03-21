import {animate, style, transition, trigger} from '@angular/animations';

export const scaleAnimation =
  trigger(
    'scaleAnimation',
    [
      transition(
        ':enter', [
          style({opacity: 0}),
          animate('200ms', style({opacity: 1, transform: 'scale(1)'}))
        ]
      ),
      transition(
        ':leave', [
          style({opacity: 1}),
          animate('200ms', style({opacity: 0, transform: 'scale(0.8)'}))
        ]
      )]
  );
