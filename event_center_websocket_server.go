package eventcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync"

	"github.com/daniel840829/eventcore/proto"
	"github.com/go-kit/kit/log/level"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EventCenterWebsocketServer struct {
	UUID               uuid.UUID
	clients            *sync.Map // ClientID : *EventCenterClient
	clientSubscription *sync.Map // EventType : *EventCenterClientList
	tokens             *sync.Map // ClientID : connection token
	EventCenter        *EventCenterCluster
}

func GetClientID(r *http.Request) string {
	id, err := r.Cookie("client_id")
	if err != nil {
		level.Error(Logger).Log("error", err)
		return ""
	}
	return id.Value
}

func StartWebsocketServer(center *EventCenterCluster, port string) {

	s := &EventCenterWebsocketServer{
		EventCenter:        center,
		clients:            &sync.Map{},
		clientSubscription: &sync.Map{},
		tokens:             &sync.Map{},
	}
	center.server = s
	r := mux.NewRouter()
	r.HandleFunc("/{file:[^\\.]+\\.[^\\.]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		http.ServeFile(w, r, path.Join("./web/dist/", vars["file"]))
	}).Methods("GET")

	r.HandleFunc("/subscript", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		info := &proto.SubscriptInfo{}
		err := json.NewDecoder(r.Body).Decode(&info)
		if err != nil {
			level.Error(Logger).Log("error", err)
			return
		}
		if info.ClientID == "" {
			info.ClientID = uuid.New().String()
		}

		actual, _ := s.clientSubscription.LoadOrStore(EventType(info.EventType), NewEventCenterClientList())
		list := actual.(*EventCenterClientList)
		result := &proto.SubscriptResult{
			ClientID: info.ClientID,
		}
		if !list.AppendIfNonExist(info.ClientID) {
			result.Success = false
			result.Error = fmt.Sprintf("client id %s duplicated", info.ClientID)
		} else {

			token := uuid.New().String()
			actaulToken, _ := s.tokens.LoadOrStore(info.ClientID, token)
			result.Token = actaulToken.(string)

			Logger.Log("Get subscriber")
			s.EventCenter.Subscribe(EventType(info.EventType), func(e Event) error {
				level.Info(Logger).Log("to websocket client", e)
				return nil
			}, info.EventType)
			result.Success = true
		}
		data, err := json.Marshal(result)
		if err != nil {
			level.Error(Logger).Log("error", err)
			return
		}
		// http.SetCookie(w, &http.Cookie{
		// 	Name:  "client_id",
		// 	Value: info.ClientID,
		// })
		w.Write(data)
	}).Methods("POST")
	r.HandleFunc("/event_tunnel", func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		token := v.Get("token")
		clientID := ""
		s.tokens.Range(func(key, value interface{}) bool {
			if value.(string) == token {
				clientID = key.(string)
				s.tokens.Delete(clientID)
				return false
			}
			return true
		})

		if clientID == "" {
			err := errors.New("no client ID")
			level.Error(Logger).Log("error", err)
			return
		}

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
			level.Error(Logger).Log("error", err)
		}
		go func() {
			_, isLoaded := s.clients.LoadOrStore(clientID, &EventCenterClient{ID: clientID, conn: conn})
			if isLoaded {
				err = fmt.Errorf("not support multi tunnel for same client id: %s", clientID)
				level.Error(Logger).Log("error", err)
				return
			}
			defer func() {
				s.clients.Delete(clientID)
			}()
			for {
				payload, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					level.Error(Logger).Log("error", err, "client id", clientID, "message", "received error")
					break
				}
				var e Event = &EventBase{}
				e, err = e.Unserialize(payload)
				if err != nil {
					level.Error(Logger).Log("error", err, "client id", clientID, "message", "received error")
					break
				}
				s.EventCenter.Emit(e)
			}
		}()
	})
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
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
