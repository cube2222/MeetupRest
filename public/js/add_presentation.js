var AddPresentation = React.createClass({
    getInitialState: function() {
      return {data: {}, title: '', description: '', speakers: ''}
    },

    handleTitleChange: function(e) {
        this.setState({title: e.target.value});
    },

    handleDescriptionChange: function(e) {
        this.setState({description: e.target.value});
    },

    handleSpeakersChange: function(e) {
        this.setState({speakers: e.target.value});
    },

    handleSubmit: function(e) {
        e.preventDefault();
        var title = this.state.title.trim();
        var description = this.state.description.trim();
        if (!title || !description) {
            return;
        }
        $.ajax({
            url: '../../presentation/' ,
            contentType: 'application/json; charset=utf-8',
            dataType: 'json',
            type: 'POST',
            data: JSON.stringify({Title: title, Description: description, Speakers: speakers})
        });
    },

    render: function() {
        return (
            <div className="panel panel-default col-md-6 col-md-offset-3">
                <div className="panel-heading">
                    <h3 className="panel-title">Add Presentation</h3>
                </div>
                <div className="panel-body">
                    <form className='form-horizontal' role='form' method='post' onSubmit={this.handleSubmit} action='../../presentation/'>
                        <div className='form-group'>
                            <label htmlFor='title' className='col-sm-2 control-label'>Title</label>
                            <div className='col-sm-10'>
                                <input 
                                    type='text' 
                                    className='form-control' 
                                    id='title' name='title' 
                                    placeholder='Title' 
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
                                    onChange={this.handleDescriptionChange}></textarea>
                            </div>
                        </div>
                        <div className='form-group'>
                            <label htmlFor='speakers' className='col-sm-2 control-label'>Speakers</label>
                            <div className='col-sm-10'>
                                <input  
                                    className='form-control' 
                                    rows='4' 
                                    name='description' 
                                    onChange={this.handleSpeakersChange}/>
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
  <AddPresentation/>,
  document.getElementById('content')
);