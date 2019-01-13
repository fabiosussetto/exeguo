import React from "react";
import { observer, inject } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import { withSnackbar } from "notistack";

import { unstable_Box as Box } from "@material-ui/core/Box";
import { Paper } from "@material-ui/core";
import HostForm from "./HostForm";

const styles = theme => ({});

class HostAdd extends React.Component {
  onAdd = async values => {
    const { app, history, enqueueSnackbar } = this.props;
    console.log(values);
    await app.hosts.create(values);
    history.push("/hosts");
    enqueueSnackbar("Successfully created host.", { variant: "success" });
  };

  render() {
    return (
      <div>
        <Typography component="h1" variant="h5">
          Add a new Host
        </Typography>
        <Box component={Paper} p={3} mt={2}>
          <HostForm
            onSubmit={this.onAdd}
            initialValues={{ name: "Fooo", address: "localhost" }}
          />
        </Box>
      </div>
    );
  }
}

export default withSnackbar(
  withStyles(styles)(inject("app")(observer(HostAdd)))
);
