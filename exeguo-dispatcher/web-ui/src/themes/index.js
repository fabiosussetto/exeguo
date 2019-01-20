import { createMuiTheme } from "@material-ui/core/styles";

const overrides = {
  MuiToolbar: {
    root: {
      minHeight: "20px"
    }
  },
  MuiCard: {
    root: {
      boxShadow: "0px 1px 1px 1px rgba(0,0,0,0.05);"
      // boxShadow: {
      // }
    }
  },
  MuiPaper: {
    elevation2: {
      boxShadow: "0px 1px 1px 1px rgba(0,0,0,0.05);"
    }
  },
  MuiButton: {
    contained: {
      boxShadow: {
        boxShadow: "none"
      }
    },
    text: {
      textTransform: "lowercase" // Some CSS
    }
  }
};

const palette = {
  primary: { main: "#3f51b5" },
  secondary: { main: "#f50057" }
};

const themeName = "San Marino Razzmatazz Gaur";

// export default createMuiTheme({
//   palette: {
//     primary: {
//       light: "#757ce8",
//       main: "#3f50b5",
//       dark: "#002884",
//       contrastText: "#fff"
//     },
//     secondary: {
//       light: "#ff7961",
//       main: "#f44336",
//       dark: "#ba000d",
//       contrastText: "#000"
//     }
//   }
// });

export default createMuiTheme({
  overrides,
  palette,
  themeName,
  shape: {
    borderRadius: 3
  },
  typography: { useNextVariants: true }
});
