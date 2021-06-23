import React from 'react';
import { Switch, Route } from 'react-router-dom';

import TabBar from 'Component/TabBar';

export const DefaultPage = (props) => (
	<React.Fragment>
		<TabBar tabs={[{ label: 'Overview', value: '' }, { label: 'Charts' }]}>
			<Switch>
				<Route path="/Charts">
					<h1>There would be charts here</h1>
				</Route>
				<Route path="/">
					<h1>This would provide an Overview</h1>
				</Route>
			</Switch>
		</TabBar>
	</React.Fragment>
);
export default DefaultPage;
