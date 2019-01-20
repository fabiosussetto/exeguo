import React, { Component } from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { Provider } from "mobx-react";
import CssBaseline from "@material-ui/core/CssBaseline";
import { SnackbarProvider } from "notistack";
import { MuiThemeProvider } from "@material-ui/core/styles";
import Main from "./components/layout/Main";
import { appStore } from "./stores";
import theme from "./themes";

class App extends Component {
  render() {
    return (
      <Router>
        <Provider app={appStore}>
          <React.Fragment>
            <CssBaseline />
            <MuiThemeProvider theme={theme}>
              <SnackbarProvider maxSnack={3}>
                <Main />
              </SnackbarProvider>
            </MuiThemeProvider>
          </React.Fragment>
        </Provider>
      </Router>
    );
  }
}

export default App;
