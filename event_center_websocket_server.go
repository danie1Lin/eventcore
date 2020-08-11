package eventcore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"runtime/debug"

	"github.com/daniel840829/eventcore/proto"
	"github.com/go-kit/kit/log/level"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type EventCenterWebsocketServer struct {
	UUID               uuid.UUID
	clients            *sync.Map // ClientID : *EventCenterClient
	clientSubscription *sync.Map // EventType : *EventCenterClientList
	tokens             *sync.Map // ClientID : connection token
	EventCenter        *EventCenterCluster
}

func StartWebsocketServer(center *EventCenterCluster, port string) {
	s := &EventCenterWebsocketServer{
		EventCenter:        center,
		clients:            &sync.Map{},
		clientSubscription: &sync.Map{},
		tokens:             &sync.Map{},
	}
	center.server = s
	s.Init(port)
}

func (s *EventCenterWebsocketServer) Init(port string) {
	route := mux.NewRouter()

	Handle(route, "/login", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"clientID": uuid.New().String()})
	}).Methods("POST")

	Handle(route, "/events", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GetEventInfoList())
	}).Methods("Get")

	Handle(route, "/subscript", func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientID(r)
		if clientID == "" {
			panic("no client id")
		}
		eventList := []string{}
		s.clientSubscription.Range(func(key, value interface{}) bool {
			list := value.(*EventCenterClientList)
			if list.Find(clientID) == -1 {
				return true
			}
			eventList = append(eventList, string(key.(EventType)))
			return true
		})
		if err := json.NewEncoder(w).Encode(eventList); err != nil {
			panic(err)
		}
	}, OPTION_AUTH).Methods("GET")
	Handle(route, "/subscript/cancel", func(w http.ResponseWriter, r *http.Request) {
		info := &proto.SubscriptInfo{}
		err := json.NewDecoder(r.Body).Decode(&info)
		if err != nil {
			panic(err)
		}
		ClientID := GetClientID(r)
		if ClientID == "" {
			panic("no client id")
		}
		if v, ok := s.clientSubscription.Load(EventType(info.EventType)); ok {
			list := v.(*EventCenterClientList)
			list.Remove(ClientID)
		} else {
			level.Warn(Logger).Log("error", "event not subscript", "event_type", info.EventType)
		}
		result := &proto.SubscriptResult{
			Success: true,
		}
		if err := json.NewEncoder(w).Encode(result); err != nil {
			panic(err)
		}
		return
	}, OPTION_AUTH).Methods("POST")
	Handle(route, "/eventTunnel/token", func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientID(r)
		if clientID == "" {
			panic("no client id")
		}
		if err := json.NewEncoder(w).Encode(map[string]string{"token": s.GetToken(clientID)}); err != nil {
			panic(err)
		}
	}, OPTION_AUTH).Methods("GET")
	Handle(route, "/subscript", func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientID(r)
		if clientID == "" {
			panic("no client id")
		}
		info := &proto.SubscriptInfo{}
		err := json.NewDecoder(r.Body).Decode(&info)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		actual, _ := s.clientSubscription.LoadOrStore(EventType(info.EventType), NewEventCenterClientList())
		list := actual.(*EventCenterClientList)
		result := &proto.SubscriptResult{
			Success: true,
		}
		if !list.AppendIfNonExist(clientID) {
			result.Success = false
			result.Error = fmt.Sprintf("client id %s duplicated", clientID)
		} else {
			s.EventCenter.Subscribe(EventType(info.EventType), func(e Event) error { return nil }, fmt.Sprintf("websocketClient-%s", clientID))
		}

		data, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		w.Write(data)
	}, OPTION_AUTH).Methods("POST")
	Handle(route, "/event_tunnel", func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		token := v.Get("token")
		clientID := s.ConsumeToken(token)
		if clientID == "" {
			panic("no client ID")
		}
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			panic(err)
		}
		_, isLoaded := s.clients.LoadOrStore(clientID, &EventCenterClient{ID: clientID, conn: conn})
		if isLoaded {
			err = fmt.Errorf("not support multi tunnel for same client id: %s", clientID)
			panic(err)
		}
		defer func() {
			s.removeClient(clientID)
			if err := recover(); err != nil {
				panic(err)
			}
		}()
		for {
			payload, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				level.Error(Logger).Log("error", err, "client id", clientID, "message", "received error")
				panic(err)
			}
			var e Event = &EventBase{}
			e, err = e.Unserialize(payload)
			if err != nil {
				level.Error(Logger).Log("error", err, "client id", clientID, "message", "received error")
				panic(err)
			}
			e.AddNode(DispatcherInfo{
				ID:   clientID,
				Type: "WebsocketClient",
				Name: "WebsocketClient",
			})
			s.EventCenter.Emit(e)
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%s", port), handlers.CORS(
		handlers.AllowedHeaders([]string{"Authorization", "X-Client-ID", "Content-Type"}),
		handlers.AllowedOrigins([]string{"http://localhost:8080"}),
		handlers.AllowedMethods([]string{"POST", "GET", "PUT"}),
		handlers.AllowCredentials(),
	)(route))
}

func (s *EventCenterWebsocketServer) removeClient(clientID string) {
	s.clients.Delete(clientID)
	s.clientSubscription.Range(func(k, v interface{}) bool {
		clientsList := v.(*EventCenterClientList)
		clientsList.Remove(clientID)
		return true
	})
}

func (s *EventCenterWebsocketServer) GetDispatcherInfo() DispatcherInfo {
	return DispatcherInfo{
		Name: "websocket server",
		ID:   s.UUID.String(),
		Type: "websocket server",
	}
}

func (s *EventCenterWebsocketServer) Emit(event Event) {
	event.AddNode(s.GetDispatcherInfo())
	eventData, err := event.Serialize()
	if err != nil {
		level.Error(Logger).Log("error", err)
		return
	}

	if v, ok := s.clientSubscription.Load(event.GetType()); ok {
		sourceID := event.GetSourceID()
		clients := v.(*EventCenterClientList)
		clients.Range(func(clientID string) bool {
			if clientID == sourceID {
				return true
			}
			v, ok := s.clients.Load(clientID)
			if !ok {
				level.Error(Logger).Log("error", fmt.Errorf("client not exist"), "client_id", clientID)
				return true
			}
			client := v.(*EventCenterClient)
			if err := wsutil.WriteServerMessage(client.conn, ws.OpText, eventData); err != nil {
				level.Error(Logger).Log("error", err)
			}
			return true
		})
	}
}

func (s *EventCenterWebsocketServer) GetToken(clientID string) string {
	token := uuid.New().String()
	actaulToken, _ := s.tokens.LoadOrStore(clientID, token)
	return actaulToken.(string)
}

func (s *EventCenterWebsocketServer) ConsumeToken(token string) string {
	clientID := ""
	s.tokens.Range(func(key, value interface{}) bool {
		if value.(string) == token {
			clientID = key.(string)
			s.tokens.Delete(clientID)
			return false
		}
		return true
	})
	return clientID
}

func GetClientID(r *http.Request) string {
	return r.Context().Value("ClientID").(string)
}

func Authorization(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if clientID := r.Header.Get("X-Client-ID"); clientID == "" {
			w.WriteHeader(401)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "no client id provided"}); err != nil {
				level.Error(Logger).Log("error", err)
			}
			return
		} else {
			handler(w, r.WithContext(context.WithValue(r.Context(), "ClientID", clientID)))
		}
	}
}

func DontPanic(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if v, ok := err.(error); ok {
					v = errors.WithStack(v)
					level.Error(Logger).Log("error", err, "trace", fmt.Sprintf("%+v", err))
				} else {
					level.Error(Logger).Log("error", err, "trace", string(debug.Stack()))
				}

				w.WriteHeader(500)
				errorMsg := ""
				switch v := err.(type) {
				case error:
					errorMsg = v.Error()
				case string:
					errorMsg = v
				case fmt.Stringer:
					errorMsg = v.String()
				default:
					errorMsg = fmt.Sprintf("%+v", err)
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"error": errorMsg})
			}
		}()
		handler(w, r)
	}
}

type HandlerOption string

const (
	OPTION_AUTH HandlerOption = "auth"
)

func Handle(router *mux.Router, path string, handler http.HandlerFunc, opts ...HandlerOption) *mux.Route {
	for _, opt := range opts {
		switch opt {
		case OPTION_AUTH:
			handler = Authorization(handler)
		}
	}
	handler = DontPanic(handler)
	return router.HandleFunc(path, handler)
}
