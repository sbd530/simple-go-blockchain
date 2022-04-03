package explorer

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/yyuurriiaa/ProjectMSSP/blockchain"
)

const (
	templateDir string = "explorer/templates/"
)

var templates *template.Template

type homeData struct {
	PageTitle string
	Blocks    []*blockchain.Block
}

func add(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templates.ExecuteTemplate(rw, "add", nil)
	case "POST":
		//r.ParseForm()
		//data := r.FormValue("blockData")
		blockchain.Blockchain().AddBlock()
		http.Redirect(rw, r, "/", http.StatusPermanentRedirect)
	}
}

func home(rw http.ResponseWriter, r *http.Request) {
	//Fprint : Writer에 출력
	// tmpl, err := template.ParseFiles("templates/home.html")
	// if err != nil {
	// 	log.Fatal(err)

	// }
	//or
	//tmpl := template.Must(template.ParseFiles("templates/pages/home.gohtml")) // template.Must가 위의 error처리와 같은 기능을 함

	// data := homeData{"Home", blockchain.GetBlockchain().AllBlocks()}
	data := homeData{"Home", nil}
	//tmpl.Execute(rw, data)
	templates.ExecuteTemplate(rw, "home", data) // 이름이 home인 templates 실행
}

func Start(port int) {
	handler := http.NewServeMux() //handler 새로 설정

	templates = template.Must(template.ParseGlob(templateDir + "pages/*.gohtml"))     //load files
	templates = template.Must(templates.ParseGlob(templateDir + "partials/*.gohtml")) // 추가로 load files. templates.ParseGlob에 주의

	handler.HandleFunc("/", home)   // http.HandleFunc -> handler.HandleFunc. explorer의 http.HandleFunc의 "/"와 링크가 겹치므로 실행불가능
	handler.HandleFunc("/add", add) //http.HandleFunc -> handler.HandleFunc
	fmt.Printf("listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler)) // ListenandServe로 서버 생성, Fatal : http.ListenAndServe(error반환)가 nil일경우 그냥 실행, error있을경우 os exit 1
	//handler 설정했으므로 nil -> handler로 변경
}
