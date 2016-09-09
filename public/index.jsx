import React from 'react';
import {render} from 'react-dom';
import { Router, Route, hashHistory } from 'react-router';
import injectTapEventPlugin from 'react-tap-event-plugin';
import AddPresentation from './modules/AddPresentation';
import AddMeetup from './modules/AddMeetup';
import AddSpeaker from './modules/AddSpeaker';
import UpdatePresentation from './modules/UpdatePresentation';
import UpdateMeetup from './modules/UpdateMeetup';
import UpdateSpeaker from './modules/UpdateSpeaker';
import Main from './modules/Main'

injectTapEventPlugin();

render((
  <Router history={hashHistory}>
    <Route path="/add_presentation" component={AddPresentation}/>
    <Route path="/add_meetup" component={AddMeetup}/>
    <Route path="/add_speaker" component={AddSpeaker}/>
    <Route path="/update_presentation/:presentationId" component={UpdatePresentation}/>
    <Route path="/update_meetup/:meetupId" component={UpdateMeetup}/>
    <Route path="/update_speaker/:speakerId" component={UpdateSpeaker}/>    
    <Route path="/main" component={Main}/>    
  </Router>
), document.getElementById('app'))


