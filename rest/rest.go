package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yyuurriiaa/ProjectMSSP/blockchain"
	"github.com/yyuurriiaa/ProjectMSSP/p2p"
	"github.com/yyuurriiaa/ProjectMSSP/utils"
	"github.com/yyuurriiaa/ProjectMSSP/wallet"
)

type url string

var port string

func (u url) MarshalText() ([]byte, error) { //TextMarshaler interface, https://cafemocamoca.tistory.com/288 참고 url을 []byte로 변환
	url := fmt.Sprintf("http://localhost%s%s", port, u)
	return []byte(url), nil
}

type errorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type urlDescription struct {
	URL         url    `json:"url"`    // json에서 표시할 형식 변경
	Method      string `json:"method"` // omitempty : 값이 비어있으면 생략해줌
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

type balanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type addTxPayload struct {
	To     string
	Amount int
}

type myWalletResponse struct {
	Address string `json:"address"`
}

type addPeerPayload struct {
	Address string
	Port    string
}

// type URLDescriptionSlice struct {
// 	URLSlice []URLDescription
// }

// func (udl *URLDescriptionSlice) AddURL(u URLDescription) []URLDescription {
// 	udl.URLSlice = append(udl.URLSlice, u)
// 	return udl.URLSlice
// }

//////http구조//////////////
//클라이언트  ------->>  서버 : HTTP Request
//서버 ------->> 클라이언트 : HTTP Response
func documentation(rw http.ResponseWriter, r *http.Request) { //documentation.
	data := []urlDescription{
		{
			URL:         url("/"), // URL 형식으로 받아야 하니까 []byte(aaaa)처럼 한듯?
			Method:      "GET",
			Description: "Check Documentation",
		},
		{
			URL:         url("/blocks"),
			Method:      "POST",
			Description: "Add a block",
			Payload:     "data:string",
		},
		{
			URL:         url("/blocks"),
			Method:      "GET",
			Description: "Check All block",
		},
		{
			URL:         url("/status"),
			Method:      "GET",
			Description: "See the status of Blockchain",
		},
		{
			URL:         url("/blocks/{hash}"),
			Method:      "GET",
			Description: "Search a block",
		},
		{
			URL:         url("/balance/{address}"),
			Method:      "GET",
			Description: "Get TxOuts for an Address",
		},
		{
			URL:         url("/ws"),
			Method:      "GET",
			Description: "Upgrade to WebSockets",
		},
	}
	// datas := URLDescriptionSlice{}
	// datas.AddURL(data)

	//rw.Header().Add("Content-Type", "application/json") //browser의 Header에 보내는 Content-Type을 json으로 변경. middleware로 설정
	//1번방법
	// b, err := json.Marshal(datas)
	// utils.HandleErr(err)
	// fmt.Fprintf(rw, "%s", b)

	//2번방법
	utils.HandleErr(json.NewEncoder(rw).Encode(data)) //data를 json으로 encode
	//NewEncoder(w) : w에 쓸 encoder를 반환
	//Encode(v) : v를 인코딩함
	// -> v를 인코딩하고 rw에 써서 json으로 반환
}

// func (u URLDescription) String() string { //Stringer() interface
// 	return "URL Description"
// }

// type addBlockBody struct {
// 	Message string // 대문자여야 api.http의 message와 통신? 가능
// }

func jsonContentTypeMiddleware(next http.Handler) http.Handler { //handler : HTTP request에 response함
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { //HandlerFunc : 일반 함수를 handler처럼 쓸수잇게해줌
		rw.Header().Add("Content-Type", "application/json") //서버가 보내는 header에 add의 내용을 추가함
		next.ServeHTTP(rw, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		next.ServeHTTP(rw, r)
	})
}

func blocks(rw http.ResponseWriter, r *http.Request) { //
	switch r.Method { //HTTP에 보내는 request의 종류에 따라 구분
	case "POST": //to use create a block

		//var addBlockBody addBlockBody
		//fmt.Println(addBlockBody)
		//utils.HandleErr(json.NewDecoder(r.Body).Decode(&addBlockBody))
		//fmt.Println(addBlockBody)
		newBlock := blockchain.Blockchain().AddBlock() //블록생성하고
		p2p.BroadcastNewBlock(newBlock)
		rw.WriteHeader(http.StatusCreated) //헤더에 http.statuscreated 쓰기
		// default:
		// 	rw.WriteHeader(http.StatusMethodNotAllowed)//mux에서 Methods 정하는걸 안햇으면 필요함

	case "GET": //to get all of block information

		//rw.Header().Add("Content-Type", "application/json") //browser의 Header에 보내는 Content-Type을 json으로 변경. middleware로 설정
		utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.Blocks(blockchain.Blockchain()))) //블록들을 인코딩해서 rw에 쓰기

	}

	//왜인지 모르는데 Method "POST"랑 "GET"이랑 구분을 못함
}

func block(rw http.ResponseWriter, r *http.Request) { //gorilla mux 사용.
	vars := mux.Vars(r)
	//id := vars["height"]
	// fmt.Println(id)
	hash := vars["hash"]

	block, err := blockchain.FindBlock(hash)
	if err == blockchain.ErrNotFound {
		utils.HandleErr(json.NewEncoder(rw).Encode(errorResponse{err.Error()})) // type error -> type string으로 바꿔서 Encode에 넣고 json으로 변환
	} else {
		utils.HandleErr(json.NewEncoder(rw).Encode(block))
	}
}

func status(rw http.ResponseWriter, r *http.Request) {
	// utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.Blockchain())) // blockchain을 encoding
	blockchain.Status(blockchain.Blockchain(), rw) // mutex
}

func balance(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	total := r.URL.Query().Get("total")
	switch total {
	case "true":
		amount := blockchain.BalanceByAddress(address, blockchain.Blockchain())
		utils.HandleErr(json.NewEncoder(rw).Encode(balanceResponse{address, amount}))
	default:
		utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.UTxOutsByAddress(address, blockchain.Blockchain()))) // UTx로 수정

	}

}

func mempool(rw http.ResponseWriter, r *http.Request) {
	// utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.Mempool().Txs)) //mempool의 txs를 json으로 인코딩해서 rw에 저장
	blockchain.MempoolMutex(blockchain.Mempool(), rw) // Mutex
}

func transactions(rw http.ResponseWriter, r *http.Request) {
	var payload addTxPayload
	utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload)) //body내용을 payload에 저장
	tx, err := blockchain.Mempool().AddTx(payload.To, payload.Amount)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorResponse{err.Error()})
		return //?
	}
	p2p.BroadcastNewTx(tx)
	rw.WriteHeader(http.StatusCreated)
}

func myWallet(rw http.ResponseWriter, r *http.Request) {
	address := wallet.Wallet().Address
	utils.HandleErr(json.NewEncoder(rw).Encode(myWalletResponse{Address: address}))
}

func peers(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var payload addPeerPayload
		json.NewDecoder(r.Body).Decode(&payload)
		p2p.AddPeer(payload.Address, payload.Port, port[1:], true)
		rw.WriteHeader(http.StatusOK)
	case "GET":
		json.NewEncoder(rw).Encode(p2p.AllPeers(&p2p.Peers))
	}
}

func Start(portnum int) {
	//handler := http.NewServeMux() //rest.go와 동일 설정. multiplexer

	port = fmt.Sprintf(":%d", portnum) //cli 시작할 때 port 숫자 입력으로 포트를 설정하고 시작
	router := mux.NewRouter()          //멀티플렉서의 역할은 경로를 특정 핸들러와 일치시키는 것. https://thebook.io/006806/ch08/03/ 참고

	router.Use(jsonContentTypeMiddleware, loggerMiddleware) //middleware를 체인에 추가함
	router.HandleFunc("/", documentation).Methods("GET")    //Handle()과 HandleFunc() 메서드는 요청된 Request Path에 어떤 Request 핸들러를 사용할 지를 지정하는 라우팅 역활을 한다.
	//http.Handler 인터페이스 : 다음과 같은 메서드 하나를 갖는 인터페이스 type Handler interface {
	//													   	 	  ServeHTTP(ResponseWriter, *Request)}

	//Handle
	router.HandleFunc("/blocks", blocks).Methods("POST", "GET") // /blocks 경로에 handler blocks를 출력.
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET")
	router.HandleFunc("/status", status)
	router.HandleFunc("/balance/{address}", balance).Methods("GET")
	router.HandleFunc("/mempool", mempool).Methods("GET")
	router.HandleFunc("/wallet", myWallet).Methods("GET")
	router.HandleFunc("/transactions", transactions).Methods("POST")
	router.HandleFunc("/ws", p2p.Upgrade).Methods("GET") //ws로 업그레이드
	router.HandleFunc("/peers", peers).Methods("GET", "POST")

	fmt.Printf("Listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router)) //ListenAndServe() 메서드는 지정된 포트에 웹 서버를 열고 클라이언트 Request를 받아들여 새 Go 루틴에 작업을 할당하는 일을 한다

}
