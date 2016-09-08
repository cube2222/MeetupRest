import React from 'react';
import {render} from 'react-dom';
import { Router, Route, hashHistory } from 'react-router'
import Example from './modules/Example';


render((
  <Router history={hashHistory}>
    <Route path="/" component={Example}/>
  </Router>
), document.getElementById('app'))


