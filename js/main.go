/*
Copyright 2017 Mathieu Lonjaret
*/

// Entry point to call the gopherjs generated code from the "regular" js code.
package main

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mpl/phoru"
)

func main() {
	js.Global.Set("gophoru", map[string]interface{}{
		"Run":    Run,
	})
}

func Run(input string, onError func(errMsg string)) string {
	out, err := phoru.Translate(strings.NewReader(input))
	if err != nil {
		println(err)
		onError(fmt.Sprintf("%v", err))
		return ""
	}
	return fmt.Sprintf("%v", string(out))
}
