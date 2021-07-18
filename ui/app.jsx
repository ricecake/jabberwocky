import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import { connect } from '@giantmachines/redux-websocket';

import store from 'Include/store';
import App from 'Page/index';

import {
	BrowserRouter as Router,
	Switch,
	Route,
	withRouter,
} from 'react-router-dom';

const RouterApp = withRouter(App);

ReactDOM.render(
	<Provider store={store}>
		<Router>
			<Switch>
				<Route path="/">
					<RouterApp />
				</Route>
			</Switch>
		</Router>
	</Provider>,
	document.body
);

var loc = window.location,
	new_uri;
if (loc.protocol === 'https:') {
	new_uri = 'wss:';
} else {
	new_uri = 'ws:';
}
new_uri += '//' + loc.host;
new_uri += '/ws/admin';

store.dispatch(connect(new_uri));
