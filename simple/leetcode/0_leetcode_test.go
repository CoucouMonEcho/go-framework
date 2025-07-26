package leetcode

import (
	"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	res := wordsAbbreviation([]string{"like", "god", "internal", "me", "internet", "interval", "int_en_sion", "face", "intrusion"})
	//res := wordsAbbreviation([]string{"internal", "internet", "interval", "int_en_sion", "intrusion"})
	//res := wordsAbbreviation([]string{"like", "l_ke", "like_k", "like_kk"})
	//res := wordsAbbreviation([]string{"aa", "aaa"})
	fmt.Println(res)
	var i rune = 48
	rn := i + 0
	fmt.Println(string(rn))
}
