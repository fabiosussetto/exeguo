import React from "react";

import { observer, inject } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import IconPlus from "@material-ui/icons/AddOutlined";
import { unstable_Box as Box } from "@material-ui/core/Box";
import { Link } from "react-router-dom";

import HostTile from "./HostTile";

// import Typography from '@material-ui/core/Typography';

const styles = theme => ({});

class HostList extends React.Component {
  componentDidMount() {
    this.props.app.hosts.fetchList();
  }

  render() {
    const {
      app: { hosts }
    } = this.props;

    return (
      <div>
        <Box mb={3} display="flex">
          <Typography component="h1" variant="h5" noWrap>
            Hosts
          </Typography>
          <Box ml="auto">
            <Button
              variant="outlined"
              color="primary"
              component={Link}
              to="/hosts/add"
            >
              <IconPlus />
              Add Host
            </Button>
          </Box>
        </Box>
        <Grid container spacing={24}>
          {hosts.list.map(host => (
            <Grid item xs={12} sm={4} key={host.id}>
              <HostTile host={host} />
            </Grid>
          ))}
        </Grid>
      </div>
    );
  }
}

export default withStyles(styles)(inject("app")(observer(HostList)));
