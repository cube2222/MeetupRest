var UpdatePresentation = React.createClass({
    getInitialState: function() {
      return {data: {}, title: '', description: ''}
    },

    getPresentation() {
        $.ajax({
            url: '/presentation/' + getParameterByName("key") + '/',
            dataType: 'json',
            cache: false,
            success: function(data) {
            this.setState({title: data.Title})
            this.setState({description: data.Description})
            this.setState({data: data});
        }.bind(this),
            error: function(xhr, status, err) {
            console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
    },

    componentDidMount: function() {
        this.getPresentation()
    },

    handleTitleChange: function(e) {
        this.setState({title: e.target.value});
    },

    handleDescriptionChange: function(e) {
        this.setState({description: e.target.value});
    },

    handleSubmit: function(e) {
        e.preventDefault();
        var title = this.state.title.trim();
        var description = this.state.description.trim();
        if (!title || !description) {
            return;
        }
        $.ajax({
            url: '../../presentation/' + getParameterByName("key") + '/update' ,
            contentType: 'application/json; charset=utf-8',
            dataType: 'json',
            type: 'POST',
            data: JSON.stringify({Title: title, Description: description})
        });
    },

    render: function() {

        return (
            <div className="panel panel-default col-md-6 col-md-offset-3">
                <div className="panel-heading">
                    <h3 className="panel-title">Update Presentation</h3>
                </div>
                <div className="panel-body">
                    <form className='form-horizontal' role='form' method='post' onSubmit={this.handleSubmit} action={'../../presentation/' + getParameterByName("key") + '/update'}>
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
  <UpdatePresentation/>,
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