import React from 'react';
import {render} from 'react-dom';
import RaisedButton from 'material-ui/RaisedButton';
import baseTheme from 'material-ui/styles/baseThemes/lightBaseTheme';
import getMuiTheme from 'material-ui/styles/getMuiTheme';

class App extends React.Component {
  getChildContext() {
    return { muiTheme: getMuiTheme(baseTheme) };
  }
  render () {
    return    <RaisedButton label="START" />;
  }
}

App.childContextTypes = {
    muiTheme: React.PropTypes.object.isRequired,
};

render(<App/>, document.getElementById('app'));
