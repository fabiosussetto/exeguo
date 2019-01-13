import React from "react";
import { observer, inject } from "mobx-react";
import { Typography, Paper } from "@material-ui/core";
import { unstable_Box as Box } from "@material-ui/core/Box";

// import { withStyles } from "@material-ui/core/styles";

class HostDetail extends React.Component {
  componentDidMount() {
    const { app, match } = this.props;

    app.hosts.fetch(match.params.id);
  }

  render() {
    const { app, match } = this.props;
    const host = app.hosts.getById(match.params.id);

    if (!host) {
      return <div>Loading</div>;
    }

    return (
      <div>
        <Typography component="h2" variant="h4" gutterBottom>
          {host.name}
        </Typography>
        <Typography
          component="h3"
          variant="h6"
          gutterBottom
          color="textSecondary"
        >
          {host.address}
        </Typography>
        <Box component={Paper} padding={2} elevation={0} borderColor="grey.500">
          <pre>{host.pem}</pre>
        </Box>
      </div>
    );
  }
}

export default inject("app")(observer(HostDetail));
