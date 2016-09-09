import React from 'react';
import ReactDOM from 'react-dom';
import Formsy from 'formsy-react';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import RaisedButton from 'material-ui/RaisedButton';
import TextField from 'material-ui/TextField';
import Paper from 'material-ui/Paper';
import MenuItem from 'material-ui/MenuItem';
import { FormsyDate, FormsyText, FormsyTime } from 'formsy-material-ui/lib';


const UpdateSpeaker = React.createClass({
    getInitialState() {
        return {
            canSubmit: false,
            data: {},
        };
    },

    componentDidMount: function() {
        this.getSpeaker()
    },

    contextTypes: {
        router: React.PropTypes.object
    },

    errorMessages: {
        wordsError: "Please only use letters",
        urlError: "Please provide a valid URL",
        emailError: "Please provide a valid email",
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
            url: '../../speaker/' + this.props.params.speakerId + '/update' ,
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

    getSpeaker() {
        $.ajax({
            url: '/speaker/' + this.props.params.speakerId + '/',
            dataType: 'json',
            cache: false,
            success: function(data) {
                        this.setState({
                            data: data,
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
        let { wordsError, urlError, emailError } = this.errorMessages;

        return (
            <MuiThemeProvider muiTheme={getMuiTheme()}>
                <Paper style={paperStyle}>
                    <h3>Forms for updating Speaker </h3>
                    <Formsy.Form
                      onValid={this.enableSubmitButton}
                      onInvalid={this.disableSubmitButton}
                      onValidSubmit={this.submitForm}
                      onInvalidSubmit={this.notifyFormError}
                    >
                        <FormsyText
                          name="Name"
                          validations="isWords"
                          validationError={wordsError}
                          value={this.state.data.Name}
                          required
                          hintText="What is your name?"
                          floatingLabelText="Name"
                        />

                        <FormsyText
                          name="Surname"
                          validations="isWords"
                          validationError={wordsError}
                          value={this.state.data.Surname}
                          required
                          hintText="What is your surname?"
                          floatingLabelText="Surname"
                        />

                        <FormsyText
                          name="Email"
                          validations="isEmail"
                          validationError={emailError}                            
                          value={this.state.data.Email}
                          required
                          hintText="What is your email?"
                          floatingLabelText="Email"
                        />  

                        <FormsyText
                          name="Company"
                          value={this.state.data.Company}
                          hintText="What is your company name?"
                          floatingLabelText="Company"
                        />     

                        <FormsyText
                          name="About"
                          multiLine={true}
                          value={this.state.data.About}
                          rows={2}
                          rowsMax={10}
                          hintText="Write somthing about yourself"
                          floatingLabelText="About"
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

export default UpdateSpeaker;