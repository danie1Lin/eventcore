const axios = require("axios").default;
const _ = require("lodash");

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
  eventTypes: ["main.EventTest"]
});

const getters = {
  allEvents: state => {
    return state.events;
  },
  groupEventByType: state => type => {
    return state.events.find(e => e.type == type);
  }
};

const actions = {
  async subscript({ commit, state }, { type, cb }) {
    const data = { eventType: type, clientID: state.clientID };
    const result = await axios({
      method: "POST",
      url: `http://${state.host}:${state.port}/subscript`,
      contentType: "application/json",
      data: JSON.stringify(data)
    }).catch(e => {
      commit("addError", e);
      commit("toggleLoading");
      throw e;
    });
    if (result?.data?.success) {
      commit("setClientID", result.data.clientID);
      commit("setToken", result.data.token);
      commit("addSubscription", {
        type,
        cb
      });
    }
    commit("setLoading", true);
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
    state.ws.send(
      JSON.stringify(
        _.merge(event, {
          traces: [
            {
              name: "websocket_client",
              type: "websocket_client",
              id: state.clientID
            }
          ]
        })
      )
    );
  },
  close({ state }) {
    state.ws.close();
  }
};

const mutations = {
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
      fs.append(cb);
    } else {
      state.eventFuncs.set(type, [cb]);
    }
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
