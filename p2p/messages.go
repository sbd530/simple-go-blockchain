package p2p

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yyuurriiaa/ProjectMSSP/blockchain"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

type MessageKind int

const (
	MessageNewestBlock MessageKind = iota
	MessageAllBlocksRequest
	MessageAllBlocksResponse
	MessageNewBlockNotify
	MessageNewTxNotify
	MessageNewPeerNotify
)

type Message struct {
	Kind    MessageKind
	Payload []byte
}

func makeMessage(kind MessageKind, payload interface{}) []byte {
	m := Message{
		Kind:    kind,
		Payload: utils.ToJSON(payload),
	}
	// fmt.Println("\nmakeMessage m:", m)
	return utils.ToJSON(m)
}

func sendNewestBlock(p *peer) { //연결한 측에서 사용됨
	fmt.Printf("\nsending newest block to %s\n", p.key)
	b, err := blockchain.FindBlock(blockchain.Blockchain().NewestHash)
	utils.HandleErr(err)
	fmt.Println("b :", b)
	m := makeMessage(MessageNewestBlock, b)
	p.inbox <- m

}

func requestAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksRequest, nil)
	p.inbox <- m
}

func sendAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksResponse, blockchain.Blocks(blockchain.Blockchain()))
	p.inbox <- m
}

func notifyNewBlock(b *blockchain.Block, p *peer) {
	m := makeMessage(MessageNewBlockNotify, b)
	p.inbox <- m
}

func notifyNewTx(tx *blockchain.Tx, p *peer) {
	m := makeMessage(MessageNewTxNotify, tx)
	p.inbox <- m
}

func notifyNewPeer(address string, p *peer) {
	m := makeMessage(MessageNewPeerNotify, address)
	p.inbox <- m
}

func handleMsg(m *Message, p *peer) { //연결된 측에서 사용됨
	switch m.Kind {
	case MessageNewestBlock: //새 블록을 보냄
		fmt.Printf("\nreceived the newest block from %s\n", p.key)
		// fmt.Println(m.Kind, m.Payload)
		var msgBlock blockchain.Block               //var payload blockchain.Block으로 빈 블록을 만든 후,
		err := json.Unmarshal(m.Payload, &msgBlock) // json.Unmarshal로 m.payload(다른 포트에서 받아온)의 내용물을 unmarshal 하여 payload 블록에 저장
		utils.HandleErr(err)
		fmt.Println("\nmsgBlock : ", msgBlock)
		b, err := blockchain.FindBlock(blockchain.Blockchain().NewestHash)
		utils.HandleErr(err)
		if msgBlock.Height >= b.Height { //다른 포트의 height가 이 포트의 height보다 크면
			//다른 포트에게 모든 블록을 요청
			fmt.Printf("\nrequest all blocks from %s\n", p.key)
			requestAllBlocks(p)
		} else {
			//이 포트의 newest block를 다른 포트로 보냄
			fmt.Printf("\nsend newest block to %s\n", p.key)
			sendNewestBlock(p)
		}
	case MessageAllBlocksRequest: //모든 블록을 다른 포트에게 요청
		fmt.Printf("\n%s wants all blocks\n", p.key)
		sendAllBlocks(p)
	case MessageAllBlocksResponse: //모든 블록을 다른 포트에게 받음
		fmt.Printf("\nreceived all blocks from %s\n", p.key)
		var msgAllBlocks []*blockchain.Block //양이 많아서 포인터를 사용하나?
		err := json.Unmarshal(m.Payload, &msgAllBlocks)
		utils.HandleErr(err)
		fmt.Println("\nmsgAllBlocks : ", msgAllBlocks)
		blockchain.Blockchain().Replace(msgAllBlocks) // replace blockchain and blocks
	case MessageNewBlockNotify:
		var msgNewBlock *blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &msgNewBlock))
		blockchain.Blockchain().AddPeerBlock(msgNewBlock)
	case MessageNewTxNotify:
		var msgNewTx *blockchain.Tx
		utils.HandleErr(json.Unmarshal(m.Payload, &msgNewTx))
		blockchain.Mempool().AddPeerTx(msgNewTx)
	case MessageNewPeerNotify:
		var msgNewPeer string // newPeer의 address이므로 string
		utils.HandleErr(json.Unmarshal(m.Payload, &msgNewPeer))
		fmt.Printf("now /ws upgrade %s", msgNewPeer)
		parts := strings.Split(msgNewPeer, ":")      // address, port, openPort로 조각냄
		AddPeer(parts[0], parts[1], parts[2], false) // broadcastNewPeer에서 이미 새로운 peer 확인을 햇으므로 false
	}
}
