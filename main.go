// Copyright 2016 Mathieu Lonjaret

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	flagHelp    = flag.Bool("h", false, "show this help")
	flagVerbose = flag.Bool("v", false, "verbose")
)

func usage() {
	fmt.Fprintf(os.Stderr, "\t echo privyét mir | phoru\n")
	flag.PrintDefaults()
	os.Exit(2)
}

/*
The Russian alphabet
А а	Б б	В в	Г  г	Д д	Е е	Ё ё	Ж ж	З з	И и	Й й
К к	Л л	М м	Н н	О о	П п	Р р	С с	Т т	У у	Ф ф
Х х	Ц ц	Ч ч	Ш ш	Щ щ	Ъ ъ	Ы ы	Ь ь	Э э	Ю ю	Я я
*/

// TODO(mpl): uppercase
// TODO(mpl): use same pseudo-phonetics as goog translate.

var (
	single = map[string]rune{
		"a": 'a',
		"b": 'б',
		"v": 'в',
		"g": 'г',
		"d": 'д',
		"j": 'ж',
		"i": 'и',
		"z": 'з',
		"k": 'к',
		"l": 'д',
		"m": 'm',
		"n": 'н',
		"o": 'o',
		"p": 'п',
		"r": 'p',
		"s": 'c',
		"t": 't',
		"f": 'ф',
		"î": 'ы',
		"è": 'э',
	}
	double = map[string]rune{
		"yé": 'e',
		"yo": 'ë',
		"iy": 'й',
		"ou": 'y',
		"kh": 'х',
		"ts": 'ц',
		"ch": 'ч',
		"sh": 'ш',
		"ya": 'я',
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
				// TODO(mpl): refactor
				if len(runes[i:]) > 2 {
					if cyril, ok := triple[string(runes[i:i+3])]; ok {
						trans = append(trans, cyril)
						skipTwo = true
					} else if cyril, ok := double[string(runes[i:i+2])]; ok {
						trans = append(trans, cyril)
						skipOne = true
					} else if cyril, ok := single[string(runes[i:i+1])]; ok {
						trans = append(trans, cyril)
					} else {
						log.Printf("unknown rune: %v", string(r))
						trans = append(trans, r)
					}
				} else if len(runes[i:]) > 1 {
					if cyril, ok := double[string(runes[i:i+2])]; ok {
						trans = append(trans, cyril)
						skipOne = true
					} else if cyril, ok := single[string(runes[i:i+1])]; ok {
						trans = append(trans, cyril)
					} else {
						log.Printf("unknown rune: %v", string(r))
						trans = append(trans, r)
					}
				} else {
					if cyril, ok := single[string(runes[i:i+1])]; ok {
						trans = append(trans, cyril)
					} else {
						log.Printf("unknown rune: %v", string(r))
						trans = append(trans, r)
					}
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
