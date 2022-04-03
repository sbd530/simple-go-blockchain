package p2p

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/yyuurriiaa/ProjectMSSP/blockchain"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

var upgrader = websocket.Upgrader{}

// var conns []*websocket.Conn

func Upgrade(rw http.ResponseWriter, r *http.Request) {
	//3000포트가 4000포트에서 온 request를 upgrade함
	openPort := r.URL.Query().Get("openPort")
	ip := utils.Splitter(r.RemoteAddr, ":", 0)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return openPort != "" && ip != ""
	}
	conn, err := upgrader.Upgrade(rw, r, nil)
	fmt.Printf("port %s upgrade!\n", openPort)

	// conns = append(conns, conn)
	utils.HandleErr(err)
	// for {
	// 	_, p, err := conn.ReadMessage()
	// 	if err != nil {
	// 		conn.Close()
	// 		break
	// 	}
	// 	for _, aConn := range conns {
	// 		if aConn != conn {
	// 			utils.HandleErr(aConn.WriteMessage(websocket.TextMessage, p))
	// 		}
	// 	}

	// }
	initPeer(conn, ip, openPort)
	fmt.Println("\nupgrade complete")

}

func AddPeer(address string, port string, openPort string, broadcast bool) { // broadcast bool : 새로운 연결인지 확인하기 위함
	//4000포트에서 3000포트로 upgrade를 request함
	fmt.Printf("\nport %s -> port %s\n", openPort, port)
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws?openPort=%s", address, port, openPort), nil) // dial의 URL을 call하면 새로운 connection을 만듬
	utils.HandleErr(err)
	fmt.Println("\naddpeer start")
	p := initPeer(conn, address, port)
	if broadcast {
		broadcastNewPeer(p)
		return //새 연결일 경우 sendNewestBlock 하지 않음
	}
	sendNewestBlock(p)

}

func BroadcastNewBlock(b *blockchain.Block) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewBlock(b, p)
	}
}

func BroadcastNewTx(tx *blockchain.Tx) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewTx(tx, p)
	}
}

func broadcastNewPeer(newPeer *peer) {
	for key, p := range Peers.v {
		if key != newPeer.key {
			portInfo := fmt.Sprintf("%s:%s", newPeer.key, p.port) // newPeer의 주소와 기존의 openPort
			notifyNewPeer(portInfo, p)                            // 다른 peer 들에게 새로운 peer의 주소를 알려줌
		}
	}
}
