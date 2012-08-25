package main

import "strings"

var Preamble string

func init() {
	PreambleLines := []string{
		"",
	        "                                                       __",
		"  ____           ____ ___          __  __         ____/ /",
		" / __ \\         / __ `__ \\        / / / /        / __  / ",
		"/ /_/ /        / / / / / /       / /_/ /        / /_/ /  ",
		"\\____/        /_/ /_/ /_/        \\__,_/         \\__,_/   ",
		"",
		"&dim;(thanks patorjk.com/software/taag/)&;",
		""}
	Preamble = strings.Join(PreambleLines, "\r\n")
}