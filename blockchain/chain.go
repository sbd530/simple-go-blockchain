package blockchain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/yyuurriiaa/ProjectMSSP/db"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

// type blockchain struct {
// 	blocks []*Block // block들이 길어질 수 있기 때문에 포인터로 짧게 받음
// }
//db로 관리하기때문에 비활성화

const (
	defaultDifficulty  int = 2 // 기본 난이도
	difficultyInterval int = 5 // 몇개의 블록마다 난이도를 확인하는가
	blockInterval      int = 2 // 약 2분마다 새로운 블록 생성
	timeRange          int = 2 // 약 2분의 여유
)

type blockchain struct {
	NewestHash     string `json:"newestHash"`
	Height         int    `json:"height"`
	CurrDifficulty int    `json:"currDifficulty"` //현재의 difficulty point
	m              sync.Mutex
}

//블록의 난이도 계산. 처음에는 디폴트 값으로 난이도 설정. block의 interval마다 난이도를 다시 계산하고 그 외의 경우에는 blockchain의 난이도를 계승.
func difficulty(b *blockchain) int {
	if b.Height == 0 { // genesis block의 난이도 = defaultDifficulty로 설정
		return defaultDifficulty
	} else if b.Height%difficultyInterval == 0 {
		//recalculate difficulty
		return recalculateDifficulty(b)
	} else {
		return b.CurrDifficulty
	}
}

//블록 생성에 걸리는 시간(interval)을 계산해서 채굴에 걸리는 시간에 따라 난이도 조절
func recalculateDifficulty(b *blockchain) int { //difficulty 다시 계산해서 currDifficulty에 넣어주기
	allBlocks := Blocks(b)
	newestBlock := allBlocks[0]
	lastCalculatedBlock := allBlocks[difficultyInterval-1]
	timeInterval := (newestBlock.Timestamp - lastCalculatedBlock.Timestamp) / 60 // newestBlock 과 lastCalculatedBlock 사이의 시간 간격. 실제 걸린 시간
	expectedTime := difficultyInterval * blockInterval                           // 예상 난이도 계산 시간
	if timeInterval > (expectedTime + timeRange) {                               // 걸린 시간에 따른 난이도 설정
		return b.CurrDifficulty - 1
	} else if (expectedTime - timeRange) < timeInterval {
		return b.CurrDifficulty
	} else {
		return b.CurrDifficulty + 1
	}
}

var b *blockchain  //singleton
var once sync.Once // 병렬처리해도 한번만 작동될 수 있도록

//singleton과 once를 이용한 blockchain생성(genesis). 이미 존재할 시 디코딩을 통해 체크포인트부터 연결
func Blockchain() *blockchain {
	once.Do(func() { //Do 안의 func가 한번 더 Do를 콜하면 데드록 발생 -> Do는 func 가 끝나기 전까지 종료되지 않기 때문
		b = &blockchain{
			Height: 0,
		} // blockchain초기화. 텅 빈 blockchain
		// fmt.Printf("newesthash: %s\n height: %d\n", b.NewestHash, b.Height)
		checkpoint := db.Checkpoint()
		if checkpoint == nil { //db에 checkpoint의 key값으로 저장된  value가 없으면
			b.AddBlock() //Genesis block 생성
			fmt.Println("Genesis Block created")
		} else { //checkpoint로 저장된 값이 있으면
			// fmt.Println("now decoding...")
			b.fromBytes(checkpoint) //checkpoint에서 decoding해서 blockchain에 값 저장
		}

	})

	// fmt.Printf("newesthash: %s\n height: %d\n", b.NewestHash, b.Height)
	return b
}

//txs에 모든 block의 tx를 역순으로 저장.
func Txs(b *blockchain) []*Tx {
	var txs []*Tx
	for _, block := range Blocks(b) {
		txs = append(txs, block.Transactions...)
	}
	return txs
}

//모든 Tx중 ID가 targetID와 같은 Tx를 찾아서 리턴
func FindTx(b *blockchain, targetID string) *Tx {
	for _, tx := range Txs(b) {
		if tx.Id == targetID {
			return tx
		}
	}
	return nil
}

//새로운 block을 생성 후, blockchain의 데이터를 변경. 그 후 새로워진 blockchain을 db에 저장 후 block을 리턴
func (b *blockchain) AddBlock() *Block { //새로운 블록 추가하는 함수
	block := createBlock(b.NewestHash, b.Height+1, difficulty(b)) //chain의 NewestHash가 Hash. Height
	b.NewestHash = block.Hash                                     // 새로운 블록의 hash 설정
	b.Height = block.Height                                       // 새로운 블록의 height 설정
	b.CurrDifficulty = block.Difficulty                           // 새로운 블록의 난이도 설정
	persistBlockchain(b)                                          //SaveBlockchain() 호출
	//블록이 생성될때마다 DB를 업데이트해주어야함
	return block
}

//blockchain을 []byte로 변환시켜서 db에 저장.
func persistBlockchain(b *blockchain) { //override. blockchain을 db에 저장하는 함수
	db.SaveBlockchain(utils.ToBytes(b))
}

//[]byte 형태였던 blockchain을 원래의 blockchain 형식으로 변환.
func (b *blockchain) fromBytes(data []byte) { // db에서 decoding해서 blockchain data로 변환 후 blockchain에 저장하는 함수
	utils.FromBytes(b, data)
}

//해당 blockchain의 block을 역순으로 찾아서 blocks에 대입 후 리턴. 가장 최근 block이 blocks[0].
func Blocks(b *blockchain) []*Block { //NewestHash로 prevHash를 갖는 블록을 찾고 그 prevHash로 또 전 블록찾고...해서 []*Block 리턴
	b.m.Lock()
	defer b.m.Unlock()
	var blocks []*Block
	hashCursor := b.NewestHash
	for {
		block, _ := FindBlock(hashCursor)
		blocks = append(blocks, block) //newest를 찾고 append하고 newest-1을 찾고 append하므로 가장 최근것이 blocks[0]에 온다
		if block.PrevHash != "" {      //Genesis block에 도달하기 전까지 blocks에 append
			hashCursor = block.PrevHash
		} else { // Genesis block. Genesis 전 블록은 없으므로 break
			break
		}

	}
	return blocks
}

// func (b *blockchain) txOuts() []*TxOut { //블록의 tx에 있는 txout을 저장
// 	var txOuts []*TxOut
// 	blocks := b.Blocks()
// 	for _, block := range blocks {
// 		for _, tx := range block.Transactions {
// 			txOuts = append(txOuts, tx.TxOuts...)
// 		}
// 	}

// 	return txOuts
// }

// func (b *blockchain) TxOutsByAddress(address string) []*TxOut { //해당 address에게 보내진 txout만을 모아서 저장
// txOuts := b.txOuts()
// var txOutsAddress []*TxOut
// for _, txOut := range txOuts {
// 	if txOut.Owner == address {
// 		txOutsAddress = append(txOutsAddress, txOut)
// 	}
// }

// return txOutsAddress
// }

//모든 TxOut에 대해 해당 address의 TxOut이 사용되었나 creatorTxs에 넣어서 판단함. (tx에는 없으나 mempool에 있을 수도 있으므로)사용되지 않은 TxOut을 초기화한 uTxOut에 대입해서 uTxOut으로 만들고 mempool에 있는지 확인해서 없으면 uTxOuts에 대입.
func UTxOutsByAddress(address string, b *blockchain) []*UTxOut { // address의 unspent tx outs
	var uTxOuts []*UTxOut
	creatorTxs := make(map[string]bool)
	for _, block := range Blocks(b) { //모든 블록 중
		for _, tx := range block.Transactions { // 모든 트랜잭션 중
			for _, input := range tx.TxIns { //모든 TxIn 중
				if input.Signature == "COINBASE" {
					break //signature가 coinbase면 break

				}
				if FindTx(b, input.TxID).TxOuts[input.Index].Address == address {
					creatorTxs[input.TxID] = true // TxIn의 Owner가 address와 같다면 -> 사용자가 TxIn에서 사용하고 있는 TxOut
				}
			}
			for index, output := range tx.TxOuts {
				// _, ok := creatorTxs[tx.Id]
				// if !ok{
				// }
				if output.Address == address { // 해당 주소의 txout
					if _, ok := creatorTxs[tx.Id]; !ok { // 해당 txout이 어느 txin에서도 참조되지 않았을 경우
						uTxOut := &UTxOut{
							TxID:   tx.Id,
							Index:  index,
							Amount: output.Amount,
						}
						if !isOnMempool(uTxOut) { //mempool에 없어야 uTxOut
							uTxOuts = append(uTxOuts, uTxOut)
						}

					}
				}

			}
		}
	}
	return uTxOuts

}

//해당 address를 가지고 사용되지 않은 TxOuts의 amount를 모두 더해서 리턴.
func BalanceByAddress(address string, b *blockchain) int { // 해당 address에게 보내진 amount를 계산해서 저장
	txOuts := UTxOutsByAddress(address, b) //사용되지 않은 txOuts
	var amount int
	for _, txOut := range txOuts {
		amount += txOut.Amount
	}

	return amount
}

// func (b *Block) getHash() {
// 	hash := sha256.Sum256([]byte(b.Data + b.PrevHash))
// 	b.Hash = fmt.Sprintf("%x", hash)

// }

// func getLastHash() string {
// 	totalBlocks := len(GetBlockchain().blocks)
// 	if totalBlocks == 0 {
// 		return "" // block의 숫자가 0 이면 prevhash 없음
// 	}
// 	return GetBlockchain().blocks[totalBlocks-1].Hash // 마지막 -1 의 hash 반환
// }

// func createBlock(data string) *Block {
// 	fmt.Println(len(GetBlockchain().blocks))
// 	newBlock := Block{data, "", getLastHash(), len(GetBlockchain().blocks) + 1} //GetBlockchain() -> AddBlock() -> createBlock()의 순이므로
// 	// createBlock()이 먼저 돌아감. 이때는 Genesis가 만들어지기 전이므로 len(GetBlockchain())이 0임
// 	newBlock.getHash()
// 	return &newBlock
// }

// func (b *blockchain) AddBlock(data string) {
// 	b.blocks = append(b.blocks, createBlock(data)) // genesisblock 생성

// }

// func (b *blockchain) AllBlocks() []*Block {
// 	return GetBlockchain().blocks
// }

// var ErrNotFound = errors.New("Block not found") //for error control

// func (b *blockchain) GetBlock(height int) (*Block, error) {
// 	if height > len(b.blocks) { //blocks의 길이보다 height가 더 크면 에러 출력
// 		return nil, ErrNotFound
// 	}
// 	return b.blocks[height-1], nil
// }

//blockchain을 보여줌.
func Status(b *blockchain, rw http.ResponseWriter) {
	b.m.Lock()
	defer b.m.Unlock()
	utils.HandleErr(json.NewEncoder(rw).Encode(b))
}

//역순으로 들어온 newBlocks의 newestBlock에서 hash, height, difficulty를 가져오고 blockchain에 저장. 그 후 db에 blockchain을 업데이트.
//blocks에 대해서도 db를 비우고 다시 newBlocks를 db에 저장.
func (b *blockchain) Replace(newBlocks []*Block) {
	b.m.Lock()
	defer b.m.Unlock()
	b.NewestHash = newBlocks[0].Hash
	b.Height = len(newBlocks)
	b.CurrDifficulty = newBlocks[0].Difficulty
	persistBlockchain(b) // db에 blockchain update

	//db에 새로운 blocks 저장
	db.EmptyBlocks()
	for _, block := range newBlocks {
		persistBlock(block)
	}

}

//blockchain의 height, hash, difficulty를 newestBlock의 것으로 바꾸고 db에 blockchain과 newBlock 저장.
//newBlock에 mempool의 tx 가 들어있으면(tx.id로 확인) mempool에서 tx 삭제.
func (b *blockchain) AddPeerBlock(newBlock *Block) { //새로 블록을 채굴할 때 실행
	b.m.Lock()
	m.m.Lock()
	defer b.m.Unlock()
	defer m.m.Unlock()
	b.Height += 1
	b.NewestHash = newBlock.Hash
	b.CurrDifficulty = newBlock.Difficulty
	persistBlockchain(b)
	persistBlock(newBlock)

	//mempool
	for _, tx := range newBlock.Transactions {
		_, ok := m.Txs[tx.Id]
		if ok {
			delete(m.Txs, tx.Id) //mempool에서 새로 채굴된 블록의 tx를 삭제해야함
		}

	}
}
