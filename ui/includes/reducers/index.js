import { combineReducers } from 'redux';
import payloadReducer from 'Include/reducers/payload';
import websocketReducer from 'Include/reducers/websocket';

const reducer = combineReducers({
	payload: payloadReducer,
	websocket: websocketReducer,
});

export default reducer;
