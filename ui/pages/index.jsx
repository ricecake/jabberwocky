import React, { Suspense, lazy } from 'react';
import { Switch, Route } from 'react-router-dom';

import {
	createMuiTheme,
	ThemeProvider,
	withStyles,
} from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Hidden from '@material-ui/core/Hidden';

import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

let theme = createMuiTheme({
	palette: {
		primary: {
			light: '#63ccff',
			main: '#009be5',
			dark: '#006db3',
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
// const Admin = lazy(() => import('Page/admin'));

// const Splc = lazy(() => import('Page/splc'));
// const SplcCategories = lazy(() => import('Page/splc/categories'));
// const SplcGroups = lazy(() => import('Page/splc/groups'));
// const SplcDomains = lazy(() => import('Page/splc/domains'));

// const Hosting = lazy(() => import('Page/hosting'));
// const HostingWebhosts = lazy(() => import('Page/hosting/webhosts'));
// const HostingAddresses = lazy(() => import('Page/hosting/addresses'));

// const Media = lazy(() => import('Page/media'));
// const MediaTemplates = lazy(() => import('Page/media/templates'));
// const MediaTweets = lazy(() => import('Page/media/tweets'));

export const App = (props) => {
	const { classes } = props;

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
const dispatchToProps = (dispatch) => bindActionCreators({}, dispatch);

export default connect(stateToProps, dispatchToProps)(withStyles(styles)(App));
