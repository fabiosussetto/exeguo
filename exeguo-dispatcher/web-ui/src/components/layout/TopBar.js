import React from "react";
import { withStyles } from "@material-ui/core/styles";
import AppBar from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";

import classNames from "classnames";

const styles = theme => ({
  appBar: {
    boxShadow: "none",
    zIndex: theme.zIndex.drawer + 1,
    minHeight: 45
  }
});

function TopBar(props) {
  const { classes } = props;

  return (
    <AppBar position="fixed" className={classNames(classes.appBar)}>
      <Toolbar variant="dense">
        <Typography variant="h6" color="inherit" noWrap>
          Exeguo
        </Typography>
      </Toolbar>
    </AppBar>
  );
}

export default withStyles(styles)(TopBar);
