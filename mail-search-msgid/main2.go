package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

func init() {
	//tpl = template.Must(template.Parse("templates/index.html"))
	tpl = template.Must(template.ParseGlob("templates/*"))

}

type person struct {
	FirstName  string
	LastName   string
	Subscribed bool
}

func main() {
	http.HandleFunc("/", foo)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	fmt.Println("Starting Server on port 8090")
	http.ListenAndServe(":8090", nil)
}

func foo(w http.ResponseWriter, req *http.Request) {
	//template loads .. initially the form values are empty...
	//then when submitting from the webform... the form fields are set and sent
	//back into the webpage.... the form field submission loads the / default page
	//which loads the template again sending the data to the webform...

	f := req.FormValue("first")
	l := req.FormValue("last")
	s := req.FormValue("subscribe") == "on" // boolean test

	err := tpl.ExecuteTemplate(w, "index.html", person{f, l, s})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}

}
