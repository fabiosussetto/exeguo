import React from "react";

import { observer, inject } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import Chip from "@material-ui/core/Chip";
import IconButton from "@material-ui/core/IconButton";
import { withSnackbar } from "notistack";

import IconPlus from "@material-ui/icons/AddOutlined";
import DeleteIcon from "@material-ui/icons/Delete";
import PlayCircleOutlinedIcon from "@material-ui/icons/PlayArrow";
import { unstable_Box as Box } from "@material-ui/core/Box";
import { Link } from "react-router-dom";

import {
  TableBody,
  TableRow,
  TableCell,
  TableHead,
  Table,
  Paper
} from "@material-ui/core";

// import Typography from '@material-ui/core/Typography';

const styles = theme => ({});

class List extends React.Component {
  componentDidMount() {
    this.props.app.execPlans.fetchList();
  }

  onItemDelete = async plan => {
    const { enqueueSnackbar } = this.props;
    await plan.delete();

    plan.store.fetchList();
    enqueueSnackbar("Successfully deleted plan.", { variant: "success" });
  };

  render() {
    const {
      app: { execPlans }
    } = this.props;

    return (
      <div>
        <Box mb={3} display="flex">
          <Typography component="h1" variant="h5" noWrap>
            Execution Plans
          </Typography>
          <Box ml="auto">
            <Button
              variant="outlined"
              color="primary"
              component={Link}
              to="/exec-plans/add"
            >
              <IconPlus />
              Add Exec Plan
            </Button>
          </Box>
        </Box>
        <Paper>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Command</TableCell>
                <TableCell>Hosts</TableCell>
                <TableCell />
              </TableRow>
            </TableHead>
            <TableBody>
              {execPlans.list.map(plan => (
                <TableRow key={plan.id}>
                  <TableCell>{plan.name}</TableCell>
                  <TableCell>
                    <Box bgcolor="text.primary" color="background.paper" p={1}>
                      {plan.cmdName} {plan.args}
                    </Box>
                  </TableCell>
                  <TableCell>
                    {plan.planHosts.map(planHost => (
                      <Chip
                        variant="outlined"
                        key={planHost.id}
                        label={planHost.targetHost.name}
                      />
                    ))}
                  </TableCell>
                  <TableCell align="right">
                    <IconButton aria-label="Run">
                      <PlayCircleOutlinedIcon />
                    </IconButton>
                    <IconButton aria-label="Delete">
                      <DeleteIcon
                        onClick={this.onItemDelete.bind(this, plan)}
                      />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>
      </div>
    );
  }
}

export default withSnackbar(withStyles(styles)(inject("app")(observer(List))));
