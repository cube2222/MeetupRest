import { Component } from '@angular/core';
import { PresentationComponent } from './presentation.component'

@Component({
  selector: 'my-app',
  template: '<h1>{{title}}</h1>' +
  '<div><presentation></presentation></div>'
})
export class AppComponent {
    title = "Meetuprest"
    presentation: PresentationComponent
}