import React from "react";
import ReactDOM from "react-dom";
import { Provider } from 'react-redux';

import store from 'Include/store';
import App from "Page/index";

import {
	BrowserRouter as Router,
	Switch,
	Route,
	withRouter,
} from "react-router-dom";


const RouterApp = withRouter(App);


ReactDOM.render((
	<Provider store={store}>
		<Router>
			<Switch>
				<Route path="/">
					<RouterApp />
				</Route>
			</Switch>
		</Router>
	</Provider>
), document.body);
