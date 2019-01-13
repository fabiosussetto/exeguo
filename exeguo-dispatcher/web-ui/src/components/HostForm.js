import React from "react";
import Button from "@material-ui/core/Button";

// import withStyles from "@material-ui/core/styles/withStyles";
import { Formik, Field } from "formik";
import * as Yup from "yup";
import { TextField } from "@material-ui/core";

const schema = Yup.object().shape({
  email: Yup.string()
    .email("Invalid email")
    .required("Required"),
  password: Yup.string().required("Required")
});

const HostForm = props => (
  <Formik
    initialValues={{ name: "", address: "" }}
    {...props}
    render={props => (
      <form onSubmit={props.handleSubmit}>
        <Field
          name="name"
          render={({ field, form: { touched, errors } }) => (
            <TextField
              {...field}
              label="Name"
              fullWidth
              margin="normal"
              helperText={touched[field.name] && errors[field.name]}
              error={Boolean(touched[field.name] && errors[field.name])}
            />
          )}
        />
        <Field
          name="address"
          render={({ field, form: { touched, errors } }) => (
            <TextField
              {...field}
              label="Address"
              margin="normal"
              fullWidth
              helperText={touched[field.name] && errors[field.name]}
              error={Boolean(touched[field.name] && errors[field.name])}
            />
          )}
        />

        <Button type="submit" variant="contained" color="primary">
          Add Host
        </Button>
      </form>
    )}
  />
);

export default HostForm;
