const axios = require("axios").default;
const _ = require("lodash");

export default class {
  constructor(host, port, vue) {
    this.clientID = null;
    this.ws = null;
    this.host = host;
    this.port = port;
    this.eventFuncs = new Map();
    this.retry = 0;
    this.vue = vue;
  }

  async subscript(event, ...funcs) {
    // funcs.push((e) => {
    // 	this.vue.commit('receive', e)
    // })
    let fs = this.eventFuncs.get(event.type);
    if (fs && _.isArray(fs)) {
      fs.append(...funcs);
    } else {
      this.eventFuncs.set(event.type, funcs);
    }
    let data = { eventType: event.type };
    _.merge(data, { clientID: this.clientID });
    const result = await axios({
      method: "POST",
      url: `http://${this.host}:${this.port}/subscript`,
      contentType: "application/json",
      data: JSON.stringify(data)
    });
    if (result.data.success) {
      this.clientID = result.data.clientID;
      this.token = result.data.token;
    }
  }

  receive(event) {
    const fs = this.eventFuncs.get(event.type);
    if (_.isArray(fs)) {
      fs.forEach(f => {
        f(event);
      });
    }
  }

  eventTunnel() {
    if (_.isEmpty(this.token)) {
      throw "token empty";
    }
    if (!_.isEmpty(this.ws)) {
      throw "ws is already connected";
    }
    this.ws = new WebSocket(
      `ws://${this.host}:${this.port}/event_tunnel?token=${this.token}`,
      this.port
    );

    this.ws.onopen = () => {
      var cnt = 1;
      setInterval(() => {
        cnt++;
        this.ws.send(
          JSON.stringify({
            type: "main.EventTest",
            message: cnt.toString(),
            from: this.clientID
          })
        );
      }, 1000);
    };

    this.ws.onmessage = data => {
      console.log("receive:", data);
      this.receive(event);
    };

    this.ws.onclose = event => {
      console.log("close", event);
      this.ws = null;
      if (this.retry < 3) {
        this.retry++;
        console.log("retry connect");
        this.eventTunnel();
      }
    };
  }
}
