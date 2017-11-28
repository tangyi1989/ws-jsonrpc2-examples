package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tangyi1989/ws-jsonrpc2"
)

type Empty struct{}

type ChatService struct {
	mu sync.Mutex

	nextId int
	users  map[int]*User
	rooms  map[int]*ChatRoom
}

type UserInfo struct {
	Username string `json:"username"`
	Id       int    `json:"id"`
}

type ChatRoom struct {
	sync.Mutex

	users map[int]*User
}

type User struct {
	sync.Mutex

	id       int
	roomId   int
	username string
	conn     *jsonrpc2.Conn
}

func NewChatService() *ChatService {
	rooms := make(map[int]*ChatRoom)
	for i := 0; i < 10; i++ {
		rooms[i] = &ChatRoom{
			users: make(map[int]*User),
		}
	}

	return &ChatService{
		users: make(map[int]*User),
		rooms: rooms,
	}
}

func (srv *ChatService) addUser(username string, conn *jsonrpc2.Conn) *User {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	id := srv.nextId
	srv.nextId++

	user := &User{
		id:       id,
		roomId:   -1,
		conn:     conn,
		username: username,
	}

	log.Println("id:", user.id, " name:", user.username, ", comming")

	conn.SetData("user", user)
	srv.users[id] = user

	return user
}

func (srv *ChatService) removeUser(user *User) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	log.Println("id ", user.id, " name", user.username, ", leave")
	delete(srv.users, user.id)
	if user.roomId >= 0 {
		srv.notifyUserInfos(user.roomId)
	}
}

func (srv *ChatService) getRoomList() []int {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var roomIds []int
	for id, _ := range srv.rooms {
		roomIds = append(roomIds, id)
	}

	return roomIds
}

func (srv *ChatService) notifyUserInfos(roomId int) {
	var roomUsers []*UserInfo
	for _, user := range srv.users {
		if user.roomId == roomId {
			roomUsers = append(roomUsers, &UserInfo{
				Username: user.username,
				Id:       user.id,
			})
		}
	}

	for _, user := range srv.users {
		if user.roomId == roomId {
			go user.conn.Notify("roomUsers", roomUsers)
		}
	}
}

func (srv *ChatService) Join(conn *jsonrpc2.Conn, roomId int, empty *Empty) error {
	user := conn.GetData("user").(*User)
	log.Println("user:", user.username, " join:", roomId)

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.rooms[roomId]; !ok {
		return errors.New("Invalid Room")
	}

	prevRoomId := user.roomId
	user.roomId = roomId
	if prevRoomId >= 0 {
		srv.notifyUserInfos(prevRoomId)
	}
	srv.notifyUserInfos(roomId)

	return nil
}

func (srv *ChatService) Say(conn *jsonrpc2.Conn, message string, empty *Empty) error {
	user := conn.GetData("user").(*User)
	log.Println("user:", user.username, " say:", message)

	srv.mu.Lock()
	defer srv.mu.Unlock()

	type UserSpeak struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
		Message  string `json:"message"`
	}

	if user.roomId < 0 {
		return errors.New("Invlid room")
	}

	for _, t := range srv.users {
		if t.roomId == user.roomId {
			go t.conn.Notify("userSpeak", &UserSpeak{
				Id:       user.id,
				Username: user.username,
				Message:  message,
			})
		}
	}

	return nil
}

func (srv *ChatService) GetUserInfo(conn *jsonrpc2.Conn, empty *Empty, userInfo *UserInfo) error {
	user := conn.GetData("user").(*User)
	*userInfo = UserInfo{
		Username: user.username,
		Id:       user.id,
	}

	return nil
}

func serveRPC(server ...*jsonrpc2.Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		jsonrpc2.ServeRPC(r, ws, server...)
	}
}

func main() {
	rpcSvr := jsonrpc2.NewServer()
	chatSrv := NewChatService()

	rpcSvr.OnConnInit(func(conn *jsonrpc2.Conn) {
		args := conn.Request.URL.Query()
		usernames := args["username"]
		if len(usernames) == 0 {
			log.Println("Http request without username!")

			conn.Notify("loginError", "Username not provided.")
			conn.Close()
		}

		user := chatSrv.addUser(usernames[0], conn)
		conn.OnClose(func() {
			chatSrv.removeUser(user)
		})

		conn.Notify("roomList", chatSrv.getRoomList())
	})

	rpcSvr.Register(chatSrv)

	http.HandleFunc("/chatsvr", serveRPC(rpcSvr))
	log.Println("err:", http.ListenAndServe(":7000", nil))
}
