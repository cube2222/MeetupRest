import React from 'react';
import ReactDOM from 'react-dom';
import Formsy from 'formsy-react';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import RaisedButton from 'material-ui/RaisedButton';
import Paper from 'material-ui/Paper';
import MenuItem from 'material-ui/MenuItem';
import { FormsyText } from 'formsy-material-ui/lib';


const UpdatePresentation = React.createClass({
    getInitialState() {
        return {
            canSubmit: false,
            data: {},
            speakersText: '',
        };
    },

    componentDidMount: function() {
        this.getPresentation()
    },

    contextTypes: {
        router: React.PropTypes.object
    },

    errorMessages: {
        wordsError: "Please only use letters",
    },

    styles: {
        paperStyle: {
            width: 300,
            margin: 'auto',
            padding: 20,
        },
        switchStyle: {
            marginBottom: 16,
        },
        submitStyle: {
            marginTop: 32,
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
            url: '../../presentation/' + this.props.params.presentationId + '/update' ,
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

    getPresentation() {
        $.ajax({
            url: '/presentation/' + this.props.params.presentationId + '/',
            dataType: 'json',
            cache: false,
            success: function(data) {
                        let speakersString = data.Speakers.map(function (i) {
                            return i.Name;
                        }).join(', ');

                        this.setState({
                            data: data,
                            speakersText: speakersString,
                        });
                    
                    }.bind(this),
            error: function(xhr, status, err) {
                        console.error(this.props.url, status, err.toString());
                    }.bind(this)
      });
    },

    notifyFormError(data) {
        console.error('Form error:', data);
    },

    render() {
        let {paperStyle, submitStyle } = this.styles;
        let { wordsError } = this.errorMessages;

        return (
            <MuiThemeProvider muiTheme={getMuiTheme()}>
                <Paper style={paperStyle}>
                    <h3>Forms for updating Presentation </h3>
                    <Formsy.Form
                      onValid={this.enableSubmitButton}
                      onInvalid={this.disableSubmitButton}
                      onValidSubmit={this.submitForm}
                      onInvalidSubmit={this.notifyFormError}
                    >
                        <FormsyText
                          name="Title"
                          validations="isWords"
                          validationError={wordsError}
                          value={this.state.data.Title}
                          required
                          hintText="What is presentation title?"
                          floatingLabelText="Title"
                        />

                        <FormsyText
                          name="Description"
                          value={this.state.data.Description}
                          required
                          multiLine={true}
                          rows={2}
                          rowsMax={10}
                          hintText="Write a few sentences what the subject is presentation"
                          floatingLabelText="Descryption"
                        />
                        <FormsyText
                          name="Speakers"
                          multiLine={true}
                          value={this.state.speakersText}
                          rows={2}
                          rowsMax={10}
                          required
                          hintText="Name and Surname of presentation speakers"
                          floatingLabelText="Speakers"
                        />

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

export default UpdatePresentation;