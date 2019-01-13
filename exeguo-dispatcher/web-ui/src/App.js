import React, { Component } from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { Provider } from "mobx-react";
import CssBaseline from "@material-ui/core/CssBaseline";
import { SnackbarProvider } from "notistack";
import Main from "./components/layout/Main";
import { appStore } from "./stores";

class App extends Component {
  render() {
    return (
      <Router>
        <Provider app={appStore}>
          <>
            <CssBaseline />
            <SnackbarProvider maxSnack={3}>
              <Main />
            </SnackbarProvider>
          </>
        </Provider>
      </Router>
    );
  }
}

export default App;
