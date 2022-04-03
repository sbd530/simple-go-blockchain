package main

import (
	"github.com/yyuurriiaa/ProjectMSSP/cli"
	"github.com/yyuurriiaa/ProjectMSSP/db"
)

func main() {
	defer db.Close()
	cli.Start()

	// blockchain.Blockchain()
	// blockchain.Blockchain().AddBlock("First Block")
	// blockchain.Blockchain().AddBlock("Second Block")
	// wallet.Wallet()
	// fmt.Println("hello")
}
