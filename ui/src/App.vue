<template>
  <v-app>
    <v-app-bar app color="primary" dark>
      <div class="d-flex align-center">
        Event Core
      </div>
      <v-spacer></v-spacer>
      <v-text-field label="host" v-model="host"></v-text-field>
      <v-spacer></v-spacer>
      <v-text-field label="port" v-model="port"></v-text-field>
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
        href="https://github.com/vuetifyjs/vuetify/releases/latest"
        target="_blank"
        text
      >
        <span class="mr-2">Github</span>
        <v-icon>mdi-open-in-new</v-icon>
      </v-btn>
    </v-app-bar>
    <v-main>
      <v-alert
        :type="info.type"
        prominent
        v-if="!$_.isEmpty(info)"
        border="left"
      >
        <v-row>
          <v-col class="grow" v-text="info.message"></v-col>
          <v-col class="shrink">
            <v-btn @click="dismissError">Dismiss</v-btn>
          </v-col>
        </v-row>
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
          <v-form ref="form" v-model="valid" :lazy-validation="true">
            <v-overflow-btn
              class="my-2"
              :items="eventTypes"
              label="Event Type"
              v-model="eventToSend.type"
            ></v-overflow-btn>
            <v-textarea
              label="Event Body"
              v-model="eventToSend.body"
            ></v-textarea>
            <v-btn color="green" class="mr-4" @click="sendEvent">
              Send Event
            </v-btn>
          </v-form>
        </v-col>
      </v-row>
    </v-main>
  </v-app>
</template>

<script>
import { mapGetters, mapMutations, mapState } from "vuex";

export default {
  name: "App",
  data() {
    return {
      wsLoading: false,
      host: "localhost",
      port: "7000",
      eventInfoExpanded: new Map(),
      eventToSend: {
        body: "",
        type: "main.EventTest"
      }
    };
  },
  computed: {
    info() {
      return this.$store.state.eventCenter.infos[
        this.$store.state.eventCenter.infos.length - 1
      ];
    },
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
    ...mapGetters("eventCenter", {
      events: "allEvents"
    }),
    ...mapState("eventCenter", ["loading", "eventTypes", "clientID"])
  },
  methods: {
    ...mapMutations("eventCenter", ["cleanEvents", "toggleLoading"]),
    sendEvent() {
      this.eventToSend;
      let event = null;
      try {
        this.$_.merge(
          { clientID: this.clientID, type: this.eventToSend.type },
          JSON.parse(this.eventToSend.body)
        );
        this.$store.dispatch("eventCenter/sendEvent", event);
      } catch (e) {
        console.log(e);
      }
    },
    expandEventInfo(event) {
      console.log(event);
    },
    dismissError() {
      this.$store.commit("eventCenter/dismissError");
    },
    connect() {
      this.toggleLoading();
      if (this.$store.state.eventCenter.wsState == "closed") {
        this.$store.commit("eventCenter/initClient", {
          host: this.host,
          port: this.port
        });
        this.$store
          .dispatch("eventCenter/subscript", {
            type: "main.EventTest",
            cb: event => {
              window.console.log(event);
              //   let e = this.$_.clone(event);
              //   e.CostumeField = "Helllllllo";
              //   this.$store.dispatch("eventCenter/sendEvent", e);
            }
          })
          .then(() => this.$store.dispatch("eventCenter/eventTunnel"));
      } else {
        this.$store.dispatch("eventCenter/close");
      }
    }
  }
};
</script>
