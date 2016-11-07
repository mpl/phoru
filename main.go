// Copyright 2016 Mathieu Lonjaret

// Given a (pseudo-)phonetic russian input, phoru outputs (cyrillic) russian.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	//	"net/http"
	"os"
	"sort"
	"strings"
	"unicode"
)

// TODO(mpl): run gopherjs on it, and see if we can somehow import and use the
// resulting js in a Google Apps script.

var (
	flagHelp    = flag.Bool("h", false, "show this help")
	flagVerbose = flag.Bool("v", false, "verbose")

//	flagHttp = flag.String("http", "", "Run in http server mode, on the given address.")
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
//	fmt.Fprintf(os.Stderr, "\nin server mode:\n\t phoru -http=:6060\n")
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
		"a": 'a',
		"b": 'б',
		"v": 'в',
		"g": 'г',
		"d": 'д',
		"e": 'e',
		"j": 'ж',
		"i": 'и',
		"ï": 'й',
		"z": 'з',
		"k": 'к',
		"l": 'л',
		"m": 'm',
		"n": 'н',
		"o": 'o',
		"p": 'п',
		"r": 'p',
		"s": 'c',
		"t": 't',
		"u": 'y',
		"f": 'ф',
		"î": 'ы',
		"è": 'э',
	}
	double = map[string]rune{
		"yo": 'ë',
		"kh": 'х',
		"ts": 'ц',
		"ch": 'ч',
		"sh": 'ш',
		"ya": 'я',
		"i_": 'й', // redundant with ï for non-accented keymaps
		"i-": 'ы', // redundant with î for non-accented keymaps
		"`e": 'э', // redundant with è for non-accented keymaps
	}
	triple = map[string]rune{
		"shh": 'щ',
		"you": 'ю',
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if *flagHelp {
		usage()
	}

	// TODO(mpl): browser mode (for windows users)
	/*
		if *flagHttp != "" {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Hello, you")
			})
			log.Fatal(http.ListenAndServe(*flagHttp, nil))
		}
	*/

	sc := bufio.NewScanner(os.Stdin)
	var skipTwo, skipOne bool
	for sc.Scan() {
		words := strings.Fields(sc.Text())
		for j, word := range words {
			if j != 0 {
				fmt.Fprintf(os.Stdout, " ")
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
			fmt.Fprintf(os.Stdout, "%v", string(trans))
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
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
