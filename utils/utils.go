package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

var logFn = log.Panic

//Error handler
func HandleErr(err error) {
	if err != nil {
		logFn(err)
	}
}

//[]byte형식으로 변환. interface -> []byte
func ToBytes(i interface{}) []byte { //interface{} : 어떤 타입이든 다 받음. 인코딩
	var aBuffer bytes.Buffer            // bytes 변수. buffer : bytes를 r/w 할 수 있음
	encoder := gob.NewEncoder(&aBuffer) //aBuffer에 encoding저장한다고 선언 writer
	HandleErr(encoder.Encode(i))        //i를 encoding한것을 aBuffer에 저장
	return aBuffer.Bytes()
}

//[]byte형식을 받아서 원래 형태로 변환. []byte -> interface
func FromBytes(i interface{}, data []byte) { //디코딩함수
	decoder := gob.NewDecoder(bytes.NewReader(data)) //data를 읽어서 디코딩 reader
	HandleErr(decoder.Decode(i))                     // i에 저장

}

//interface -> string -> []byte ->[32]byte 해싱 -> 해싱한걸 다시 string으로 변환.
func GetHash(i interface{}) string { //해싱함수
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(i))))
	// fmt.Println("hashed data : ", hash)
	return hash
}

//string을 sep으로 나누는 함수. index로 []중 몇번째인지 선택할 수 있음
func Splitter(s string, sep string, index int) string {
	r := strings.Split(s, sep)
	if index > len(r)-1 {
		return ""
	} else {
		return r[index]
	}
}

//interface 형식을 받아서 json으로 변환. interface -> json
func ToJSON(i interface{}) []byte {
	r, err := json.Marshal(i)
	HandleErr(err)
	return r
}
