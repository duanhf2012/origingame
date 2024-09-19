package gm

import "strconv"

const (
	OK = "OK"
)

func getArgInt(arg []string, argIndex int) int {
	if argIndex >= len(arg) {
		return 0
	}

	value, err := strconv.Atoi(arg[argIndex])
	if err != nil {
		return 0
	}

	return value
}

func getArgString(arg []string, argIndex int) string {
	if argIndex >= len(arg) {
		return ""
	}

	return arg[argIndex]
}
