import React from "react";
import Button from "@material-ui/core/Button";

import { Formik, Field, FieldArray, ErrorMessage } from "formik";
import * as Yup from "yup";
import {
  TextField,
  Select,
  MenuItem,
  IconButton,
  Chip
} from "@material-ui/core";

import { Delete as DeleteIcon, Add as AddIcon } from "@material-ui/icons";
import { unstable_Box as Box } from "@material-ui/core/Box";

const schema = Yup.object().shape({
  name: Yup.string().required("Required"),
  cmdName: Yup.string().required("Required"),
  args: Yup.string(),
  planHosts: Yup.array()
    .of(
      Yup.object({
        targetHostId: Yup.string().required("Required")
      })
    )
    .required()
});

const MuiTextField = ({ name, label, margin = "normal", ...rest }) => {
  return (
    <Field
      name={name}
      render={({ field, form: { touched, errors } }) => (
        <TextField
          label={label}
          margin={margin}
          helperText={touched[field.name] && errors[field.name]}
          error={Boolean(touched[field.name] && errors[field.name])}
          {...field}
          {...rest}
        />
      )}
    />
  );
};

const initials = {
  name: "ls",
  cmdName: "",
  args: "",
  // planHosts: [{ targetHostId: 10 }, { targetHostId: 20 }]
  planHosts: [{ targetHostId: "" }]
};

class Form extends React.Component {
  state = {
    hostOptions: [],
    hostOptionsLoaded: false
  };

  loadHostOptions = async () => {
    const { loadHosts } = this.props;
    const { hostOptionsLoaded } = this.state;

    if (hostOptionsLoaded) {
      return;
    }

    const hosts = await loadHosts();
    this.setState({
      hostOptions: hosts,
      hostOptionsLoaded: true
    });
  };

  render() {
    const { onSubmit } = this.props;
    const { hostOptions } = this.state;

    return (
      <Formik
        validationSchema={schema}
        onSubmit={onSubmit}
        initialValues={initials}
        render={({ handleSubmit, values }) => (
          <form onSubmit={handleSubmit}>
            <MuiTextField name="name" label="name" fullWidth />
            <MuiTextField name="cmdName" label="Command" fullWidth />
            <MuiTextField name="args" label="Args" fullWidth />

            <FieldArray
              name="planHosts"
              render={arrayHelpers => (
                <div>
                  <h4>Target HostList</h4>
                  {values.planHosts && values.planHosts.length > 0 ? (
                    <Box width="500px">
                      {values.planHosts.map((friend, index) => (
                        <Box
                          key={index}
                          display="flex"
                          alignItems="center"
                          marginBottom={2}
                        >
                          <Box flex="0 0 400px">
                            <Field
                              name={`planHosts.${index}.targetHostId`}
                              render={({
                                field,
                                form: { touched, errors }
                              }) => (
                                <React.Fragment>
                                  <Select
                                    {...field}
                                    MenuProps={{
                                      onEnter: this.loadHostOptions
                                    }}
                                    fullWidth
                                    error={Boolean(
                                      touched[field.name] && errors[field.name]
                                    )}
                                    renderValue={value => {
                                      const selectedHost = hostOptions.find(
                                        h => h.id === value
                                      );
                                      return (
                                        <Box display="flex" alignItems="center">
                                          <Box>{selectedHost.name}</Box>
                                          <Box
                                            marginLeft="auto"
                                            component={Chip}
                                            label={selectedHost.address}
                                            variant="outlined"
                                          />
                                        </Box>
                                      );
                                    }}
                                  >
                                    <MenuItem value="" disabled>
                                      Select Host...
                                    </MenuItem>
                                    {hostOptions.map(host => (
                                      <Box
                                        component={MenuItem}
                                        key={host.id}
                                        value={host.id}
                                        display="flex"
                                      >
                                        <Box>{host.name}</Box>
                                        <Box
                                          marginLeft="auto"
                                          component={Chip}
                                          label={host.address}
                                          variant="outlined"
                                        />
                                      </Box>
                                    ))}
                                  </Select>
                                  <ErrorMessage name={field.name} />
                                </React.Fragment>
                              )}
                            />
                          </Box>

                          <Box marginLeft="auto">
                            <IconButton
                              onClick={() =>
                                arrayHelpers.insert(index + 1, {
                                  targetHostId: ""
                                })
                              }
                            >
                              <AddIcon />
                            </IconButton>
                            <IconButton
                              onClick={() => arrayHelpers.remove(index)}
                            >
                              <DeleteIcon />
                            </IconButton>
                          </Box>
                        </Box>
                      ))}
                    </Box>
                  ) : (
                    <IconButton
                      onClick={() => arrayHelpers.push({ targetHostId: "" })}
                    >
                      <AddIcon />
                    </IconButton>
                  )}
                </div>
              )}
            />

            <Button type="submit" variant="contained" color="primary">
              Add Plan
            </Button>
          </form>
        )}
      />
    );
  }
}

export default Form;
