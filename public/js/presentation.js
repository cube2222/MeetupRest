var PresentationList = React.createClass({
    updateState() {
      $.ajax({
        url: '/presentation/' + getParameterByName("key") + '/',
        dataType: 'json',
        cache: false,
        success: function(data) {
          this.setState({data: data});
        }.bind(this),
        error: function(xhr, status, err) {
          console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
      $.ajax({
        url: '/presentation/' + getParameterByName("key") + '/hasUpvoted',
        dataType: 'text',
        cache: false,
        success: function(data) {
          this.setState({hasUpvoted: data});
        }.bind(this),
        error: function(xhr, status, err) {
          console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
      $.ajax({
        url: '/isLoggedIn',
        dataType: 'text',
        cache: false,
        success: function(data) {
          this.setState({isLoggedIn: data});
        }.bind(this),
        error: function(xhr, status, err) {
          console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
    },

    upvote() {
      $.ajax({
        url: '/presentation/' + getParameterByName("key") + '/upvote',
        dataType: 'text',
        cache: false,
        success: function(data) {
          this.updateState()
        }.bind(this),
        error: function(xhr, status, err) {
          console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
    },

    downvote() {
      $.ajax({
        url: '/presentation/' + getParameterByName("key") + '/downvote',
        dataType: 'text',
        cache: false,
        success: function(data) {
          this.updateState()
        }.bind(this),
        error: function(xhr, status, err) {
          console.error(this.props.url, status, err.toString());
        }.bind(this)
      });
    },

    login() {
        window.open("/_ah/login?continue=" + window.location.href,"_self")
    },

    getInitialState: function() {
      return {data: {}, hasUpvoted: "false", isLoggedIn: "false"}
    },

    componentDidMount: function() {
      this.updateState()
      setInterval(this.updateState, 2000);
    },

    render: function() {
      return (
          <div className="presentationList">
            <div className="container theme-showcase">
                <h1>{this.state.data.Title}</h1>
                <h2>{this.state.data.Speaker}</h2>
                <h4>Votes: {this.state.data.Votes} <button onClick={this.state.isLoggedIn == "true" ? (this.state.hasUpvoted == "false" ? this.upvote : this.downvote) : this.login} className={this.state.hasUpvoted == "false" ? "btn btn-info" : "btn btn-success"}>{this.state.hasUpvoted == "false" ? "Upvote!" : "Undo upvote."}</button></h4>
                <div className="well">{this.state.data.Description}</div>
            </div>
          </div>
      );
    }
  });

ReactDOM.render(
  <PresentationList/>,
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