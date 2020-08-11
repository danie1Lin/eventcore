const axios = require("axios").default;
const _ = require("lodash");
const request = axios.create({
  headers: { "Content-Type": "application/json" }
});

const state = () => ({
  clientID: null,
  ws: null,
  eventFuncs: new Map(),
  retry: 0,
  events: [
    {
      Type: "main.EventTest",
      Traces: [
        { Name: "WebSocket", Type: "WebSocket", ID: "0000" },
        { Name: "EventCluster", Type: "EventCluster", ID: "0010" }
      ]
    }
  ],
  token: null,
  wsState: "closed",
  infos: [],
  loading: false,
  eventTypes: ["main.EventTest"],
  eventInfoMap: new Map()
});

const getters = {
  allEvents: state => {
    return state.events;
  },
  groupEventByType: state => type => {
    return state.events.find(e => e.type == type);
  },
  allEventTypes: state => {
    return [...state.eventInfoMap.keys()].map(k => {
      let e = { isSubscribed: false, type: k, text: k, value: k };
      if (state.eventFuncs.has(k)) {
        e.isSubscribed = true;
      }
      return e;
    });
  }
};

const actions = {
  async login({ commit, state }) {
    const result = await request.post(
      `http://${state.host}:${state.port}/login`
    );
    if (result?.data?.clientID) {
      commit("setClientID", result.data.clientID);
    }
  },
  async getToken({ commit, state }) {
    const result = await request.get(
      `http://${state.host}:${state.port}/eventTunnel/token`,
      {
        headers: { "X-Client-ID": state.clientID }
      }
    );
    if (result?.data?.token) {
      commit("setToken", result.data.token);
    }
  },
  async getEventList({ commit, state }) {
    const result = await request
      .get(`http://${state.host}:${state.port}/events`, {
        headers: { "X-Client-ID": state.clientID }
      })
      .catch(e => {
        commit("addError", e);
        commit("toggleLoading");
        throw e;
      });
    if (result?.data && _.isArray(result.data)) {
      const list = result.data.reduce((map, eventInfo) => {
        return map.set(eventInfo.Type, eventInfo);
      }, new Map());
      commit("updateEventLists", list);
    }
  },

  async getSubscript({ state, commit }) {
    const result = await request
      .get(`http://${state.host}:${state.port}/subscript`, {
        headers: { "X-Client-ID": state.clientID }
      })
      .catch(e => {
        commit("addError", e);
        commit("toggleLoading");
        throw e;
      });
    if (result?.data && _.isArray(result.data)) {
      commit("updateSubscription", result.data);
    }
  },
  async unsubscript({ commit, state }, type) {
    const data = { eventType: type };
    const result = await request
      .post(`http://${state.host}:${state.port}/subscript/cancel`, data, {
        headers: { "X-Client-ID": state.clientID }
      })
      .catch(e => {
        commit("addError", e);
        commit("toggleLoading");
        throw e;
      });
    if (result?.data?.success) {
      commit("removeSubscription", type);
    }
  },
  async subscript({ commit, state }, { type, cb }) {
    const data = { eventType: type };
    const result = await request
      .post(`http://${state.host}:${state.port}/subscript`, data, {
        headers: { "X-Client-ID": state.clientID }
      })
      .catch(e => {
        commit("addError", e);
        commit("toggleLoading");
        throw e;
      });
    if (result?.data?.success) {
      commit("addSubscription", {
        type,
        cb
      });
    }
  },
  eventTunnel({ dispatch, commit, state }) {
    try {
      if (_.isEmpty(state.token)) {
        throw new Error("token empty");
      }
      if (!_.isEmpty(state.ws)) {
        throw new Error("ws is already connected");
      }
      dispatch(
        "InitWs",
        new WebSocket(
          `ws://${state.host}:${state.port}/event_tunnel?token=${state.token}`
        )
      );
    } catch (e) {
      commit("addError", e);
      commit("toggleLoading");
      throw e;
    } finally {
      commit("setToken", null);
    }
  },
  InitWs({ dispatch, commit }, ws) {
    ws.onopen = () => {
      commit("setWsState", "opened");
    };
    ws.onmessage = resp => {
      const event = JSON.parse(resp.data);
      commit("setWsState", "messaging");
      commit("receiveEvent", event);
      dispatch("emit", event);
    };
    ws.onclose = () => {
      commit("setWsState", "closed");
    };
    commit("setWs", ws);
  },
  emit({ state }, event) {
    const fs = state.eventFuncs.get(event.Type);
    if (fs && _.isArray(fs)) {
      fs.forEach(f => {
        f(event);
      });
    }
  },
  sendEvent({ state }, event) {
    state.ws.send(JSON.stringify(event));
  },
  close({ state, commit }) {
    state.ws.close();
    commit("clearSubscription");
  }
};

const mutations = {
  updateEventLists(state, eventList) {
    state.eventInfoMap = eventList;
  },
  cleanEvents(state) {
    state.events = [];
  },
  addError(state, e) {
    state.infos.unshift({ type: "error", message: e.message, show: true });
  },
  dismissError(state, index) {
    state.infos.deleteAt(index);
  },
  toggleLoading(state) {
    state.loading = !state.loading;
  },
  setLoading(state, v) {
    state.loading = v;
  },
  setWs(state, ws) {
    state.ws = ws;
  },
  receiveEvent(state, event) {
    state.events.unshift(event);
  },
  setWsState(state, wsState) {
    state.wsState = wsState;
    state.loading = false;
  },
  initClient(state, { host, port }) {
    state.host = host;
    state.port = port;
  },
  setClientID(state, clientID) {
    state.clientID = clientID;
  },
  setToken(state, token) {
    state.token = token;
  },
  addSubscription(state, { type, cb }) {
    let fs = state.eventFuncs.get(type);
    if (fs && _.isArray(fs)) {
      fs.push(cb);
      state.eventFuncs = new Map(state.eventFuncs.set(type, fs));
    } else {
      state.eventFuncs = new Map(state.eventFuncs.set(type, [cb]));
    }
  },
  removeSubscription(state, type) {
    if (!state.eventFuncs.delete(type)) {
      throw new Error("not deleted");
    }
    state.eventFuncs = new Map(state.eventFuncs);
  },
  updateSubscription() {},
  clearSubscription(state) {
    state.eventFuncs = new Map();
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
