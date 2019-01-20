import React from "react";

import { observer, inject } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import { withSnackbar } from "notistack";

import { unstable_Box as Box } from "@material-ui/core/Box";
import { Paper } from "@material-ui/core";

const styles = theme => ({});

class Detail extends React.Component {
  async componentDidMount() {
    const {
      app,
      match: { params }
    } = this.props;

    app.execPlanRuns.fetch(params.id);

    setInterval(() => {
      app.execPlanRuns.fetch(params.id);
    }, 1000);
  }

  render() {
    const {
      match,
      app: { execPlanRuns }
    } = this.props;

    const model = execPlanRuns.getById(match.params.id);

    if (!model) {
      return <div>Loading</div>;
    }

    return (
      <div>
        <Box mb={3} display="flex">
          <Typography component="h1" variant="h5" noWrap>
            Execution Plan Run
          </Typography>
        </Box>
        <Paper>
          <Typography>
            Run #{model.id} - {model.executionPlan.name}
          </Typography>
        </Paper>
        {model.runStatuses.map(runStatus => (
          <Paper key={runStatus.id}>
            <Typography>
              Host: {runStatus.executionPlanHost.targetHost.name} (
              {runStatus.executionPlanHost.targetHost.address})
            </Typography>
            <Typography>
              Completed: {runStatus.complete ? "yes" : "no"}
            </Typography>
            <Typography>Runtime: {runStatus.runtime}</Typography>
            <Typography>Exit code: {runStatus.exitCode}</Typography>
            <Box
              css={{
                backgroundColor: "#222",
                color: "white",
                overflow: "auto",
                maxHeight: 400
              }}
              my={3}
              p={2}
            >
              <pre>{runStatus.stdout}</pre>
            </Box>
          </Paper>
        ))}
      </div>
    );
  }
}

export default withSnackbar(
  withStyles(styles)(inject("app")(observer(Detail)))
);
