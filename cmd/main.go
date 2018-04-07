// Copyright 2016 Mathieu Lonjaret

// Given a (pseudo-)phonetic russian input, phoru outputs (cyrillic) russian.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"unicode"

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
	for k, _ := range single {
		sortedSingle = append(sortedSingle, k)
	}
	for k, _ := range double {
		sortedDouble = append(sortedDouble, k)
	}
	for k, _ := range triple {
		sortedTriple = append(sortedTriple, k)
	}
	sort.Strings(sortedSingle)
	sort.Strings(sortedDouble)
	sort.Strings(sortedTriple)
	for _, v := range sortedSingle {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(single[v]))
	}
	for _, v := range sortedDouble {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(double[v]))
	}
	for _, v := range sortedTriple {
		fmt.Fprintf(os.Stderr, "\t %v : %v\n", v, string(triple[v]))
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

var (
	single = map[string]rune{
		"a": 'а',
		"b": 'б',
		"v": 'в',
		"g": 'г',
		"d": 'д',
		"e": 'е',
		"j": 'ж',
		"i": 'и',
		"ï": 'й',
		"z": 'з',
		"k": 'к',
		"l": 'л',
		"m": 'м',
		"n": 'н',
		"o": 'о',
		"p": 'п',
		"r": 'р',
		"s": 'с',
		"t": 'т',
		"u": 'у',
		"f": 'ф',
		"î": 'ы',
		"è": 'э',
	}
	double = map[string]rune{
		"yo": 'ё',
		"kh": 'х',
		"ch": 'ч',
		"sh": 'ш',
		"ya": 'я',
		"b-": 'ь',
		"i_": 'й', // redundant with ï for non-accented keymaps
		"i-": 'ы', // redundant with î for non-accented keymaps
		"`e": 'э', // redundant with è for non-accented keymaps
	}
	triple = map[string]rune{
		"shh": 'щ',
		"you": 'ю',
		// TODO(mpl): find a better solution for the "тс" VS "ц"
		// conflict. Especially since ц is way more frequent than тс (in my
		// limited experience).
		"ts-": 'ц',
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if *flagHelp {
		usage()
	}

	if *flagHttp != "" {
		baseData = &Translation{
			Single: single,
			Double: double,
			Triple: triple,
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

	trans, err := phoru(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "%v", string(trans))
}

func phoru(r io.Reader) ([]rune, error) {
	var out []rune
	sc := bufio.NewScanner(r)
	var skipTwo, skipOne bool
	firstLine := true
	for sc.Scan() {
		if !firstLine {
			out = append(out, '\n')
		} else {
			firstLine = false
		}
		words := strings.Fields(sc.Text())
		for j, word := range words {
			if j != 0 {
				out = append(out, rune(' '))
			}
			var runes, trans []rune
			for _, v := range word {
				runes = append(runes, v)
			}
			for i, r := range runes {
				isUpper := false
				if skipTwo {
					skipTwo = false
					skipOne = true
					continue
				}
				if skipOne {
					skipOne = false
					continue
				}
				if *flagVerbose {
					fmt.Fprintf(os.Stderr, "%v", string(r))
				}
				isUpper = unicode.IsUpper(r)
				if isUpper {
					r = unicode.ToLower(r)
					runes[i] = r
				}
				cyril, n := toCyrillic(runes, i)
				if isUpper {
					cyril = unicode.ToUpper(cyril)
				}
				trans = append(trans, cyril)
				// TODO(mpl): something more elegant?
				if n == 3 {
					skipTwo = true
				} else if n == 2 {
					skipOne = true
				}
				if *flagVerbose {
					fmt.Fprintf(os.Stderr, " -> %v\n", string(trans[len(trans)-1:]))
				}
			}
			out = append(out, trans...)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// toCyrillic converts to a cyrillic rune the next rune from runes, starting at
// runes[i]. It returns the converted rune, as well as the number of runes that
// were "consumed" from runes.
func toCyrillic(runes []rune, index int) (cyril rune, read int) {
	i := index
	switch {
	case len(runes[i:]) > 2:
		if cyril, ok := triple[string(runes[i:i+3])]; ok {
			return cyril, 3
		}
		fallthrough
	case len(runes[i:]) > 1:
		if cyril, ok := double[string(runes[i:i+2])]; ok {
			return cyril, 2
		}
		fallthrough
	case len(runes[i:]) == 1:
		if cyril, ok := single[string(runes[i:i+1])]; ok {
			return cyril, 1
		}
		fallthrough
	default:
		log.Printf("unknown rune: %v", string(runes[i]))
		return runes[i], 1
	}
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
	cyril, err := phoru(strings.NewReader(latin))
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
	trans, err := phoru(strings.NewReader(input))
	if err != nil {
		log.Printf("translation error: %v", err)
		http.Error(w, "translation error", 500)
		return
	}
	data := &Translation{
		Input:  input,
		Output: string(trans),
		IsPost: true,
		Single: single,
		Double: double,
		Triple: triple,
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
