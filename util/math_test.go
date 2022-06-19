//go:build unit
// +build unit

package util_test

import (
	"fmt"
	"github.com/go-rest-api-example/util"
)

func ExampleRoundTo() {
	v := 3.141592
	n := 4 // 4 decimal places.

	fmt.Println(util.RoundTo(v, n))
	// Output: 3.1416
}
