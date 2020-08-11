<template>
  <v-app>
    <v-app-bar app color="primary" dark>
      <div class="d-flex align-center">
        <h1>
          Event Core
        </h1>
      </div>
      <v-spacer></v-spacer>
      <v-text-field label="host" v-model="host" class="mb-n5"></v-text-field>
      <v-spacer></v-spacer>
      <v-text-field label="port" v-model="port" class="mb-n5"></v-text-field>
      <v-spacer></v-spacer>
      <v-btn
        :loading="loading"
        :disabled="false"
        :color="connectionButton.color"
        v-text="connectionButton.text"
        class="ma-2 white--text"
        @click="connect"
      >
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn
        href="https://github.com/daniel840829/eventcore"
        target="_blank"
        text
      >
        <span class="mr-2">Github</span>
        <v-icon>mdi-open-in-new</v-icon>
      </v-btn>
    </v-app-bar>
    <v-main>
      <v-alert
        v-for="(info, index) in infos"
        :key="index"
        :type="info.type"
        border="left"
        elevation="10"
        dismissible
        transition="v-fade-transition"
        @toggle="dismissError(index)"
      >
        {{ info.message }}
      </v-alert>
      <v-row>
        <v-col cols="6">
          <v-sheet elevation="10">
            <v-list-item-group>
              <v-row>
                <v-col>
                  <v-subheader>Event Recieved</v-subheader>
                </v-col>
                <v-col>
                  <v-btn color="red" @click="cleanEvents">Reset</v-btn>
                </v-col>
              </v-row>
              <v-list-item
                multiple="true"
                v-for="(event, i) in events"
                :key="i"
                @change="expandEventInfo"
              >
                <template v-slot:default="{ active }">
                  <template v-if="!active">
                    <v-list-item-content>
                      <v-list-item-title
                        v-text="event.Type"
                      ></v-list-item-title>
                    </v-list-item-content>
                    <v-list-item-content>
                      <v-list-item-title
                        v-text="event.CostumeField"
                      ></v-list-item-title>
                    </v-list-item-content>
                    <v-list-item-content
                      v-for="trace in event.Traces"
                      :key="trace.ID"
                      v-text="trace.Name + trace.ID"
                    >
                    </v-list-item-content>
                  </template>
                  <v-list-item-content
                    v-if="active"
                    style="white-space: pre-wrap"
                    v-text="JSON.stringify(event, null, '\t')"
                  >
                  </v-list-item-content>
                </template>
              </v-list-item>
            </v-list-item-group>
          </v-sheet>
        </v-col>
        <v-col cols="6">
          <v-row>
            <v-col>
              <v-form ref="form" :lazy-validation="true">
                <v-overflow-btn
                  class="my-2"
                  :items="allEventTypes"
                  label="Event Type"
                  v-model="eventToSend.type"
                >
                </v-overflow-btn>
                <v-textarea
                  label="Event Body"
                  v-model="eventToSend.body"
                  :rules="eventDataRules"
                  @keyup="formatJson"
                ></v-textarea>
                <v-btn color="green" class="mr-4" @click="sendEvent">
                  Send Event
                </v-btn>
              </v-form>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-checkbox
                v-for="(event, i) in allEventTypes"
                :label="event.type"
                :value="event.isSubscribed"
                :key="i"
                @change="toggleSubscription(event.type)"
                >{{ event.type }}</v-checkbox
              >
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-main>
  </v-app>
</template>

<script>
import { mapGetters, mapMutations, mapState, mapActions } from "vuex";
//import JsonEditor from 'vue-json-ui-editor'
export default {
  name: "App",
  //components: { JsonEditor },
  data() {
    return {
      wsLoading: false,
      host: "localhost",
      port: "7000",
      eventInfoExpanded: new Map(),
      eventToSend: {
        body: "",
        type: "main.EventTest"
      },
      eventDataRules: [
        value => {
          try {
            JSON.parse(value);
          } catch (e) {
            return e.message;
          }
          return "";
        }
      ],
      schema: {
        type: "object",
        title: "Event Body",
        properties: {
          name: {
            type: "string"
          },
          email: {
            type: "string"
          }
        }
      }
    };
  },
  computed: {
    connectionButton() {
      let info = {
        color: "red",
        text: "Connect",
        on: false
      };
      switch (this.$store.state.eventCenter.wsState) {
        case "messaging":
          info.color = "blue";
          info.text = "Close";
          info.on = true;
          break;
        case "opened":
          info.color = "green";
          info.text = "Close";
          info.on = true;
          break;
        default:
          info.color = "red";
          info.text = "Connect";
          info.on = false;
      }
      return info;
    },
    connectionState() {
      return this.$store.state.eventCenter.wsState;
    },
    ...mapGetters("eventCenter", ["allEvents", "allEventTypes"]),
    ...mapState("eventCenter", [
      "loading",
      "eventTypes",
      "clientID",
      "infos",
      "eventMap",
      "events",
      "eventFuncs"
    ])
  },
  methods: {
    toggleSubscription(eventType) {
      if (this.eventFuncs.has(eventType)) {
        this.unsubscript(eventType);
      } else {
        this.subscript({
          type: eventType,
          cb: e => {
            console.log(e);
          }
        });
      }
    },
    formatJson(event) {
      try {
        const s = this.eventToSend.body;
        let start = -1;
        let startIdOfSameChar = 0;
        let offset = 1;
        let c = s[event.target.selectionStart - offset];

        while (this.$_.includes("\n\r\t ", c)) {
          offset++;
          c = s[event.target.selectionStart - offset];
        }
        while (start++ != event.target.selectionStart - offset) {
          start = s.indexOf(c, start);
          startIdOfSameChar++;
        }
        this.eventToSend.body = JSON.stringify(
          JSON.parse(this.eventToSend.body),
          null,
          "\t"
        );
        start = 0;
        for (let i = 0; i < startIdOfSameChar; i++) {
          start = this.eventToSend.body.indexOf(c, start) + 1;
        }
        event.target.selectionStart = start + offset - 1;
      } catch (e) {
        console.log(e);
      }
    },
    ...mapMutations("eventCenter", [
      "cleanEvents",
      "toggleLoading",
      "dismissError",
      "initClient",
      "setLoading"
    ]),
    ...mapActions("eventCenter", [
      "getEventList",
      "unsubscript",
      "subscript",
      "getToken",
      "login",
      "eventTunnel"
    ]),
    sendEvent() {
      try {
        if (
          !this.$store.state.eventCenter.ws ||
          this.$store.state.eventCenter.wsState == "closed"
        ) {
          throw new Error("not connect server");
        }
        const event = this.$_.merge(
          { Type: this.eventToSend.type },
          JSON.parse(this.eventToSend.body)
        );
        this.$store.dispatch("eventCenter/sendEvent", event);
      } catch (e) {
        this.$store.commit("eventCenter/addError", e);
      }
    },
    expandEventInfo(event) {
      console.log(event);
    },
    connect() {
      this.setLoading(true);
      if (this.$store.state.eventCenter.wsState == "closed") {
        this.initClient({
          host: this.host,
          port: this.port
        });
        this.login()
          .then(this.getEventList)
          .then(() => {
            return this.subscript({
              type: "main.EventTest",
              cb: event => {
                window.console.log(event);
                //   let e = this.$_.clone(event);
                //   e.CostumeField = "Helllllllo";
                //   this.$store.dispatch("eventCenter/sendEvent", e);
              }
            });
          })
          .then(() => {
            return this.getToken();
          })
          .then(() => {
            return this.eventTunnel();
          })
          .finally(() => {
            return this.setLoading(false);
          });
      } else {
        this.$store.dispatch("eventCenter/close");
      }
    }
  }
};
</script>
