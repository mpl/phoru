// Copyright 2016 Mathieu Lonjaret

// Given a (pseudo-)phonetic russian input, phoru outputs (cyrillic) russian.
package phoru

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

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

var Verbose bool

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
		"x": 'х',
		"î": 'ы',
		"è": 'э',
	}
	double = map[string]rune{
		"yo": 'ё',
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

func Translate(r io.Reader) ([]rune, error) {
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
				if Verbose {
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
				if Verbose {
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

