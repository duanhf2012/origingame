package util

import (
	"testing"
)

func TestPickSlice(t *testing.T) {

	a := []int{1, 2, 3, 4, 5, 6}

	e := PickSlice(a, func(pickElement any) bool {
		val := pickElement.(int)
		switch val {
		case 0:
			return true
		}

		return false
	})
	t.Log(e)

	d := PickSlice(a, func(pickElement any) bool {
		val := pickElement.(int)
		switch val {
		case 2, 4, 6:
			return true
		}

		return false
	})
	t.Log(d)

	c := PickSlice(a, func(pickElement any) bool {
		val := pickElement.(int)
		switch val {
		case 2, 3, 4, 5, 6:
			return true
		}

		return false
	})
	t.Log(c)

	b := PickSlice(a, func(pickElement any) bool {
		val := pickElement.(int)
		switch val {
		case 1, 2, 3, 4, 5, 6:
			return false
		}

		return true
	})

	t.Log(b)

}
