import { combineReducers } from 'redux';
import payloadReducer from 'Include/reducers/payload';

const reducer = combineReducers({
	payload: payloadReducer,
});

export default reducer;
