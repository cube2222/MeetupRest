import React from 'react';
import {render} from 'react-dom';
import { Router, Route, hashHistory } from 'react-router';
import Example from './modules/Example';
import AddPresentation from './modules/AddPresentation';
import Main from './modules/Main'


render((
  <Router history={hashHistory}>
    <Route path="/" component={Example}/>
    <Route path="/add_presentation" component={AddPresentation}/>
    <Route path="/main" component={Main}/>    
  </Router>
), document.getElementById('app'))


