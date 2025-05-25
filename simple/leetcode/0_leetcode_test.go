package leetcode

import (
	"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	res := wordsAbbreviation([]string{"like", "god", "internal", "me", "internet", "interval", "intension", "face", "intrusion"})
	//res := wordsAbbreviation([]string{"internal", "internet", "interval", "intension", "intrusion"})
	//res := wordsAbbreviation([]string{"like", "loke", "likk", "likkk"})
	//res := wordsAbbreviation([]string{"aa", "aaa"})
	fmt.Println(res)
	var i rune = 48
	rn := i + 0
	fmt.Println(string(rn))
}
