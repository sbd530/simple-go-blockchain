package cli

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/yyuurriiaa/ProjectMSSP/explorer"
	"github.com/yyuurriiaa/ProjectMSSP/rest"
)

func usage() {
	fmt.Printf("welcome to MSSP!\n\n")
	fmt.Printf("please use the following flags:\n\n")
	fmt.Printf("-port=4000 : set the port of the server\n")
	fmt.Printf("-mode=rest : start the REST API(recommended)\n")
	//os.Exit(1) //강제종료. error code 1
	runtime.Goexit() //모든 함수 제거(defer 먼저 실행 후)
}

func Start() {
	/////////////////////////////////////////////////////
	/////////////////////더 많은 기능 사용하려면 cobra CLI/////
	/////////////////////////////////////////////////////
	// fmt.Println(os.Args[2:])
	// if len(os.Args) < 2 {
	// 	usage()
	// }

	port := flag.Int("port", 4000, "Set port of the server") //4000이 default

	mode := flag.String("mode", "rest", "Choose between 'html' and 'rest'") //rest가 default

	flag.Parse()

	switch *mode {
	case "rest":
		rest.Start(*port)
	case "html":
		explorer.Start(*port)
	default:
		usage()
	}
	fmt.Println(*port, *mode)
	//flag : os.Args에서 각각의 요소. [/var/folders/lm/m3glhds93412266yqvhrwn1h0000gn/T/go-build2223486451/b001/exe/main rest] 는 두개의 flag로 이루어져있음
	// restCommand := flag.NewFlagSet("rest", flag.ExitOnError) //FlagSet : 어떤 command가 어떤 flag를 가질 것인지 알려주는 역할. Flag가 많을 때 좋음

	// portFlag := restCommand.Int("port", 4000, "Sets the port of the server") //restCommand 라는 FlagSet의 port라는 이름을 가진 Flag.
	//Int type으로 설정되었고 int값이 들어오지 않았을 때 유저에게 usage를 보냄

	// switch os.Args[1] {
	// case "explorer":
	// 	fmt.Println("Start Explorer")
	// case "rest":
	// 	//fmt.Println("Start REST API")
	// 	restCommand.Parse(os.Args[2:]) // 실행되고 자동적으로 restCommand라는 FlagSet에서 "port"를 찾음(FlagSet에 속해있는 Flag이기때문에)
	// 	//CLI에서 커맨드들의 순서가 바뀌어도 똑같이 실행될 수 있음
	// default:
	// 	usage()
	// }

	// if restCommand.Parsed() { //restCommand가 Parse되었다면
	// 	fmt.Println("Start the server!")
	// 	fmt.Println(*portFlag)
	// }
}
