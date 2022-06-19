//go:build unit
// +build unit

package util_test

import (
	"fmt"
	"github.com/go-rest-api-example/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleSliceFilter() {
	ages := []int{6, 17, 21, 45, 52}
	fmt.Println(util.SliceFilter(ages, func(el interface{}, i int) bool {
		age := el.(int)

		return age > 40
	}).([]int))
	// Output: [45 52]
}

func ExampleSliceContain() {
	countries := []string{"Germany", "France", "Spain"}
	fmt.Println(util.SliceContain(countries, "Germany"))
	// Output: true
}

func ExampleSliceContain_second() {
	integers := []int{3, 55, 200}
	fmt.Println(util.SliceContain(integers, 3.0))
	// Output: false
}

func ExampleSliceFind() {
	type User struct {
		ID   uint
		Name string
	}

	users := []User{
		{
			ID:   1,
			Name: "John",
		},
		{
			ID:   2,
			Name: "Jane",
		},
		{
			ID:   3,
			Name: "Richard",
		},
	}

	fmt.Println(util.SliceFind(users, func(el interface{}, index int) bool {
		u := el.(User)
		return u.ID == 3
	}).(User))
	// Output: {3 Richard}
}

func ExampleSliceFindIndex() {
	fruits := []string{"Apple", "Banana", "Orange"}

	fmt.Println(util.SliceFindIndex(fruits, func(el interface{}, i int) bool {
		fruit := el.(string)
		return fruit == "Banana"
	}))
	// Output: 1
}

func TestSliceFilter(t *testing.T) {
	type args struct {
		s  interface{}
		cb util.SlicePredicateCallback
	}

	testCases := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "SliceFilter_EmptySlice_EmptySlice",
			args: args{
				s:  []string{},
				cb: func(el interface{}, i int) bool { return false },
			},
			want: []string{},
		},
		{
			name: "SliceFilter_NilFilterFunc_SliceNotFiltered",
			args: args{
				s:  []int{1, 3, 44, 121},
				cb: nil,
			},
			want: []int{1, 3, 44, 121},
		},
		{
			name: "SliceFilter_NonEmptySliceAndNonNilFilterFunc_SliceFiltered",
			args: args{
				s: []uint{11, 15, 29, 22, 34, 51},
				cb: func(el interface{}, i int) bool {
					num := el.(uint)
					return num > 20
				},
			},
			want: []uint{29, 22, 34, 51},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := util.SliceFilter(tc.args.s, tc.args.cb)

			assert.Equal(t, tc.want, got)
		})
	}
}
