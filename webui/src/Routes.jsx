import React from 'react';
import {Switch, Redirect} from 'react-router-dom';

import {RouteWithLayout} from './components';
import {Main as MainLayout, Minimal as MinimalLayout} from './layouts';
import {AuthProvider} from './AuthContext'
import ProtectedRoute from './ProtectedRoute'

import {
    Dashboard as DashboardView,
    SignIn as SignInView,
    NotFound as NotFoundView
} from './views';

const Routes = () => {
    return (
        // https://codesandbox.io/s/p71pr7jn50
        <AuthProvider>
            <Switch>
                <Redirect exact from="/" to="/dashboard"/>
                <ProtectedRoute path="/dashboard" layout={MainLayout} component={DashboardView}/>
                <RouteWithLayout component={SignInView} exact layout={MinimalLayout} path="/signin"/>
                <RouteWithLayout component={NotFoundView} exact layout={MinimalLayout} path="/notfound"/>
                <Redirect to="/notfound"/>
            </Switch>
        </AuthProvider>
    );
};

export default Routes;
