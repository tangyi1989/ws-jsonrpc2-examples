<template>
  <div id="app">
    <div v-show="!showRoom">
      <h1>ChatRoom via JSON RPC 2.0</h1>
      <form v-on:submit.prevent="beginChat">
        <span>Username</span>
        <input v-model="username" type="text" />
        <input type="submit" value="Begin Chat"/>
      </form>
    </div>
    <div v-show="showRoom">
      <h1>Chat Room</h1>
      <span>websocket status: {{wsStatus}}</span>
      <br>
      <span>Userinfo: {{userinfo.username}}[{{userinfo.id}}]</span>
      <br>
      <div>
        <span>RoomList</span>
        <span v-for="id in roomList">
          <a href="javascript:void(0)" @click="joinRoom(id)">Room:{{id}}</a>
        </span>
      </div>
      <div>
        <span>Users</span>
        <span v-for="user in roomUsers">{{user.username}}[{{user.id}}]</span>
      </div>
      <span>CurrentRoom</span></span>{{currentRoom}}</span>
      <form v-on:submit.prevent="speak">
        <input v-model="message" type="text" placeholder="Input text here..." style="width:400px;"></input>
        <input type="submit" value="Send" />
      </form>
      <div>
        <p v-for="msg in this.messages.slice().reverse()">{{msg.username}}[{{msg.id}}]:{{msg.message}}</p>
      </div>
    </div>
    
  </div>
</template>

<script>
var RPCSocket = require('rpc-websockets').Client

export default {
  name: 'app',
  data () {
    return {
      username: "",
      message: "",
      showRoom: false,
      rpcSocket: null,
      userinfo: {},
      roomUsers: [],
      roomList: [],
      currentRoom: null,
      messages:[],
      wsStatus: "initing..."
    }
  },
  methods: {
    beginChat: function() {
      if (this.username == "") {
        alert("Username is empty")
        return
      }
      this.showRoom = true
      console.log("username", this.username)

      var rs = new RPCSocket("ws://127.0.0.1:7000/chatsvr?username="+this.username)
      rs.on("open", () => {
        this.wsStatus = "connected"
        rs.call("ChatService.GetUserInfo", []).then((userinfo) => {
          this.userinfo = userinfo
        })
      })
      rs.on("close", () => {
        this.wsStatus = "discconected"

        this.userinfo= {}
        this.roomUsers = []
        this.roomList = []
        this.currentRoom = null
      })
      rs.on("error", (evt) => {
        this.wsStatus = "connect error"
      })
      rs.on("roomList", (roomList) => {
        //console.log("roomList", roomList)
        roomList.sort()
        this.roomList = roomList
      })
      rs.on("roomUsers", (roomUsers) => {
        //console.log("roomUsers", roomUsers)
        this.roomUsers = roomUsers
      })
      rs.on("userSpeak", (msg) => {
        msg.date = new Date()
        this.messages.push(msg)
      })
      this.rpcSocket = rs
    },
    joinRoom: function(id) {
      this.rpcSocket.call("ChatService.Join", [id]).then((result) => {
        this.currentRoom = id
      })
    },
    speak: function() {
      if(this.currentRoom == null) {
        alert("Join room first")
        return
      }
      if (this.message.length > 0) {
        this.rpcSocket.call("ChatService.Say", [this.message]).then(() => {
        })
        this.message = ""
      }
    }
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  margin-top: 60px;
}
span {
  padding:10px;
}
form {
  padding:10px;
}
p {
  padding:0px;
  margin:0px;
}
</style>
