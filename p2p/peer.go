package p2p

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type peers struct {
	v map[string]*peer
	m sync.Mutex
}

var Peers = peers{
	v: make(map[string]*peer),
}

type peer struct {
	conn    *websocket.Conn
	inbox   chan []byte
	key     string // 연결 주소. address + port
	address string
	port    string
}

//peers의 key(localhost:4000같은) 값을 keys []string에 저장하고 keys 리턴. 즉 모든 peer의 address 를 []string 형태로 반환.
func AllPeers(p *peers) []string {
	p.m.Lock()
	defer p.m.Unlock()

	var keys []string
	for key := range p.v {
		keys = append(keys, key)
	}
	return keys
}

//peer의 연결을 끊을 때 websocket을 Close하고 Peers 에서 peer 삭제.
func (p *peer) close() {
	Peers.m.Lock()
	defer Peers.m.Unlock()

	p.conn.Close()
	delete(Peers.v, p.key)
}

//p.conn에서 (json으로 된)message를 받으면 handleMsg 실행.
func (p *peer) read() {
	defer p.close()
	for {
		m := Message{}
		err := p.conn.ReadJSON(&m)
		if err != nil { // err가 nil 이면 m에 값이 들어왔다는 뜻. 즉, m에 값이 안들어왔으면(메세지를 못받았으면) break
			break
		}
		handleMsg(&m, p) //값을 받으면 실행됨
	}
}

//p.inbox에 message를 받으면 p.conn에 message를 씀.
func (p *peer) write() {
	defer p.close()
	for {
		m, ok := <-p.inbox
		if !ok {
			break //값을 못받으면 break
		}
		p.conn.WriteMessage(websocket.TextMessage, m) //값을 받으면 실행됨
	}
}

//address 와 port를 받아서 key(localhost:4000같은)를 만들고 새로운 peer에 대입 후, Peers에 새로 만든 peer을 추가.
// 그 후 go routine으로 새로 만들어진 peer에 들어오는 inbox값을 읽고 쓰기.
func initPeer(conn *websocket.Conn, address string, port string) *peer { // 새로 peer 만들고 message read, write
	Peers.m.Lock()
	defer Peers.m.Unlock()
	key := fmt.Sprintf("%s:%s", address, port)
	p := &peer{
		conn:    conn,
		inbox:   make(chan []byte),
		key:     key,
		address: address,
		port:    port,
	}
	Peers.v[key] = p
	go p.read() //go routine. 계속 실행되고있다고 봐야하나?
	go p.write()
	return p
}
