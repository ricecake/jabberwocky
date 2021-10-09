import React from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import * as monaco from 'monaco-editor';
import Editor from '@monaco-editor/react';

//window.monaco = monaco;

import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import Select from '@material-ui/core/Select';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormHelperText from '@material-ui/core/FormHelperText';
import FormControl from '@material-ui/core/FormControl';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';

import { saveScript } from 'Include/reducers/payload';

const defaultScript = `/*********
 * Function Name
 * =============
 * Detail what your function does
 */

log.Info("Starting work");
//Place your function body here
log.Info("Finished");
`;

const useStyles = makeStyles((theme) => ({
	formControl: {
		margin: theme.spacing(1),
		minWidth: 120,
	},
	selectEmpty: {
		marginTop: theme.spacing(2),
	},
	bottomBar: {
		position: 'fixed',
		background: theme.palette.primary.light,
		top: 'auto',
		width: '100%',
		bottom: 0,
	},
}));

export const DefaultPage = (props) => {
	const classes = useStyles();
	const [access, setAccess] = React.useState('');
	const [name, setName] = React.useState('');
	const [scriptBody, setScriptBody] = React.useState('');

	const handleChangeAccess = (event) => {
		setAccess(event.target.value);
	};

	const handleChangeName = (event) => {
		setName(event.target.value);
	};

	const handleChangeScriptBody = (value, event) => {
		setScriptBody(value);
	};

	const submitScript = () => {
		props.saveScript(access, name, scriptBody);
	};

	let isDisabled = access === '' || name === '' || scriptBody === '';

	return (
		<React.Fragment>
			<Editor
				height="87vh"
				defaultLanguage="javascript"
				theme="vs-dark"
				defaultValue={defaultScript}
				onChange={handleChangeScriptBody}
			/>
			<Paper square className={classes.bottomBar}>
				<Toolbar>
					<Grid
						container
						direction="row"
						spacing={3}
						alignItems="center"
						justifyContent="space-around"
					>
						<Grid item xs>
							<TextField
								id="script-name"
								label="Payload Name"
								helperText="Payload display name"
								value={name}
								onChange={handleChangeName}
							/>
						</Grid>
						<Grid item xs>
							<FormControl className={classes.formControl}>
								<InputLabel id="script-security-level-input-label">
									Access Level
								</InputLabel>
								<Select
									labelId="script-security-level-input-label"
									id="demo-simple-select-helper"
									value={access}
									onChange={handleChangeAccess}
								>
									<MenuItem value={0}>
										Internal State Access
									</MenuItem>
									<MenuItem value={1}>
										Primitive Access
									</MenuItem>
									<MenuItem value={2}>
										Read-Only File Access
									</MenuItem>
									<MenuItem value={3}>
										Read/Write File Access
									</MenuItem>
									<MenuItem value={4}>
										Arbitrary Command Execution
									</MenuItem>
								</Select>
								<FormHelperText>
									Host System Access Level
								</FormHelperText>
							</FormControl>
						</Grid>
						<Grid item xs>
							<Button
								variant="contained"
								color="primary"
								onClick={submitScript}
								disabled={isDisabled}
							>
								Save
							</Button>
						</Grid>
					</Grid>
				</Toolbar>
			</Paper>
		</React.Fragment>
	);
};

const stateToProps = ({ payload }) => ({ payload });
const dispatchToProps = (dispatch) =>
	bindActionCreators({ saveScript }, dispatch);

export default connect(stateToProps, dispatchToProps)(DefaultPage);
