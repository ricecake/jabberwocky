import { createActions, handleActions } from 'redux-actions';
import { MakeMerge } from 'Include/reducers/helpers';
import _ from 'lodash';

const defaultState = () => ({});

const mapContent = ({ Content }) => Content;
const mapMeta = ({ Content, ...rest }) => rest;
const simpleMessage = [mapContent, mapMeta];

export const actions = createActions(
	{
		agent: {
			sync: simpleMessage,
		},
		// script: {},
		// server: {},
		unknown: simpleMessage,
	},
	{ prefix: 'jabberwocky/ws' }
);

export const websocketMessageMiddleware = (store) => (next) => (action) => {
	if (action.type === 'REDUX_WEBSOCKET::MESSAGE') {
		let body = action.payload.message;
		// action.type = `jabberwocky/${body.Type}::${body.SubType}`;
		// action.payload = body;
		let actionCreator = _.get(
			actions,
			[body.Type, body.SubType],
			actions.unknown
		);
		action = actionCreator(body);
	}

	next(action);
};

const reducer = handleActions({}, defaultState());

const merge = MakeMerge((newState) => {
	return newState;
});

export default reducer;
