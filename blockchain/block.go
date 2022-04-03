package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/yyuurriiaa/ProjectMSSP/db"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

// const difficulty int = 2 // 2개의 0을 시작으로 하는 hash

type Block struct {
	// Data       string `json:"data"`
	Transactions []*Tx  `json:"transactions"`
	Hash         string `json:"hash"`
	PrevHash     string `json:"prevHash,omitempty"`
	Height       int    `json:"height"`     // rest api에서 /blocks/Height 식으로 접근
	Difficulty   int    `json:"difficulty"` //n개의 0을 앞에 가지는 hash
	Nonce        int    `json:"nonce"`      // 채굴자들이 변경할 수 있는 유일한 값. Nonce를 변경해서 n개의 0을 가지는 hash를 찾는다
	Timestamp    int    `json:"timestamp"`
}

var ErrNotFound = errors.New("Block not found")

func createBlock(prevHash string, height int, diff int) *Block { //블록 생성하는 함수
	block := Block{
		//Data:       data,
		//Transactions: []*Tx{makeCoinbaseTx("OMT")}, TxToConfirm()에 들어가잇음
		Hash:       "",
		PrevHash:   prevHash,
		Height:     height,
		Difficulty: diff,
		Nonce:      0,
		Timestamp:  0, // 빈 블록을 만들 때 시간을 설정하면 안됨.
	}
	// payload := block.Data + block.PrevHash + fmt.Sprint(block.Height)
	// block.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(payload))) //payload hashing
	block.Transactions = Mempool().TxToConfirm()
	block.mine()
	persistBlock(&block)
	return &block
}

func (b *Block) mine() { // nonce값 변경해가면서 난이도에 맞는 block 채굴
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Timestamp = int(time.Now().Unix()) //Unix : int64로 return.
		hash := utils.GetHash(b)
		if strings.HasPrefix(hash, target) {
			b.Hash = hash
			// fmt.Printf("Hash:%s, target:%s, nonce:%d", hash, target, b.Nonce)
			break
		} else {
			b.Nonce += 1
		}

	}

}

func persistBlock(b *Block) { //override. db에 block을 저장하는 함수
	db.SaveBlock(b.Hash, utils.ToBytes(b))
}

func (b *Block) fromBytes(data []byte) { //data를 decoding해서 block으로 저장하는 함수
	utils.FromBytes(b, data)
}

func FindBlock(hash string) (*Block, error) { // 특정 hash값을 가지는 block을 찾는 함수
	blockBytes := db.Block(hash)
	if blockBytes == nil {
		return nil, ErrNotFound
	}
	block := &Block{}           //block을 Block type으로 초기화
	block.fromBytes(blockBytes) // 결국 block은 새로 만들어진 것이고 새로만들어진 block에 blockBytes를 디코딩 한 값을 넣어서 리턴
	return block, nil
}
