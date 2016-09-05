var PresentationList = React.createClass({
    getInitialState: function() {
      return {data: []}
    },

    componentDidMount: function() {
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

    render: function() {
      var presentations = this.state.data.map(function(presentation) {
          return (
            <tr key={presentation.Key}>
              <td>{presentation.Title}</td>
              <td>{presentation.Speakers.join(', ')}</td>
              <td>{presentation.Votes}</td>
              <td><a href={"/public/html/presentation.html?key=" + presentation.Key} className="btn btn-info" role="button">Open</a></td>
            </tr>
          );
      });
      return (
        <div className="presentationList">
            <table className="table">
                <thead>
                    <tr>
                        <th>Title</th>
                        <th>Speakers</th>
                        <th>Votes</th>
                    </tr>
                </thead>
                <tbody>
                    {presentations}
                </tbody>
            </table>
        </div>
      );
    }
  });

ReactDOM.render(
  <PresentationList/>,
  document.getElementById('content')
);