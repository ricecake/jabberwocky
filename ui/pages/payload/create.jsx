import React from 'react';
import Editor from '@monaco-editor/react';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import Select from '@material-ui/core/Select';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import FormHelperText from '@material-ui/core/FormHelperText';
import FormControl from '@material-ui/core/FormControl';
import { makeStyles } from '@material-ui/core/styles';

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
	const [age, setAge] = React.useState('');

	const handleChange = (event) => {
		setAge(event.target.value);
	};

	return (
		<React.Fragment>
			<Editor
				height="87vh"
				defaultLanguage="javascript"
				theme="vs-dark"
				defaultValue={defaultScript}
			/>
			<Paper square className={classes.bottomBar}>
				<Toolbar>
					<FormControl className={classes.formControl}>
						<InputLabel id="script-security-level-input-label">
							Access Level
						</InputLabel>
						<Select
							labelId="script-security-level-input-label"
							id="demo-simple-select-helper"
							value={age}
							onChange={handleChange}
						>
							<MenuItem value={0}>Internal State Access</MenuItem>
							<MenuItem value={1}>Primitive Access</MenuItem>
							<MenuItem value={2}>Read-Only File Access</MenuItem>
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
					<TextField
						id="script-name"
						label="Payload Name"
						helperText="Payload display name"
					/>
					<Button variant="contained" color="primary">
						Save
					</Button>
				</Toolbar>
			</Paper>
		</React.Fragment>
	);
};
export default DefaultPage;
