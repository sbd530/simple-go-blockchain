package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/yyuurriiaa/ProjectMSSP/utils"
)

//"data" + PrivateKey = Signature
//"data" + Signature + PublicKey = Verify(True / False)

const (
	walletName string = "MSSP.wallet"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

func hasWalletFile() bool { // wallet파일이 있나 검사
	_, err := os.Stat(walletName) // wallet파일이 없으면 err 발생
	return !os.IsNotExist(err)    // 에러가 'notexist'에러이면 true 반환이므로 역을 취함
}

var w *wallet // singleton

func Wallet() *wallet { // wallet 생성
	if w == nil {
		w = &wallet{}
		if hasWalletFile() { //지갑 파일이 있으면 지갑파일로부터 복원
			w.privateKey = restoreKey(walletName)
		} else { //지갑파일이 없으면 새로 생성
			key := createPublicKey()
			persistKey(key) // wallet 파일 r/w
			w.privateKey = key
		}
		w.Address = addressFromKey(w.privateKey)
	}
	return w
}

func createPublicKey() *ecdsa.PrivateKey { //publicKey 생성 후 반환
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleErr(err)
	return privateKey
}

func persistKey(key *ecdsa.PrivateKey) { // wallet 생성
	bytes, err := x509.MarshalECPrivateKey(key)
	utils.HandleErr(err)
	err = os.WriteFile(walletName, bytes, 0644) //0644 = r/w 파일 생성
	utils.HandleErr(err)
}

func restoreKey(walletName string) *ecdsa.PrivateKey {
	KeyAsBytes, err := os.ReadFile(walletName)
	utils.HandleErr(err)
	privateKey, err := x509.ParseECPrivateKey(KeyAsBytes)
	utils.HandleErr(err)
	return privateKey
}

// func addressFromKey(key *ecdsa.PrivateKey) string {
// 	x := key.X.Bytes() // big Int -> []byte
// 	y := key.Y.Bytes()
// 	addressBytes := append(x, y...)
// 	var address string
// 	utils.FromBytes(address, addressBytes)
// 	return address
// }

func bytesToHex(a []byte, b []byte) string { // []byte -> Hex
	z := append(a, b...)
	zHex := fmt.Sprintf("%x", z)
	return zHex
}

func addressFromKey(key *ecdsa.PrivateKey) string {
	z := bytesToHex(key.X.Bytes(), key.Y.Bytes())
	return z
}

func Sign(payload string, w *wallet) string { // payload : "Data"
	payloadAsBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, payloadAsBytes)
	utils.HandleErr(err)
	signature := bytesToHex(r.Bytes(), s.Bytes())
	return signature
}

func restoreBigInt(payload string) (*big.Int, *big.Int, error) {
	signatureBytes, err := hex.DecodeString(payload) // 16진수 string -> []byte
	if err != nil {
		return nil, nil, err
	}

	aBytes := signatureBytes[:len(signatureBytes)/2]
	bBytes := signatureBytes[len(signatureBytes)/2:]
	BigA := big.Int{}
	BigB := big.Int{}
	BigA.SetBytes(aBytes)
	BigB.SetBytes(bBytes)
	fmt.Println(BigA, BigB)
	return &BigA, &BigB, nil
}

func Verify(signature string, payload string, address string) bool {
	r, s, err := restoreBigInt(signature)
	utils.HandleErr(err)

	payloadBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)

	x, y, err := restoreBigInt(address)
	utils.HandleErr(err)
	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	ok := ecdsa.Verify(&publicKey, payloadBytes, r, s)
	return ok

}

// const (
// 	hashedMessage = "3ca99a79f64dddf0908a601da81512f29012ced12861eb1b26ab5719f60eb08b"
// 	privateKey    = "307702010104207a99790fdd329cf85c0e239e363b06e2edecc8bad76c984f798a97d6a952cddda00a06082a8648ce3d030107a14403420004cb40f305132b17f6cbd1cd5a5b582df9d20fe9e5986b9f449c33df7f758584e420c45c84b622f0745393d3a74560543720bc1dd88029318e10e7347100dc9613"
// 	signature     = "87b01d597e6f9876b347865e7370fd8565e11e2c2f82aff0daa1b24b72bb3c9c764eb9f56009491d185839e455ec0d524eb5ca96a201da486dc5a17d1a27e9f9"
// )

// func Start() {
// 	privateKeyAsBytes, err := hex.DecodeString(privateKey) //privateKey가 16진수 string인지 확인
// 	utils.HandleErr(err)
// 	restoredPrivateKey, err := x509.ParseECPrivateKey(privateKeyAsBytes) // []byte를 받고 privatekey 형식으로 return
// 	utils.HandleErr(err)
// 	fmt.Println("restoredKey:", restoredPrivateKey)

// 	signatureAsBytes, err := hex.DecodeString(signature)
// 	utils.HandleErr(err)
// 	fmt.Println("signatureAsBytes:", signatureAsBytes)
// 	rBytes := signatureAsBytes[:len(signatureAsBytes)/2]
// 	sBytes := signatureAsBytes[len(signatureAsBytes)/2:]
// 	fmt.Printf("r: %x\ns: %x\n", rBytes, sBytes)

// 	var bigR, bigS = big.Int{}, big.Int{}
// 	bigR.SetBytes(rBytes)
// 	bigS.SetBytes(sBytes)
// 	fmt.Println("bigR:", bigR, "\nbigS:", bigS)

// 	bytesMessage, err := hex.DecodeString(hashedMessage)
// 	utils.HandleErr(err)

// 	ok := ecdsa.Verify(&restoredPrivateKey.PublicKey, bytesMessage, &bigR, &bigS)
// 	fmt.Println(ok)

// }
