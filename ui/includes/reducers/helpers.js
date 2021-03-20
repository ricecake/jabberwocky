import _ from "lodash";

const merge = (oldState, newState) => _.mergeWith({}, oldState, newState, (o, s)=> _.isArray(o)? s : undefined);
const noopValidate = (input) => (input);

export const MakeMerge = (validate = noopValidate) => (oldState, newState) => {
	return validate(merge(oldState, newState))
}
