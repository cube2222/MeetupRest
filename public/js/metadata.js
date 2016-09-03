var MetadataManger = React.createClass({
    getInitialState: function() {
      return {key: '', value: ''}
    },

    handleKeyChange: function(e) {
        this.setState({key: e.target.value});
    },

    handleValueChange: function(e) {
        this.setState({value: e.target.value});
    },

    handleSetClick: function(e) {
        e.preventDefault();
        var key = this.state.key.trim();
        var value = this.state.value.trim();
        if (!key) {
            return;
        }
        $.ajax({
            url: '/metadata/' + key + '/?data=' + value ,
            dataType: 'text',
            type: 'POST',
            success: function (val) {
                this.setState({value: val});
            }.bind(this),
                error: function (xhr, status, err) {
                console.error(this.props.url, status, err.toString());
            }.bind(this)
        });
    },

    handleGetClick: function(e) {
        e.preventDefault();
        var key = this.state.key.trim();
        if (!key) {
            return;
        }
        $.ajax({
            url: '/metadata/' + key + '/',
            dataType: 'text',
            type: 'GET',
            success: function (val) {
                this.setState({value: val});
            }.bind(this),
                error: function (xhr, status, err) {
                console.error(this.props.url, status, err.toString());
            }.bind(this)
        });

    },

    render: function() {

        return (
            <div className="panel panel-default col-md-6 col-md-offset-3">
                <div className="panel-heading">
                    <h3 className="panel-title">Metadata manager</h3>
                </div>
                <div className="panel-body">
                    <form className='form-horizontal' role='form' method='post'>
                        <div className='col-sm-10'>    
                            <div className='form-group'>
                                <label htmlFor="key">Key</label>
                                <input className="form-control"
                                        id="key"
                                        name="key"
                                        size="30" 
                                        type="text" 
                                        value={this.state.key}
                                        onChange={this.handleKeyChange}  />
                            </div>
                        </div>
                        
                        <div className='col-sm-10'>
                            <div className='form-group'>
                                <label htmlFor="value">Value</label>
                                <input className="form-control"
                                        id="value"
                                        name="value"
                                        size="30" 
                                        type="text"
                                        value={this.state.value}
                                        onChange={this.handleValueChange}  />
                            </div>
                        </div>
                        
                        <div className='col-sm-10'>
                            <div className='form-group'>
                                <div className='col-sm-10 col-sm-offset-2'>
                                    <input id='get' name='get' type='submit' value='Get' className='btn btn-primary' onClick={this.handleGetClick} />
                                    <input id='set' name='set' type='submit' value='Set' className='btn btn-primary' onClick={this.handleSetClick}/>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
        )
    }
});

ReactDOM.render(
  <MetadataManger/>,
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