import React from 'react'
import {Route, Redirect} from 'react-router-dom'
import {AuthConsumer} from './AuthContext'

const ProtectedRoute = props => {
    const {layout: Layout, component: Component, ...rest} = props;

    return (<AuthConsumer>
        {({isAuth}) => (
            <Route
                render={matchProps =>
                    isAuth ? <Layout><Component {...matchProps} /> </Layout> : <Redirect to="/signin"/>
                }
                {...rest}
            />
        )}
    </AuthConsumer>)
}

export default ProtectedRoute
