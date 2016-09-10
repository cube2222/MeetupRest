import React from 'react';
import ReactDOM from 'react-dom';
import Formsy from 'formsy-react';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import RaisedButton from 'material-ui/RaisedButton';
import GridList from 'material-ui/GridList'
import Paper from 'material-ui/Paper';
import MenuItem from 'material-ui/MenuItem';
import TextField from 'material-ui/TextField';
import { FormsyDate, FormsyText, FormsyTime } from 'formsy-material-ui/lib';
import {Gmaps, Marker, InfoWindow, Circle} from 'react-gmaps';
import InlineBlock from 'react-inline-block';
import Geosuggest from 'react-geosuggest';


const AddMeetup = React.createClass({
    getInitialState() {
        return {
            canSubmit: false,
            lat: 51.5258541,
            lng: -0.08040660000006028,
        };
    },

    contextTypes: {
        router: React.PropTypes.object
    },

    errorMessages: {
        wordsError: "Please only use letters",
    },

    styles: {
        paperStyle: {
            width: 670,
            margin: 'auto',
            padding: 20,
        },
        switchStyle: {
            marginBottom: 16,
        },
        submitStyle: {
            marginTop: 32,
        },
        block: {
            display: 'flex',
        },
        block2: {
            margin: 20,
        },
        h3: {
            marginTop: 20,
            marginLeft: 20,
            fontWeight: 400,
        },
    },

    enableSubmitButton() {
        this.setState({
            canSubmit: true,
        });
    },

    disableSubmitButton() {
        this.setState({
            canSubmit: false,
        });
    },

    submitForm(data) {
        $.ajax({
            url: '../../meetup/' ,
            contentType: 'application/json; charset=utf-8',
            dataType: 'json',
            type: 'POST',
            data: JSON.stringify(data, null, 4),
            success: function (data) {
                        //TODO: add correct redirect url
                        this.context.router.push('/main')
                    }.bind(this),
            error: function(xhr, status, err) {
                        switch (xhr.status) {
                            case 403:
                            window.location.href = xhr.responseText
                            //this.context.router.push(xhr.responseText);
                            break;
                        }
                        console.error(this.props.url, status, err.toString());
                    }.bind(this)
        });
    },

    notifyFormError(data) {
        console.error('Form error:', data);
    },

    onMapCreated(map) {
        map.setOptions({
            disableDefaultUI: true
        });
    },

    /**
     * When a suggest got selected
     * @param  {Object} suggest The suggest
     */
    onSuggestSelect: function(suggest) {
        this.setState({
            lat: suggest.location.lat,
            lng: suggest.location.lng,
        });
        console.log(suggest); // eslint-disable-line
    },

    /**
     * When there are no suggest results
     * @param {String} userInput The user input
     */
    onSuggestNoResults: function(userInput) {
        console.log('onSuggestNoResults for :' + userInput); // eslint-disable-line
    },

    render() {
        let { paperStyle, submitStyle, block, block2, h3 } = this.styles;
        let { wordsError } = this.errorMessages;
        var fixtures = [
            {label: 'New York', location: {lat: 40.7033127, lng: -73.979681}},
            {label: 'Rio', location: {lat: -22.066452, lng: -42.9232368}},
            {label: 'Tokyo', location: {lat: 35.673343, lng: 139.710388}}
        ];

        return (
            <MuiThemeProvider muiTheme={getMuiTheme()}>
                <Paper style={paperStyle} zDepth={1}>
                    <h3 style={h3}>Adding meetup form</h3>
                    
                    <Formsy.Form
                      onValid={this.enableSubmitButton}
                      onInvalid={this.disableSubmitButton}
                      onValidSubmit={this.submitForm}
                      onInvalidSubmit={this.notifyFormError}
                    >
                    <div style={block}>
                        <div style={block2}>
                            <FormsyText
                                name="Title"
                                validations="isWords"
                                validationError={wordsError}
                                required
                                hintText="What is meetup title?"
                                floatingLabelText="Title"
                            />
                            <FormsyDate
                                name="date"
                                required
                                floatingLabelText="Date"
                            />
                            <FormsyDate
                                name="voteTimeEnd"
                                required
                                floatingLabelText="Vote Time End"
                            />
                            <FormsyText
                                name="Description"
                                required
                                multiLine={true}
                                rows={4}
                                rowsMax={10}
                                hintText="Description for meetup.com"
                                floatingLabelText="Descryption"
                            />   
                    
                        </div>
                        <div style={block2}>
                            <Geosuggest 
                                onChange={this.onChange}
                                onSuggestSelect={this.onSuggestSelect}
                                onSuggestNoResults={this.onSuggestNoResults}
                                location={new google.maps.LatLng(this.state.lat, this.state.lng)}
                                radius="20" />
                            <Gmaps
                                width={'300px'}
                                height={'300px'}
                                lat={this.state.lat}
                                lng={this.state.lng}
                                zoom={16}
                                loadingMessage={'Be happy'}
                                params={{v: '3.exp', key: 'AIzaSyByTOp78icwH2oRmfcC9zTarst10suM42I'}}
                                onMapCreated={this.onMapCreated}>
                                <Marker
                                lat={this.state.lat}
                                lng={this.state.lng}
                                draggable={true}
                                onDragEnd={this.onDragEnd} />
                                
                            </Gmaps>
                        </div> 
                    </div>   
                        <RaisedButton
                          style={submitStyle}
                          type="submit"
                          label="Submit"
                          disabled={!this.state.canSubmit}
                        /> 
                    </Formsy.Form>
                </Paper>
            </MuiThemeProvider>
        );
    },
});

export default AddMeetup;




