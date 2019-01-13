import axios from "axios";
import _ from "lodash";
import { types, getParentOfType, flow } from "mobx-state-tree";

const AuthStore = types
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

const Host = types
  .model("HostModel", {
    id: types.identifierNumber,
    name: types.string,
    address: types.string,
    pem: types.string
  })
  .views(self => ({
    get store() {
      return getParentOfType(self, HostStore);
    }
  }))
  .actions(self => ({
    delete() {
      return self.store.delete(self);
    }
  }));

const HostStore = types
  .model("HostStore", {
    list: types.array(types.reference(Host)),
    entities: types.map(Host)
  })
  .views(self => ({
    get appStore() {
      return getParentOfType(self, AppStore);
    },
    getById(id) {
      return self.entities.get(id);
    }
  }))
  .actions(self => ({
    fetchList: flow(function*() {
      self.state = "pending";
      try {
        const resp = yield axios.get(
          "http://localhost:8080/v1/hosts/",
          {},
          {
            withCredentials: true
          }
        );
        self.entities.merge(_.keyBy(resp.data, "id"));

        self.list = resp.data.map(h => h.id);
      } catch (error) {
        console.error("Failed to fetch projects", error);
      }
    }),
    fetch: flow(function*(id) {
      try {
        const resp = yield axios.get(`http://localhost:8080/v1/hosts/${id}`, {
          withCredentials: true
        });

        self.entities.put(resp.data);
      } catch (error) {
        console.error("Failed to fetch projects", error);
      }
    }),
    create: flow(function*(data) {
      try {
        yield axios.post("http://localhost:8080/v1/hosts/", data, {
          withCredentials: true
        });
      } catch (error) {
        console.error("Failed to fetch projects", error);
      }
    }),
    delete: flow(function*(host) {
      try {
        yield axios.delete(`http://localhost:8080/v1/hosts/${host.id}`, {
          withCredentials: true
        });
      } catch (error) {
        console.error("Failed to fetch projects", error);
      }
    })
  }));

const AppStore = types.model("AppStore", {
  auth: types.late(() => AuthStore),
  hosts: types.late(() => HostStore)
});

export const appStore = AppStore.create({
  auth: {},
  hosts: {
    list: []
  }
});
