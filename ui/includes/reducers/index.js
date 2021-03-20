import { combineReducers } from 'redux';
import fileReducer from 'Include/reducers/file';

const reducer = combineReducers({
	file: fileReducer,
});

export default reducer;
