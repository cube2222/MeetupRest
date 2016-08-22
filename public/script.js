class Profile extends React.Component {
    render() {
        return <h2> Hello world !!</h2>;
    }
}
//ReactDOM.render(<Profile />, document.getElementById('app'));

class DeletePresentation extends React.Component {
    render() {
        // return (
        //     <div>
        //         <h1> Deleting Presentation Form </h1>
        //         <label>By Title:</label>
        //         <select name="PresentationId">
                    
        //             <option value="123">   </option>
                    
        //         </select>
        //         <input type="submit" value="Remove"/>
        //     </div>
        // );
        return (
            React.createElement('form', {className: 'ContactForm'},
                React.createElement('input', {
                    type: 'text',
                    placeholder: 'Name (required)',
                    value: this.props.value.name,
                }),
                React.createElement('input', {
                    type: 'email',
                    placeholder: 'Email',
                    value: this.props.value.email,
                }),
                React.createElement('textarea', {
                    placeholder: 'Description',
                    value: this.props.value.description,
                }),
                React.createElement('button', {type: 'submit'}, "Add Contact")
            )
            );        
    }
}

ReactDOM.render(<DeletePresentation />, document.getElementById('delete'));
