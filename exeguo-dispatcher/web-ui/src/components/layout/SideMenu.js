import React from "react";
import { Link } from "react-router-dom";
import { withStyles } from "@material-ui/core/styles";
import Drawer from "@material-ui/core/Drawer";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";

import DvrOutlined from "@material-ui/icons/DvrOutlined";
import FlashOnOutlined from "@material-ui/icons/FlashOnOutlined";

const drawerWidth = 210;

const styles = theme => ({
  drawer: {
    width: drawerWidth,
    flexShrink: 0
  },
  drawerPaper: {
    width: drawerWidth
  },
  toolbar: theme.mixins.toolbar
});

const MENU_ITEMS = [
  {
    label: "Hosts",
    icon: DvrOutlined,
    to: "/hosts"
  },
  {
    label: "Execution Plans",
    icon: FlashOnOutlined,
    to: "/exec-plans"
  }
];

function SideMenu(props) {
  const { classes } = props;

  return (
    <Drawer
      className={classes.drawer}
      variant="permanent"
      classes={{
        paper: classes.drawerPaper
      }}
    >
      <div className={classes.toolbar} />
      <List>
        {MENU_ITEMS.map(item => {
          const ItemIcon = item.icon;

          return (
            <ListItem button key={item.label} component={Link} to={item.to}>
              <ItemIcon />
              <ListItemText primary={item.label} />
            </ListItem>
          );
        })}
      </List>
    </Drawer>
  );
}

export default withStyles(styles)(SideMenu);
