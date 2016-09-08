import React from 'react';
import {render} from 'react-dom';
import RaisedButton from 'material-ui/RaisedButton';
import baseTheme from 'material-ui/styles/baseThemes/lightBaseTheme';
import getMuiTheme from 'material-ui/styles/getMuiTheme';

export default class Example extends React.Component {

    getChildContext() {
        return { muiTheme: getMuiTheme(baseTheme) };
    }

    render() {
        return (
            <RaisedButton label="START" />
        );
    }        
}

Example.childContextTypes = {
    muiTheme: React.PropTypes.object.isRequired,
};

// export default React.createClass({
//   contextTypes: {
//         muiTheme: React.PropTypes.string.isRequired,
//   },
//   getChildContext() {
//     return { muiTheme: getMuiTheme(baseTheme) };
//   },
//   render () {
//     return <RaisedButton label="START" />;
//   }
// });
