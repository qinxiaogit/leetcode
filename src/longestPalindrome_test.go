package t66y

import "testing"

func TestLongestPalindrome(t *testing.T) {
	if LongestPalindrome("aaaaaa") !="aaaaaa"{
		t.Error("fail")
	}
}
//"babad"
func TestLongestPalindromeBab(t *testing.T) {
	if LongestPalindrome("babad") !="bab"{
		t.Error("fail")
	}
}