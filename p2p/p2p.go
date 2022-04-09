package p2p

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/yyuurriiaa/ProjectMSSP/blockchain"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

// var conns []*websocket.Conn
var upgrader = websocket.Upgrader{} //initialize

//openPort값을 가지는 port를 ws로 upgrade 하고 해당 port의 값을 가지는 peer를 새로 만들고 Peers에 추가. 그리고 peer의 inbox에 들어오는 값을 go routine으로 write, read.
func Upgrade(rw http.ResponseWriter, r *http.Request) {
	//3000포트가 4000포트에서 온 request를 upgrade함
	openPort := r.URL.Query().Get("openPort")           //query로 url에서 openPort 가져옴
	ip := utils.Splitter(r.RemoteAddr, ":", 0)          //컴퓨터 주소의 ip 가져옴
	upgrader.CheckOrigin = func(r *http.Request) bool { //openPort와 ip 값이 존재하면 CheckOrigin을 true로 함
		return openPort != "" && ip != ""
	}
	conn, err := upgrader.Upgrade(rw, r, nil) //ws으로 업그레이드
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

//port : 새로 연결하려는 포트, openPort : 기존에 연결된 포트. gorilla websocket으로 websocket.Conn 을 생성하고 해당 Conn을 가지는 peer를 만듬.
//그 후 Peers에 만들어진 peer를 추가하고 만약 이 peer가 새로 연결된 peer(기존에 연결하고 끊었다가 다시 연결한게 아닌)일 경우 다른 Peers에게 새로운 peer를 전파함.
//기존에 연결되었던 peer 라면 peer에 가장 최근의 block을 보내어 통신.
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

//peers에 있는 모든 peer의 inbox 채널에 newBlock을 대입.
func BroadcastNewBlock(b *blockchain.Block) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewBlock(b, p)
	}
}

//연결된 모든 peer의 inbox 채널에 새로 검증한 tx를 대입.
func BroadcastNewTx(tx *blockchain.Tx) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewTx(tx, p)
	}
}

//newPeer를 제외한 다른 peer 들에게 newPeer의 key와 기존 peer의 port를 알려줌. openPort 를 알아야 하기 때문.
func broadcastNewPeer(newPeer *peer) {
	for key, p := range Peers.v {
		if key != newPeer.key {
			portInfo := fmt.Sprintf("%s:%s", newPeer.key, p.port) // newPeer의 주소와 기존의 openPort
			notifyNewPeer(portInfo, p)                            // 다른 peer 들에게 새로운 peer의 주소를 알려줌
		}
	}
}
