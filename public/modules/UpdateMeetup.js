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
import DatePicker from 'material-ui/DatePicker';


const UpdateMeetup = React.createClass({
    getInitialState() {
        return {
            canSubmit: false,
            data: {},
            date: null,
            voteTimeEnd: null,
        };
    },

    componentDidMount: function() {
        this.getMeetup()
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
            url: '../../meetup/' + this.props.params.meetupId + '/update',
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

    getMeetup() {
        $.ajax({
            url: '/meetup/' + this.props.params.meetupId + '/',
            dataType: 'json',
            cache: false,
            success: function(data) {
                        var tmpdate = new Date(data.Date);
                        var tmpVoteTimeEnd = new Date(data.VoteTimeEnd);

                        //TODO: What I should do with Presentations list?

                        // let speakersString = data.Speakers.map(function (i) {
                        //     return i.Name;
                        // }).join(', ');

                        this.setState({
                            data: data,
                            date: tmpdate,
                            voteTimeEnd: tmpVoteTimeEnd,
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

    formatDate(date){
        return new Date(date.getFullYear() + "/" + (date.getMonth() + 1) + "/" + date.getDate());
    },

    render() {
        let {paperStyle, submitStyle } = this.styles;
        let { wordsError } = this.errorMessages;

        return (
            <MuiThemeProvider muiTheme={getMuiTheme()}>
                <Paper style={paperStyle}>
                    <h2>Forms for updating meetup</h2>
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
                          hintText="What is meetup title?"
                          floatingLabelText="Title"
                        />
                        <FormsyDate
                          name="Date"
                          value={this.state.date}
                          required
                          floatingLabelText="Date"
                        />
                        <FormsyDate
                          name="VoteTimeEnd"
                          value={this.state.voteTimeEnd}
                          required
                          floatingLabelText="Vote Time End"
                        />

                        <FormsyText
                          name="Description"
                          value={this.state.data.Description}
                          required
                          multiLine={true}
                          rows={2}
                          rowsMax={10}
                          hintText="Description for meetup.com"
                          floatingLabelText="Descryption"
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

export default UpdateMeetup;




