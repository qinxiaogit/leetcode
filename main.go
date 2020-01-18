package main

import (
	"fmt"
	t66y "go-es/src"
)

func main() {
	//fmt.Println(t66y.LengthOfLongestSubstring("dvdf"))
	//fmt.Println(t66y.LengthOfLongestSubstring("cdd"))
	//fmt.Println(t66y.LengthOfLongestSubstring("abcabcbb"))
	//fmt.Println(t66y.LengthOfLongestSubstring("bbbbb"))
	//fmt.Println(t66y.LengthOfLongestSubstring("pwwkew"))
	//fmt.Println(t66y.LengthOfLongestSubstring("bbbbb"))

	//ret := t66y.FindMedianSortedArrays([]int{1,3},[]int{2})
	ret := t66y.LongestPalindrome("babad")
	fmt.Println(ret)
}
