import axios from "axios";
import { types as t, getParentOfType, flow } from "mobx-state-tree";
import {BaseModel, ModelStore} from './base'


const apiClient = axios.create({
  baseURL: "http://localhost:8080/v1",
  withCredentials: true
});

const AuthStore = t
  .model("AuthStore", {})
  .views(self => ({
    get appStore() {
      return getParentOfType(self, AppStore);
    }
  }))
  .actions(self => ({
    login: flow(function*(data) {
      self.state = "pending";
      try {
        yield axios.post("http://localhost:8080/auth/login", data, {
          withCredentials: true
        });
        self.state = "done";
      } catch (error) {
        console.error("Failed to fetch projects", error);
        self.state = "error";
      }
    })
  }));


const Host = BaseModel.named("HostModel").props({
  name: t.string,
  address: t.string,
  pem: t.maybe(t.string)
});

const HostStore = ModelStore.named("HostStore").props({
  list: t.array(t.reference(Host)),
  entities: t.map(Host)
});

const ExecPlanModel = BaseModel.named("ExecPlanModel").props({
  name: t.optional(t.string, ""),
  cmdName: t.string,
  args: t.string,
  planHosts: t.maybeNull(
    t.array(t.late(() => ExecutionPlanHostModel))
  )
});

const ExecPlanStore = ModelStore.named("ExecPlanStore").props({
  list: t.array(t.reference(ExecPlanModel)),
  entities: t.map(ExecPlanModel)
});

const ExecPlanRunModel = BaseModel.named("ExecPlanRunModel").props({
  runStatuses: t.maybeNull(t.array(t.late(() => RunStatusModel))),
  executionPlan: t.late(() => ExecPlanModel)
});

const ExecutionPlanHostModel = BaseModel.named("ExecutionPlanHostModel").props({
  targetHost: Host
});

const ExecPlanRunStore = ModelStore.named("ExecPlanRunStore").props({
  list: t.array(t.reference(ExecPlanRunModel)),
  entities: t.map(ExecPlanRunModel)
});

const RunStatusModel = BaseModel.named("RunStatusModel").props({
  cmd: t.string,
  stdout: t.string,
  stderr: t.string,
  complete: t.boolean,
  runtime: t.number,
  exitCode: t.number,
  executionPlanHost: t.late(() => ExecutionPlanHostModel)
});

// const RunStatusModel = BaseModel.named("RunStatusModel").props({
//   executionPlan: t.optional(ExecPlanModel),
//   cmdName: t.string,
//   args: t.string,
//   planHosts: t.frozen()
// });

const AppStore = t.model("AppStore", {
  auth: t.late(() => AuthStore),
  hosts: t.late(() => HostStore),
  execPlans: t.late(() => ExecPlanStore),
  execPlanRuns: t.late(() => ExecPlanRunStore)
});

export const appStore = AppStore.create({
  auth: {},
  hosts: {
    basePath: "/hosts/",
    list: []
  },
  execPlans: {
    basePath: "/exec-plans/",
    list: []
  },
  execPlanRuns: {
    basePath: "/exec-plan-runs/",
    list: []
  }
}, {
  apiClient
});
