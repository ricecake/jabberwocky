import { createActions, handleActions } from 'redux-actions';
import { MakeMerge } from 'Include/reducers/helpers';

import { send } from '@giantmachines/redux-websocket';

const defaultState = () => ({
	payloads: [],
});

export const saveScript = (name, access, body) => (dispatch, getState) => {
	console.log([name, access, body]);
	dispatch(
		send({
			Type: 'script',
			SubType: 'create',
			Content: {
				Name: name,
				Access: access,
				Body: body,
			},
		})
	);
};

/*
export const {
	createPayload,
} = createActions({
	createPayload: (name, access, body) => ({name, access, body}),
}, { prefix: "jabberwocky/payload" });

*/
const reducer = handleActions({}, defaultState());

const merge = MakeMerge((newState) => {
	return newState;
});

export default reducer;
