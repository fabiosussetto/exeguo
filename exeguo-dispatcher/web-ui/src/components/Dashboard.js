import React from 'react';
import { withStyles } from '@material-ui/core/styles';
// import Typography from '@material-ui/core/Typography';


const styles = theme => ({
  appBar: {
    boxShadow: "none",
    zIndex: theme.zIndex.drawer + 1,
    minHeight: 45,
  },
  toolbar: theme.mixins.toolbar,
});

function Dashboard(props) {

  return (
    <div>
      Dashboard
    </div>
  );
}

export default withStyles(styles)(Dashboard);
