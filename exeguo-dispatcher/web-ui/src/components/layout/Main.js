import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import { Route, Switch } from "react-router-dom";

import TopBar from './TopBar'
import SideMenu from './SideMenu'
import Dashboard from '../Dashboard'
import Login from '../Login'
import PasswordReset from '../PasswordReset'
import HostList from '../HostList'
import HostDetail from '../HostDetail'
import HostAdd from '../HostAdd'

const drawerWidth = 210;

const styles = theme => ({
  root: {
    display: 'flex',
  },
  drawer: {
    width: drawerWidth,
    flexShrink: 0,
  },
  drawerPaper: {
    width: drawerWidth,
  },
  content: {
    flexGrow: 1,
    padding: theme.spacing.unit * 3,
  },
  toolbar: theme.mixins.toolbar,
});

function Main(props) {
  const { classes } = props;

  return (
    <div className={classes.root}>
      <Switch>
        <Route path="/auth" render={({ match }) => (
          <React.Fragment>
            <TopBar />
            <main className={classes.content}>
              <div className={classes.toolbar} />
              <Route path={`${match.url}/login`} component={Login} />
              <Route path={`${match.url}/password-recovery`} component={PasswordReset} />
            </main>
          </React.Fragment>
        )}/>
        <Route render={props => (
          <React.Fragment>
            <TopBar />
            <SideMenu />
            <main className={classes.content}>
              <div className={classes.toolbar} />
              <Route path="" exact component={Dashboard} />

              <Route path="/hosts" render={({ match: { url } }) => (
                <Switch>
                  <Route exact path={`${url}/`} component={HostList} />
                  <Route path={`${url}/add`} component={HostAdd} />
                  <Route path={`${url}/:id`} component={HostDetail} />
                </Switch>
              )}/>
            </main>
          </React.Fragment>
        )}/>
      </Switch>
    </div>
  );
}

export default withStyles(styles)(Main);
