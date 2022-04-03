package db

import (
	"fmt"
	"os"

	"github.com/yyuurriiaa/ProjectMSSP/utils"
	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB // singleton pattern

const (
	dbName       = "blockchain" //db 이름
	dataBucket   = "data"
	blocksBucket = "blocks"
	//bucket : table같은 것. 분류를 위해

	checkpoint = "checkpoint"
)

func getDbName() string {
	// for i, a := range os.Args {
	// 	fmt.Println(i, a)
	// }
	port := os.Args[2][6:]
	return fmt.Sprintf("%s_%s.db", dbName, port)

}

func DB() *bolt.DB { // singleton pattern으로 data bucket과 blocks bucket을 가지고 있는 db 생성
	if db == nil {
		dbPointer, err := bolt.Open(getDbName(), 0600, nil)
		db = dbPointer // initialize
		utils.HandleErr(err)
		err = db.Update(func(t *bolt.Tx) error {
			_, err := t.CreateBucketIfNotExists([]byte(dataBucket)) // bucket이 존재하지 않으면 생성. data bucket 생성
			utils.HandleErr(err)

			_, err = t.CreateBucketIfNotExists([]byte(blocksBucket)) // blocks bucket todtjd
			return err                                               // error를 반환해야하기때문에 error handling을 다 하지 않고 return
		})

		utils.HandleErr(err) // 위에서 받은 err handling
	}

	return db
}

func SaveBlock(hash string, data []byte) { //block bucket에 key : Hash, value : data 형으로 저장
	// fmt.Printf("Saving Block Hash: %s\nData: %b\n", hash, data)
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(blocksBucket))
		err := bucket.Put([]byte(hash), data) // bucket에 저장
		return err
	})

	utils.HandleErr(err)
}

func SaveBlockchain(data []byte) { // data bucket에 blockchain 저장
	// fmt.Printf("Saving Blockchain\n Data: %x\n", data)
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(dataBucket))
		err := bucket.Put([]byte(checkpoint), data) //key : dummy값, value : blockchain 데이터
		return err
	})

	utils.HandleErr(err)
}

func Checkpoint() []byte { // 해당 블록체인의 checkpoint의 hash값을 리턴하는 함수
	var data []byte
	DB().View(func(t *bolt.Tx) error { //읽기전용으로
		bucket := t.Bucket([]byte(dataBucket)) // dataBucket이 byte형식으로 저장된 이름의 Bucket을 bucket이라고 지정
		data = bucket.Get([]byte(checkpoint))  // bucket에서 checkpoint가 byte형식으로 저장된 key값의 value를 가져옴
		return nil
	})

	return data
}

func Block(hash string) []byte { //Checkpoint와 동일. 해당 블록의 hash값을 리턴하는 함수
	var data []byte
	DB().View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(blocksBucket))
		data = bucket.Get([]byte(hash))
		return nil
	})

	return data
}

func Close() { // DB 열었던거 닫기
	DB().Close()
}

func EmptyBlockss() { // bucket 비우기
	DB().Update(func(t *bolt.Tx) error {
		utils.HandleErr(t.DeleteBucket([]byte(blocksBucket))) //blocksBucket 이름을 가진 bucket 삭제
		_, err := t.CreateBucket([]byte(blocksBucket))        //다시 생성
		utils.HandleErr(err)
		return nil
	})

}
