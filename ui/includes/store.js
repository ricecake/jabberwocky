import { createStore, applyMiddleware, compose } from 'redux';
import ReduxThunk from 'redux-thunk';

import reducer from 'Include/reducers';

const loggerMiddleware = (store) => (next) => (action) => {
	console.log('Action type:', action.type);
	console.log('action', action);
	console.log('State before:', store.getState());
	next(action);
	console.log('State after:', store.getState());
};

const initialState = {};

const createStoreWithMiddleware = compose(
	applyMiddleware(loggerMiddleware),
	applyMiddleware(ReduxThunk)
)(createStore);

const store = createStoreWithMiddleware(reducer, initialState);

export default store;
