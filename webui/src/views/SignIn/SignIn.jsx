import React, { useContext, useState } from 'react';
import { Redirect, withRouter } from 'react-router-dom';
import PropTypes from 'prop-types';
import axios from 'axios';
import { makeStyles } from '@material-ui/styles';
import {
    Grid,
    Button,
    TextField,
    Typography
} from '@material-ui/core';
import { AuthContext, AuthConsumer } from '../../AuthContext';

const useStyles = makeStyles(theme => ({
    root: {
        backgroundColor: theme.palette.background.default,
        height: '100%'
    },
    grid: {
        height: '100%'
    },
    name: {
        marginTop: theme.spacing(3),
        color: theme.palette.white
    },
    contentContainer: {},
    content: {
        height: '100%',
        display: 'flex',
        flexDirection: 'column'
    },
    contentBody: {
        flexGrow: 1,
        display: 'flex',
        alignItems: 'center',
        [theme.breakpoints.down('md')]: {
            justifyContent: 'center'
        }
    },
    form: {
        paddingLeft: 100,
        paddingRight: 100,
        paddingBottom: 125,
        flexBasis: 700,
        margin: '0 auto',
        [theme.breakpoints.down('sm')]: {
            paddingLeft: theme.spacing(2),
            paddingRight: theme.spacing(2)
        }
    },
    title: {
        marginTop: theme.spacing(3)
    },
    textField: {
        marginTop: theme.spacing(2)
    },
    signInButton: {
        margin: theme.spacing(2, 0)
    }
}));

const SignIn = (props) => {
    const contextValue = useContext(AuthContext);
    const [token, setToken] = useState(null);
    const [error, setError] = useState(null);

    const { history } = props;

    const classes = useStyles();

    const handleChange = event => {
        setToken(event.target.value)
    };

    const handleSignIn = event => {
        event.preventDefault();
        checkToken(token);
    };

    const onSuccess = () => {
        contextValue.login();

        history.push('/');
    };

    const checkToken = (token) => {
        axios.get(`/stat/livestat?jwt=${token}`).then(res => {
            onSuccess();
        }, err => {
            if (!token) return;

            if (err && err.response && err.response.status === 403) {
                console.log(err.response.data.data)
                setError(err.response.data.data);
            } else if (err.response && typeof err.response.data === 'string') {
                setError(err.response.data);
            } else if (err.message) {
                setError(err.message);
            } else {
                setError(JSON.stringify(err));
            }
        });
    };

    // Check if have not the jwt enabled
    checkToken();

    return (
        <div className={classes.root}>
            <AuthConsumer>
                {({ isAuth }) => (
                    !isAuth ? (<Grid className={classes.grid} container>
                        <Grid className={classes.content} item lg={12} xs={12}>
                            <div className={classes.content}>
                                <div className={classes.contentBody}>
                                    <form className={classes.form} onSubmit={handleSignIn}>
                                        <Typography className={classes.title} variant="h2">
                                            Sign in
                                        </Typography>
                                        <TextField error={!!error} helperText={error} className={classes.textField} fullWidth label="Token with your secret" name="token"
                                            onChange={handleChange} type="text" variant="outlined" />
                                        <Button className={classes.signInButton} color="primary" fullWidth size="large"
                                            type="submit" variant="contained">
                                            Sign in now
                                        </Button>
                                    </form>
                                </div>
                            </div>
                        </Grid>
                    </Grid>) : <Redirect to="/" />
                )}
            </AuthConsumer>
        </div>
    );
};

SignIn.propTypes = {
    history: PropTypes.object
};

export default withRouter(SignIn);
