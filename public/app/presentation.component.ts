import { Component, OnInit } from '@angular/core';

@Component({
    selector: 'presentation',
    template: '<h1>{{Title}}</h1>' +
    '<h2>{{Speakers}}</h2>' +
    '<div>{{Votes}}</div>' +
    '<div>{{Description}}</div>'
})
export class PresentationComponent implements OnInit {
    Title = "Hello World!!!"
    Speakers = ["Jaacob Martin", "Jonatan Borkowski"]
    Votes = 3
    Description = "Assimilatios crescere, tanquam flavum medicina." +
        "Ubi est camerarius epos?Castus xiphias cito manifestums adiurator est." +
        "Tabes cantares, tanquam salvus bubo.Castus, brevis torquiss unus perdere de raptus, peritus rumor."

    constructor() {}
    ngOnInit() { }
}