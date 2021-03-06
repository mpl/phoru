// Copyright 2016 Mathieu Lonjaret

// Given a (pseudo-)phonetic russian input, phoru outputs (cyrillic) russian.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/mpl/phoru"
	"github.com/mpl/simpletls"
)

// TODO(mpl): make it use github.com/mpl/phoru.Translate

var (
	flagHelp    = flag.Bool("h", false, "show this help")
	flagVerbose = flag.Bool("v", false, "verbose")

	flagTLS  = flag.Bool("tls", true, "Enable TLS for the http server mode.")
	flagHttp = flag.String("http", "", "Run in http server mode, on the given address.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "example:\n\t echo privet mir | phoru\n")
	fmt.Fprintf(os.Stderr, "\ntranslation tables:\n")
	var sortedSingle, sortedDouble, sortedTriple []string
	for k, _ := range phoru.Single {
		sortedSingle = append(sortedSingle, k)
	}
	for k, _ := range phoru.Double {
		sortedDouble = append(sortedDouble, k)
	}
	for k, _ := range phoru.Triple {
		sortedTriple = append(sortedTriple, k)
	}
	sort.Strings(sortedSingle)
	sort.Strings(sortedDouble)
	sort.Strings(sortedTriple)
	for _, v := range sortedSingle {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(phoru.Single[v]))
	}
	for _, v := range sortedDouble {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(phoru.Double[v]))
	}
	for _, v := range sortedTriple {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(phoru.Triple[v]))
	}
	fmt.Fprintf(os.Stderr, "\nin server mode:\n\t phoru -http=:6060\n")
	flag.PrintDefaults()
	os.Exit(2)
}

/*
The Russian alphabet
А а	Б б	В в	Г  г	Д д	Е е	Ё ё	Ж ж	З з	И и	Й й
К к	Л л	М м	Н н	О о	П п	Р р	С с	Т т	У у	Ф ф
Х х	Ц ц	Ч ч	Ш ш	Щ щ	Ъ ъ	Ы ы	Ь ь	Э э	Ю ю	Я я
*/

// TODO(mpl): use same pseudo-phonetics as goog translate maybe?
// TODO(mpl): greediness can break some cases. ex: with iy->й, and ya->я, when I
// want to write ия, I get йa instead. I need to rethink all bi and tri letter
// combinations. worst case scenario, I add delimiters (single quotes maybe?)
// around combinations. For now, changing iy into ï. I didn't hit any other
// breaking example so far, but we'll see when my vocabulary extends.

func main() {
	flag.Usage = usage
	flag.Parse()
	if *flagHelp {
		usage()
	}

	if *flagHttp != "" {
		baseData = &Translation{
			Single: phoru.Single,
			Double: phoru.Double,
			Triple: phoru.Triple,
		}
		tmpl = template.Must(template.New("root").Parse(HTML))
		http.HandleFunc("/phoru/", makeHandler(apiHandler))
		http.HandleFunc("/", makeHandler(rootHandler))
		if *flagTLS {
			listener, err := simpletls.Listen(*flagHttp)
			if err != nil {
				log.Fatal(err)
			}
			log.Fatal(http.Serve(listener, nil))
		}
		log.Fatal(http.ListenAndServe(*flagHttp, nil))
	}

	trans, err := phoru.Translate(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%v", string(trans))
}

const idstring = "http://golang.org/pkg/http/#ListenAndServe"

var (
	tmpl     *template.Template
	baseData *Translation
)

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
		}()
		title := r.URL.Path
		w.Header().Set("Server", idstring)
		fn(w, r, title)
	}
}

type Translation struct {
	Input  string
	Output string
	IsPost bool // TODO(mpl): meh. tired. do better later.
	Single map[string]rune
	Double map[string]rune
	Triple map[string]rune
}

func apiHandler(w http.ResponseWriter, r *http.Request, url string) {
	if r.URL.Path != "/phoru/" {
		http.Error(w, "not found", 404)
		return
	}
	latin := r.FormValue("q")
	if latin == "" {
		if _, err := w.Write([]byte("")); err != nil {
			log.Printf("error sending translation: %v", err)
		}
		return
	}
	cyril, err := phoru.Translate(strings.NewReader(latin))
	if err != nil {
		log.Printf("translation error: %v", err)
		http.Error(w, "translation error", 500)
		return
	}
	if _, err := w.Write([]byte(string(cyril))); err != nil {
		log.Printf("error sending translation: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, url string) {
	if r.Method == "GET" {
		if err := tmpl.Execute(w, baseData); err != nil {
			log.Printf("template error: %v", err)
		}
		return
	}
	if r.Method != "POST" {
		http.Error(w, "not a POST", http.StatusMethodNotAllowed)
		return
	}
	input := r.FormValue("inputtext")
	trans, err := phoru.Translate(strings.NewReader(input))
	if err != nil {
		log.Printf("translation error: %v", err)
		http.Error(w, "translation error", 500)
		return
	}
	data := &Translation{
		Input:  input,
		Output: string(trans),
		IsPost: true,
		Single: phoru.Single,
		Double: phoru.Double,
		Triple: phoru.Triple,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

var HTML = `
<!DOCTYPE html>
<html>
<head>
  <title>Phoru</title>
</head>
<body>
	<h1>Translate from pseudo-phonetics Russian, into Cyrillic Russian.</h1>

	<form action="/translate" method="POST" id="translateform" enctype="multipart/form-data">
	{{if .IsPost}}
	<textarea rows="5" cols="100" name="inputtext" form="translateform">{{.Input}}</textarea>
	{{else}}
	<textarea rows="5" cols="100" name="inputtext" form="translateform">Privet Mir!</textarea>
	{{end}}
    <input type="submit" id="textsubmit" value="Translate">
  </form>

	{{if .Output}}
	<h2>Conversion result</h1>
	<p>
	<pre>{{.Output}}</pre>
	</p>
	{{end}}

	<h2>Conversion table</h2>
	<p>
	<table style="border: 1px solid black; border-collapse: collapse">
	<tr>
	{{range $latin,$cyr := .Single}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{$latin}}</td>
	{{end}}
	</tr>
	<tr>
	{{range $latin,$cyr := .Single}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{printf "%c" $cyr}}</td>
	{{end}}
	</tr>
	</table>
	</p>
	<p>
	<table style="border: 1px solid black; border-collapse: collapse">
	<tr>
	{{range $latin,$cyr := .Double}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{$latin}}</td>
	{{end}}
	</tr>
	<tr>
	{{range $latin,$cyr := .Double}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{printf "%c" $cyr}}</td>
	{{end}}
	</tr>
	</table>
	</p>
	<p>
	<table style="border: 1px solid black; border-collapse: collapse">
	<tr>
	{{range $latin,$cyr := .Triple}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{$latin}}</td>
	{{end}}
	</tr>
	<tr>
	{{range $latin,$cyr := .Triple}}
	<td style="border: 1px solid black; padding: 7px; text-align: center">{{printf "%c" $cyr}}</td>
	{{end}}
	</tr>
	</table>
	</p>

	<p>
	Source code at: <a href="https://github.com/mpl/phoru">mpl/phoru</a>
	</p>

</body>
</html>
`
