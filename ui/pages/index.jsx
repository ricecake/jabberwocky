import React, { Suspense, lazy } from 'react';
import { Switch, Route } from 'react-router-dom';

import {
	createTheme,
	ThemeProvider,
	withStyles,
} from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Hidden from '@material-ui/core/Hidden';

import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { connect as wsConnect } from '@giantmachines/redux-websocket';

let theme = createTheme({
	palette: {
		primary: {
			light: '#63ccff',
			main: '#009be5',
			dark: '#006db3',
			background: '#18202c',
		},
	},
	typography: {
		h5: {
			fontWeight: 500,
			fontSize: 26,
			letterSpacing: 0.5,
		},
	},
	shape: {
		borderRadius: 8,
	},
	props: {
		MuiTab: {
			disableRipple: true,
		},
	},
	mixins: {
		toolbar: {
			minHeight: 48,
		},
	},
});

theme = {
	...theme,
	overrides: {
		MuiDrawer: {
			paper: {
				backgroundColor: '#18202c',
			},
		},
		MuiButton: {
			label: {
				textTransform: 'none',
			},
			contained: {
				boxShadow: 'none',
				'&:active': {
					boxShadow: 'none',
				},
			},
		},
		MuiTabs: {
			root: {
				marginLeft: theme.spacing(1),
			},
			indicator: {
				height: 3,
				borderTopLeftRadius: 3,
				borderTopRightRadius: 3,
				backgroundColor: theme.palette.common.white,
			},
		},
		MuiTab: {
			root: {
				textTransform: 'none',
				margin: '0 16px',
				minWidth: 0,
				padding: 0,
				[theme.breakpoints.up('md')]: {
					padding: 0,
					minWidth: 0,
				},
			},
		},
		MuiIconButton: {
			root: {
				padding: theme.spacing(1),
			},
		},
		MuiTooltip: {
			tooltip: {
				borderRadius: 4,
			},
		},
		MuiDivider: {
			root: {
				backgroundColor: '#404854',
			},
		},
		MuiListItemText: {
			primary: {
				fontWeight: theme.typography.fontWeightMedium,
			},
		},
		MuiListItemIcon: {
			root: {
				color: 'inherit',
				marginRight: 0,
				'& svg': {
					fontSize: 20,
				},
			},
		},
		MuiAvatar: {
			root: {
				width: 32,
				height: 32,
			},
		},
	},
};

const drawerWidth = 256;

const styles = {
	root: {
		display: 'flex',
		minHeight: '100vh',
	},
	drawer: {
		[theme.breakpoints.up('sm')]: {
			width: drawerWidth,
			flexShrink: 0,
		},
	},
	app: {
		flex: 1,
		display: 'flex',
		flexDirection: 'column',
	},
	main: {
		flex: 1,
		padding: theme.spacing(6, 4),
		background: '#eaeff1',
	},
	footer: {
		padding: theme.spacing(2),
		background: '#eaeff1',
	},
};

import Navigator from 'Component/Navigator';
import Header from 'Component/Header';

const Home = lazy(() => import('Page/home'));

const Payload = lazy(() => import('Page/payload'));
const PayloadCreate = lazy(() => import('Page/payload/create'));
const PayloadEdit = lazy(() => import('Page/payload/create'));

const Server = lazy(() => import('Page/server'));

const Utility = lazy(() => import('Page/utility'));
const UtilityTail = lazy(() => import('Page/utility/tail'));

const Job = lazy(() => import('Page/job'));
const JobCreate = lazy(() => import('Page/job/create'));
const JobEdit = lazy(() => import('Page/job/edit'));

const Agent = lazy(() => import('Page/agent'));
const AgentManage = lazy(() => import('Page/agent/manage'));
const AgentTerminal = lazy(() => import('Page/agent/terminal'));

export const App = (props) => {
	const { classes, wsConnect } = props;

	React.useEffect(() => {
		var loc = window.location,
			new_uri;

		if (loc.protocol === 'https:') {
			new_uri = 'wss:';
		} else {
			new_uri = 'ws:';
		}
		new_uri += '//' + loc.host;
		new_uri += '/ws/admin';

		wsConnect(new_uri);
	}, []);

	const [mobileOpen, setMobileOpen] = React.useState(false);

	const handleDrawerToggle = () => {
		setMobileOpen(!mobileOpen);
	};

	return (
		<ThemeProvider theme={theme}>
			<CssBaseline />
			<div className={classes.root}>
				<nav className={classes.drawer}>
					<Hidden smUp implementation="js">
						<Navigator
							PaperProps={{ style: { width: drawerWidth } }}
							variant="temporary"
							open={mobileOpen}
							onClose={handleDrawerToggle}
						/>
					</Hidden>
					<Hidden xsDown implementation="css">
						<Navigator
							PaperProps={{ style: { width: drawerWidth } }}
						/>
					</Hidden>
				</nav>
				<div className={classes.app}>
					<Header onDrawerToggle={handleDrawerToggle} />
					<Suspense fallback={<div>Loading...</div>}>
						<Switch>
							<Route path="/agent/terminal">
								<AgentTerminal />
							</Route>
							<Route path="/agent/manage">
								<AgentManage />
							</Route>
							<Route path="/agent">
								<Agent />
							</Route>

							<Route path="/utility/tail">
								<UtilityTail />
							</Route>
							<Route path="/utility">
								<Utility />
							</Route>

							<Route path="/server">
								<Server />
							</Route>

							<Route path="/job/create">
								<JobCreate />
							</Route>
							<Route path="/job/edit">
								<JobEdit />
							</Route>
							<Route path="/job">
								<Job />
							</Route>

							<Route path="/payload/create">
								<PayloadCreate />
							</Route>
							<Route path="/payload/edit">
								<PayloadEdit />
							</Route>
							<Route path="/payload/">
								<Payload />
							</Route>

							<Route path="/">
								<Home />
							</Route>
						</Switch>
					</Suspense>
				</div>
			</div>
		</ThemeProvider>
	);
};

const stateToProps = ({}) => ({});
const dispatchToProps = (dispatch) =>
	bindActionCreators({ wsConnect }, dispatch);

export default connect(stateToProps, dispatchToProps)(withStyles(styles)(App));
