import {NgModule}      from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {AppComponent}   from './app.component';
import {PresentationComponent} from './presentation.component'

@NgModule({
    imports: [BrowserModule],
    declarations: [
        AppComponent,
        PresentationComponent
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
}