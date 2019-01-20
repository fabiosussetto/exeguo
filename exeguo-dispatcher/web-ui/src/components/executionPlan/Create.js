import React from "react";
import { observer, inject } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import { withSnackbar } from "notistack";

import { unstable_Box as Box } from "@material-ui/core/Box";
import { Paper } from "@material-ui/core";
import Form from "./Form";

const styles = theme => ({});

class Create extends React.Component {
  onAdd = async values => {
    const { app, history, enqueueSnackbar } = this.props;
    await app.execPlans.create(values);
    history.push("/exec-plans");
    enqueueSnackbar("Successfully created plan.", { variant: "success" });
  };

  render() {
    const { app } = this.props;
    return (
      <div>
        <Typography component="h1" variant="h5">
          Add a new Exec Plan
        </Typography>
        <Box component={Paper} p={3} mt={2}>
          <Form onSubmit={this.onAdd} loadHosts={() => app.hosts.fetchList()} />
        </Box>
      </div>
    );
  }
}

export default withSnackbar(
  withStyles(styles)(inject("app")(observer(Create)))
);
