var UpdateMeetup = React.createClass({

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
                                    placeholder='Title' />
                            </div>
                        </div>
                        <div className='form-group'>
                            <label htmlFor='description' className='col-sm-2 control-label'>Description</label>
                            <div className='col-sm-10'>
                                <textarea 
                                    className='form-control' 
                                    rows='4' 
                                    name='description'></textarea>
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