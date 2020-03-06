import React from 'react'
import { BrowserRouter as Router, Route } from 'react-router-dom';
import Login from "./pages/login"
import Field from "./pages/field/field";
import Socket from "./socket/socket";


class App extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            isLoggedIn: false
        }

        Socket.waitForSocket()
    }

    onLogin() {
        this.setState({
            isLoggedIn: true
        })
    }

    render() {
        if (!this.state.isLoggedIn) {
            return (
                <Login onLoggedIn={this.onLogin}></Login>
            )
        }

        return (
            <Router>
                <Route path="/" component={Field}></Route>
            </Router>
        )
    }
}

export default App