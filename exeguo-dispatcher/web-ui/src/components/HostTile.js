import React from "react";
import { observer } from "mobx-react";
import { withStyles } from "@material-ui/core/styles";
import Card from "@material-ui/core/Card";
import { withSnackbar } from "notistack";
import { withRouter } from "react-router-dom";

import { Link } from "react-router-dom";
import MoreVertIcon from "@material-ui/icons/MoreVert";
import IconButton from "@material-ui/core/IconButton";
import OfflineBoltOutlined from "@material-ui/icons/OfflineBoltOutlined";
import CardHeader from "@material-ui/core/CardHeader";

import {
  Menu,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  DialogContentText
} from "@material-ui/core";
import PopupState, { bindTrigger, bindMenu } from "material-ui-popup-state";

const styles = {
  card: {
    minWidth: 275
  },
  bullet: {
    display: "inline-block",
    margin: "0 2px",
    transform: "scale(0.8)"
  },
  title: {
    fontSize: 14
  },
  pos: {
    marginBottom: 12
  }
};

class SimpleCard extends React.Component {
  state = {
    confirmDeleteDialogOpen: false
  };

  toggleConfirm = () => {
    this.setState({
      confirmDeleteDialogOpen: !this.state.confirmDeleteDialogOpen
    });
  };

  onDelete = async popupState => {
    const { host, history, enqueueSnackbar } = this.props;
    popupState.close();
    await host.delete();

    this.toggleConfirm();
    host.store.fetchList();
    history.push("/hosts");
    enqueueSnackbar("Successfully deleted host.", { variant: "success" });
  };

  render() {
    const { host, classes = {} } = this.props;
    const { confirmDeleteDialogOpen } = this.state;

    return (
      <Card className={classes.card}>
        <PopupState variant="popover" popupId="demoMenu">
          {popupState => (
            <React.Fragment>
              <CardHeader
                avatar={<OfflineBoltOutlined />}
                action={
                  <div onClick={e => e.preventDefault()}>
                    <IconButton {...bindTrigger(popupState)}>
                      <MoreVertIcon />
                    </IconButton>
                    <Menu {...bindMenu(popupState)}>
                      <MenuItem onClick={this.toggleConfirm}>Delete</MenuItem>
                    </Menu>
                  </div>
                }
                title={<Link to={`/hosts/${host.id}`}>{host.name}</Link>}
                subheader={host.address}
              />

              <Dialog
                open={confirmDeleteDialogOpen}
                onClose={this.toggleConfirm}
              >
                <DialogTitle>Confirm delete?</DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    Do you really want to delete host{" "}
                    <strong>"{host.name}"</strong>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button
                    onClick={this.onDelete.bind(null, popupState)}
                    color="primary"
                  >
                    Confirm
                  </Button>
                  <Button onClick={this.toggleConfirm}>Cancel</Button>
                </DialogActions>
              </Dialog>
            </React.Fragment>
          )}
        </PopupState>
      </Card>
    );
  }
}

export default withSnackbar(
  withRouter(withStyles(styles)(observer(SimpleCard)))
);
