var UpdateMeetup = React.createClass({
    getInitialState: function() {
      return {
            date : '',
            title: '',
            description: '',
            presentations: [],
            voteTimeEnd: ''
        }
    },

    getMeetup() {
        $.ajax({
            url: '/meetup/' + getParameterByName("key") + '/',
            dataType: 'json',
            cache: false,
            success: function(data) {
            this.setState({title: data.Title})
            this.setState({description: data.Description})
            this.setState({voteTimeEnd: data.VoteTimeEnd.slice(0,-1)})
            this.setState({presentations: data.Presentations})
            this.setState({date: data.Date.slice(0,-1)});
        }.bind(this),
            error: function(xhr, status, err) {
            console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
    },

    componentDidMount: function() {
        this.getMeetup()
    },

    onDateChange(e) {
        let state = this.state;
        state['date'] = e.target.value;
        // Or (you can use below method to access component in another method)
        state['date'] = this.dateRef.value;
        this.setState(state);
    },

     onVoteTimeEndChange(e) {
        let state = this.state;
        state['voteTimeEnd'] = e.target.value;
        // Or (you can use below method to access component in another method)
        state['voteTimeEnd'] = this.dateRef.value;
        this.setState(state);
    },

    handleTitleChange: function(e) {
        this.setState({title: e.target.value});
    },

    handleDescriptionChange: function(e) {
        this.setState({description: e.target.value});
    },

    render: function() {
        return (
            <div className="panel panel-default col-md-6 col-md-offset-3">
                <div className="panel-heading">
                    <h3 className="panel-title">Update Meetup</h3>
                </div>
                <div className="panel-body">
                    <form className='form-horizontal' role='form' method='post' onSubmit={this.handleSubmit} action={'../../meetup/' + getParameterByName("key") + '/update'}>
                        <div className='form-group'>
                            <label htmlFor='title' className='col-sm-2 control-label'>Title</label>
                            <div className='col-sm-10'>
                                <input 
                                    type='text' 
                                    className='form-control' 
                                    id='title' name='title' 
                                    placeholder='Title' 
                                    value={this.state.title}
                                    onChange={this.handleTitleChange} />
                            </div>
                        </div>
                        <div className='form-group'>
                            <label htmlFor='date' className='col-sm-2 control-label'>Event Date</label>
                            <div className='col-sm-10'>
                                <div className='col-sm-6'>
                                    <div className='form-group'>
                                        <div className='input-group date'>
                                            <input className='form-control' type='datetime-local' ref={(date) => {this.dateRef = date;}} value={this.state.date} onChange={this.onDateChange}/>
                                            <span className='input-group-addon'>
                                                <span className='glyphicon glyphicon-calendar'></span>
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>    
                        <div className='form-group'>
                            <label htmlFor='date' className='col-sm-2 control-label'>Vote Time End</label>
                            <div className='col-sm-10'>
                                <div className='col-sm-6'>
                                    <div className='form-group'>
                                        <div className='input-group date'>
                                            <input className='form-control' type='datetime-local' value={this.state.voteTimeEnd} onChange={this.onVoteTimeEndChange}/>
                                            <span className='input-group-addon'>
                                                <span className='glyphicon glyphicon-calendar'></span>
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div> 
                        <div className='form-group'>
                            <label htmlFor='description' className='col-sm-2 control-label'>Description</label>
                            <div className='col-sm-10'>
                                <textarea 
                                    className='form-control' 
                                    rows='4'
                                    name='description'
                                    value={this.state.description}
                                    onChange={this.handleDescriptionChange}></textarea>
                            </div>
                        </div>
                        <div className='form-group'>
                            <label htmlFor='date' className='col-sm-2 control-label'>Presentations</label>
                                <div className='col-sm-10'>
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th>Title</th>
                                            <th>Speaker</th>
                                            <th>Action</th>
                                        </tr>
                                    </thead>
                                </table>
                            </div>
                        </div>
                        <div className='form-group'>
                            <div className='col-sm-10 col-sm-offset-2'>
                                <input id='submit' name='submit' type='submit' value='Save' className='btn btn-primary' />
                                <a href="#" id='cancel' name='cancel' type='cancel' className='btn btn-default' >Cancel</a>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
        )
    }
});


ReactDOM.render(
  <UpdateMeetup/>,
  document.getElementById('content')
);

function getParameterByName(name, url) {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}