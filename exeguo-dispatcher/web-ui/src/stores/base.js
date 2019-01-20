import _ from "lodash";
import { types as t, flow, getParent, getEnv } from "mobx-state-tree";

export const BaseModel = t
  .model("BaseModel", {
    id: t.identifierNumber
  })
  .views(self => ({
    get store() {
      // return getParentOfType(self, self.storeType);
      return getParent(self, 2);
    }
  }))
  .actions(self => ({
    delete() {
      return self.store.delete(self);
    }
  }));

export const ModelStore = t
  .model("ModelStore", {
    basePath: t.string
  })
  .views(self => ({
    // get appStore() {
    //   return getParentOfType(self, AppStore);
    // },
    get apiClient () {
      return getEnv(self).apiClient
    },
    getById(id) {
      return self.entities.get(id);
    },
    getListUrl() {
      return self.basePath;
    },
    getDetailUrl(pk) {
      return `${self.basePath}${pk}`;
    }
  }))
  .actions(self => ({
    fetchList: flow(function*() {
      const resp = yield self.apiClient.get(self.getListUrl());
      self.entities.merge(_.keyBy(resp.data, "id"));
      self.list = resp.data.map(h => h.id);
      return resp.data;
    }),
    fetch: flow(function*(id) {
      const resp = yield self.apiClient.get(self.getDetailUrl(id));
      self.entities.put(resp.data);
    }),
    create: flow(function*(data) {
      const resp = yield self.apiClient.post(self.getListUrl(), data);
      return resp.data;
    }),
    delete: flow(function*(instance) {
      yield self.apiClient.delete(self.getDetailUrl(instance.id));
    })
  }));
