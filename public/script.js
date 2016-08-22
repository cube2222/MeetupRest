var DeleteForm = React.createClass({
    getInitialState: function() {
      return {data: []}
    },

    fetchPresentation() {
        $.ajax({
            url: '/presentation/list',
            dataType: 'json',
            cache: false,
            success: function(data) {
                this.setState({data: data});
            }.bind(this),
            error: function(xhr, status, err) {
                console.error(this.props.url, status, err.toString());
            }.bind(this)
      });
    },

    componentDidMount: function() {
      this.fetchPresentation()
    },

    render: function() {
        var FormatItem = function(item) {
            return <option value={item.Key}>{item.Title} - {item.Speaker}</option>;
        };

        return (
        <form className="commentForm " action='/presentation/delete' value="Post">
            <div className="panel panel-default col-md-6 col-md-offset-3">
                <div className="panel-heading">
                    <h3 className="panel-title">Delete Presentation</h3>
                </div>
                <div className="panel-body">
                    <select name="PresentationId" onChange={this.handleOptionChange}>{this.state.data.map(FormatItem)}</select>
                    <div className="text-right">
                        <button type="submit" className="btn btn-default text-right">Remove</button>
                    </div>
                </div>
            </div>
        </form>
        )
        
    },
});

ReactDOM.render(<DeleteForm/>, document.getElementById('app'));
