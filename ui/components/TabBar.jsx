import React from 'react';
import AppBar from '@material-ui/core/AppBar';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import { withStyles } from '@material-ui/core/styles';

import { HashRouter as Router, NavLink, withRouter } from 'react-router-dom';

import { Show } from 'Component/Helpers';

const RouterTabs = withRouter((props) => (
	<Tabs value={props.location.pathname} {...props}>
		{props.children}
	</Tabs>
));

const lightColor = 'rgba(255, 255, 255, 0.7)';

const styles = (theme) => ({
	secondaryBar: {
		zIndex: 0,
	},
	menuButton: {
		marginLeft: -theme.spacing(1),
	},
	iconButtonAvatar: {
		padding: 4,
	},
	link: {
		textDecoration: 'none',
		color: lightColor,
		'&:hover': {
			color: theme.palette.common.white,
		},
	},
	button: {
		borderColor: lightColor,
	},
});

function TabBar(props) {
	const { classes } = props;

	return (
		<React.Fragment>
			<Show If={props.tabs}>
				<Router hashType="noslash">
					<AppBar
						component="div"
						className={classes.secondaryBar}
						color="secondary"
						position="static"
						elevation={0}
					>
						<RouterTabs textColor="inherit">
							{(props.tabs || [])
								.sort(({ order: a }, { order: b }) => a - b)
								.map(({ label, ...rest }) => ({
									label,
									value: label.replace(/[]/g, ''),
									...rest,
								}))
								.map(({ label, value }) => (
									<Tab
										textColor="inherit"
										label={label}
										value={`/${value}`}
										component={NavLink}
										to={`/${value}`}
									/>
								))}
						</RouterTabs>
					</AppBar>
					<Show If={props.children}>{props.children}</Show>
				</Router>
			</Show>
		</React.Fragment>
	);
}

export default withStyles(styles)(TabBar);
