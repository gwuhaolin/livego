import React from 'react';
import {Switch, Redirect} from 'react-router-dom';

import {RouteWithLayout} from './components';
import {Main as MainLayout, Minimal as MinimalLayout} from './layouts';

import {
    Dashboard as DashboardView,
    SignIn as SignInView,
    NotFound as NotFoundView
} from './views';

function loggedIn() {
    return false
}

function requireAuth(nextState, replace) {
    console.log("holaa")
    if (!loggedIn()) {
        replace({
            pathname: '/signin'
        })
    }
}

const Routes = () => {
    return (
        <Switch>
            <Redirect exact from="/" to="/dashboard"/>
            <RouteWithLayout component={DashboardView} exact layout={MainLayout} path="/dashboard" onEnter={requireAuth}/>
            <RouteWithLayout component={SignInView} exact layout={MinimalLayout} path="/signin"/>
            <RouteWithLayout component={NotFoundView} exact layout={MinimalLayout} path="/notfound"/>
            <Redirect to="/notfound"/>
        </Switch>
    );
};

export default Routes;
