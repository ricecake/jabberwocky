import React, { Suspense, lazy } from 'react';
import {
  Switch,
  Route,
} from "react-router-dom";

import { connect } from "react-redux";
import { bindActionCreators } from 'redux'


export const App = (props) => {
	console.log(props);
	return (
		<div>
			This is a potato
		</div>
	);
};

const stateToProps = (state) => (state);
const dispatchToProps = (dispatch) => bindActionCreators({}, dispatch);

export default connect(stateToProps, dispatchToProps)(App);
