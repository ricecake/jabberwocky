import React from 'react';
import PropTypes from 'prop-types';
import AppBar from '@material-ui/core/AppBar';
import Avatar from '@material-ui/core/Avatar';
import Grid from '@material-ui/core/Grid';
import Hidden from '@material-ui/core/Hidden';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import NotificationsIcon from '@material-ui/icons/Notifications';
import Toolbar from '@material-ui/core/Toolbar';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';
import Identicon from 'react-identicons';
import { withStyles } from '@material-ui/core/styles';
import NavigateNextIcon from '@material-ui/icons/NavigateNext';

import Link from '@material-ui/core/Link';
import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import { Link as RouterLink } from 'react-router-dom';

import { Show } from 'Component/Helpers';

import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { withRouter } from 'react-router-dom';

const urlNameMap = {
	'/': 'Home',
};

const ucfirst = (string = '') => string[0].toUpperCase() + string.slice(1);

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
	root: {
		display: 'flex',
		flexDirection: 'column',
		width: 360,
	},
	lists: {
		backgroundColor: theme.palette.background.paper,
		marginTop: theme.spacing(1),
	},
	nested: {
		paddingLeft: theme.spacing(4),
	},
});

const LinkRouter = (props) => <Link {...props} component={RouterLink} />;

const RouterBreadcrumbs = withRouter(
	withStyles(styles)((props) => {
		const { classes, location } = props;
		const pathnames = [
			'',
			...location.pathname.split('/').filter((x) => x),
		];

		return (
			<div className={classes.root}>
				<Breadcrumbs
					aria-label="breadcrumb"
					separator={<NavigateNextIcon fontSize="small" />}
				>
					{pathnames.map((value, index) => {
						const last = index === pathnames.length - 1;
						const to = `/${pathnames
							.slice(1, index + 1)
							.join('/')}`;

						return last ? (
							<Typography
								color="inherit"
								key={to}
								variant="h5"
								component="h1"
							>
								{urlNameMap[to] || ucfirst(value)}
							</Typography>
						) : (
							<LinkRouter
								color="inherit"
								to={to}
								key={to}
								variant="h5"
								component="h1"
							>
								{urlNameMap[to] || ucfirst(value)}
							</LinkRouter>
						);
					})}
				</Breadcrumbs>
			</div>
		);
	})
);

function Header(props) {
	const { classes, onDrawerToggle } = props;

	return (
		<React.Fragment>
			<AppBar color="primary" position="sticky" elevation={0}>
				<Toolbar>
					<Grid container spacing={1} alignItems="center">
						<Hidden smUp>
							<Grid item>
								<IconButton
									color="inherit"
									aria-label="open drawer"
									onClick={onDrawerToggle}
									className={classes.menuButton}
								>
									<MenuIcon />
								</IconButton>
							</Grid>
						</Hidden>
						<Grid item xs>
							<RouterBreadcrumbs />
						</Grid>
						<Grid item>
							<Tooltip title="Alerts • No alerts">
								<IconButton color="inherit">
									<NotificationsIcon />
								</IconButton>
							</Tooltip>
						</Grid>
						<Grid item>
							<IconButton
								color="inherit"
								className={classes.iconButtonAvatar}
							>
								<Avatar alt={props.name}>
									<Identicon string="Example!" size="25" />
								</Avatar>
							</IconButton>
							<Show If={props.name}>
								<Typography color="inherit" variant="caption">
									{props.name}
								</Typography>
							</Show>
						</Grid>
					</Grid>
				</Toolbar>
			</AppBar>
		</React.Fragment>
	);
}

Header.propTypes = {
	classes: PropTypes.object.isRequired,
	onDrawerToggle: PropTypes.func.isRequired,
};

const stateToProps = ({}) => ({});
const dispatchToProps = (dispatch) => bindActionCreators({}, dispatch);

export default connect(
	stateToProps,
	dispatchToProps
)(withStyles(styles)(Header));
