import React from "react";
import { observer, inject } from "mobx-react";
import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";

import LockIcon from "@material-ui/icons/LockOutlined";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";
import withStyles from "@material-ui/core/styles/withStyles";
import { Formik, Field } from "formik";
import * as Yup from "yup";
import { TextField } from "@material-ui/core";

const LoginSchema = Yup.object().shape({
  email: Yup.string()
    .email("Invalid email")
    .required("Required"),
  password: Yup.string().required("Required")
});

const styles = theme => ({
  main: {
    width: "auto",
    display: "block", // Fix IE 11 issue.
    marginLeft: theme.spacing.unit * 3,
    marginRight: theme.spacing.unit * 3,
    [theme.breakpoints.up(400 + theme.spacing.unit * 3 * 2)]: {
      width: 400,
      marginLeft: "auto",
      marginRight: "auto"
    }
  },
  paper: {
    marginTop: theme.spacing.unit * 8,
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    padding: `${theme.spacing.unit * 2}px ${theme.spacing.unit * 3}px ${theme
      .spacing.unit * 3}px`
  },
  avatar: {
    margin: theme.spacing.unit,
    backgroundColor: theme.palette.secondary.main
  },
  form: {
    width: "100%", // Fix IE 11 issue.
    marginTop: theme.spacing.unit
  },
  submit: {
    marginTop: theme.spacing.unit * 3
  }
});

const LoginForm = withStyles(styles)(
  ({ handleSubmit, handleChange, handleBlur, values, errors, classes }) => (
    <form className={classes.form} onSubmit={handleSubmit}>
      <Field
        name="email"
        render={({ field, form: { touched, errors } }) => (
          <TextField
            {...field}
            label="Email"
            fullWidth
            margin="normal"
            helperText={touched[field.name] && errors[field.name]}
            error={Boolean(touched[field.name] && errors[field.name])}
          />
        )}
      />
      <Field
        name="password"
        render={({ field, form: { touched, errors } }) => (
          <TextField
            {...field}
            label="Password"
            margin="normal"
            fullWidth
            helperText={touched[field.name] && errors[field.name]}
            error={Boolean(touched[field.name] && errors[field.name])}
          />
        )}
      />

      <Button
        type="submit"
        fullWidth
        variant="contained"
        color="primary"
        className={classes.submit}
      >
        Sign in
      </Button>
    </form>
  )
);

class SignIn extends React.Component {
  onSubmit = (values, { setSubmitting }) => {
    this.props.app.auth.login(values);
  };

  render() {
    const { classes } = this.props;

    return (
      <main className={classes.main}>
        <Paper className={classes.paper}>
          <Avatar className={classes.avatar}>
            <LockIcon />
          </Avatar>
          <Typography component="h1" variant="h5">
            Sign in
          </Typography>

          <Formik
            component={LoginForm}
            initialValues={{ email: "fabio@test.com", password: "$Test1234" }}
            validationSchema={LoginSchema}
            onSubmit={this.onSubmit}
          />
        </Paper>
      </main>
    );
  }
}

export default withStyles(styles)(inject("app")(observer(SignIn)));
