package blockchain

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/yyuurriiaa/ProjectMSSP/utils"
	"github.com/yyuurriiaa/ProjectMSSP/wallet"
)

const (
	minerReward int = 50 //블록 채굴 시 보상
)

type Tx struct {
	Id        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOuts"`
}

type TxIn struct {
	TxID      string `json:"txID"`  // 어떤 TxOut으로부터 만들어졌는지 알려줌. 즉, TxOut이 속해있는 Tx의 Id
	Index     int    `json:"index"` // 그 TxOut의 위치를 알려줌
	Signature string `json:"signature"`
	// Amount int    `json:"Amount"`
}

type TxOut struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type UTxOut struct {
	TxID   string `json:"txID"`
	Index  int    `json:"index"`
	Amount int    `json:"amount"`
}

type mempool struct { // mempool define
	// Txs []*Tx
	Txs map[string]*Tx //[] 형태는 삭제하기 번거롭고 속도도 느리므로 map 형태로 바꾸어서 delete사용가능하게. string은 tx id 사용
	m   sync.Mutex
}

// var mempool *mempool = &mempool{} //Mempool initialize.

//singleton
var m *mempool
var memOnce sync.Once

func Mempool() *mempool {
	memOnce.Do(func() {
		m = &mempool{
			Txs: make(map[string]*Tx), //map으로 만드므로 초기화
		}
	})
	return m
}

func MempoolMutex(m *mempool, rw http.ResponseWriter) {
	m.m.Lock()
	defer m.m.Unlock()
	utils.HandleErr(json.NewEncoder(rw).Encode(Mempool().Txs)) //mempool의 txs를 json으로 인코딩해서 rw에 저장
}

func (t *Tx) getId() { // tx 해싱해서 id 얻기
	// utils.GetHash(t)
	t.Id = utils.GetHash(t)
}

func (t *Tx) sign() {
	for _, txIn := range t.TxIns {
		txIn.Signature = wallet.Sign(t.Id, wallet.Wallet())
	}
}

func validate(tx *Tx) bool {
	valid := true
	for _, txIn := range tx.TxIns {
		prevTx := FindTx(Blockchain(), txIn.TxID)
		if prevTx == nil {
			valid = false
			break
		}
		address := prevTx.TxOuts[txIn.Index].Address
		valid = wallet.Verify(txIn.Signature, tx.Id, address)
		if !valid {
			break
		}
	}
	return valid
}

func makeCoinbaseTx(address string) *Tx { // coinbase에서 address에 보상 tx 만들기
	txIns := []*TxIn{
		{"", -1, "COINBASE"},
	}

	txOuts := []*TxOut{
		{address, minerReward},
	}
	tx := Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	// fmt.Printf("Id : %s\n", tx.Id)
	return &tx
}

var ErrorNotFund error = errors.New("not enough funds")
var ErrorNotValid error = errors.New("not valid tx")

func makeTx(from string, to string, amount int) (*Tx, error) { // mempool에 들어갈 tx를 생성
	if BalanceByAddress(from, Blockchain()) < amount {
		return nil, ErrorNotFund

	}
	var txOuts []*TxOut
	var txIns []*TxIn
	total := 0 // 보낼 수 있는 코인의 합
	uTxOuts := UTxOutsByAddress(from, Blockchain())

	for _, uTxOut := range uTxOuts {
		if total >= amount { // 보낼 수 있는 코인의 합이 amount보다 크거나 같아야 보낼 수 있음. 이걸 만족하면 더이상 total에 더하지 않아도 됨
			break
		}
		txIn := &TxIn{
			TxID:      uTxOut.TxID,
			Index:     uTxOut.Index,
			Signature: from,
		}
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	// fmt.Println("total:", total)

	if change := total - amount; change > 0 { //잔돈이 남앗을 때 다시 Txout을 만들고 추가해야함
		// fmt.Println("change added")
		changeTxOut := &TxOut{
			Address: from,
			Amount:  change,
		}
		txOuts = append(txOuts, changeTxOut)

	}

	txOut := &TxOut{ // 돈 받는 사람
		Address: to,
		Amount:  amount,
	}

	txOuts = append(txOuts, txOut)

	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId() // Id 추가
	tx.sign()
	valid := validate(tx)
	if !valid {
		return nil, ErrorNotValid
	}

	return tx, nil

	// if Blockchain().BalanceByAddress(from) <= amount { // from 이 amount이상의 돈을 가지고 있나 확인
	// 	return nil, errors.New("not enough money")

	// }

	// var txIns []*TxIn   // 보내고자 하는 amount만큼의 tx 생성
	// var txOuts []*TxOut // 보내고 남은 amount만큼의 tx 생성
	// oldTxOuts := Blockchain().TxOutsByAddress(from)
	// total := 0
	// for _, txOut := range oldTxOuts {
	// 	if total >= amount {
	// 		break
	// 	}
	// 	txIn := &TxIn{txOut.Owner, txOut.Amount} // txout ownerㅗ부터 txout amount 만큼 coinbase??로 이동
	// 	txIns = append(txIns, txIn)
	// 	total += txOut.Amount
	// }
	// change := total - amount //보내고 남은 거스름돈 반환
	// if change != 0 {
	// 	changeTxOut := &TxOut{from, change} // coinbase?? 에서 from에게 change만큼의 잔돈 반환
	// 	txOuts = append(txOuts, changeTxOut)

	// }
	// txOut := &TxOut{to, amount} //coinbase?? 에서 to에게 amount만큼 보내는 tx
	// txOuts = append(txOuts, txOut)
	// tx := &Tx{ //tx 생성
	// 	Id:        "",
	// 	Timestamp: int(time.Now().Unix()),
	// 	TxIns:     txIns,
	// 	TxOuts:    txOuts,
	// }
	// return tx, nil

}

func (m *mempool) AddTx(to string, amount int) (*Tx, error) { // mempool에 transaction 추가.
	tx, err := makeTx(wallet.Wallet().Address, to, amount)
	//utils.HandleErr(err) 이거로 하면 return값이 error가 아니고 log.panic이기때문에 안댐
	if err != nil {
		return nil, err
	}

	// m.Txs = append(m.Txs, tx)
	m.Txs[tx.Id] = tx
	return tx, nil
}

func (m *mempool) AddPeerTx(tx *Tx) {
	m.m.Lock()
	defer m.m.Unlock()
	// m.Txs = append(m.Txs, tx)
	m.Txs[tx.Id] = tx
}

//mempool의 tx를 승인하고 mempool을 비우는 역할
func (m *mempool) TxToConfirm() []*Tx {
	coinbase := makeCoinbaseTx(wallet.Wallet().Address) //coinbase에서 채굴자에게 주는 보상 tx
	// txs := m.Txs                // 처음에 coinbase 에서 보낸 tx는 들어가있지 않으므로
	var txs []*Tx

	// txs = append(txs, coinbase) // 여기서 추가해줌. 순서는 바뀌는데 상관없는듯?
	for _, tx := range m.Txs {
		txs = append(txs, tx)
	}

	txs = append(txs, coinbase)

	// m.Txs = nil // mempool 비우기
	m.Txs = make(map[string]*Tx) //map과 [] 형식과 비우는 방법 차이

	return txs
}

func isOnMempool(uTxOut *UTxOut) bool {
	exists := false
Outer:
	for _, tx := range Mempool().Txs {
		for _, input := range tx.TxIns {
			if input.TxID == uTxOut.TxID && input.Index == uTxOut.Index {
				exists = true // true를  찾아내도 for루프가 끝나진 않음 -> 속도가느려짐
				// break         //하나의 for 루프 나옴
				break Outer // label 붙여서 바깥 루프를 나옴

			}
		}
	}
	return exists
}
