package t66y

func LongestPalindrome(s string) string {
	maxLength := 0
	maxStr := ""
	for index, _ := range s {
		mOdd := checkLongestPalindrome(s, index, index)    //奇数个
		//mEven := checkLongestPalindrome(s, index, index+1) //偶数个时

		oDdLength := mOdd*2 +1
	//	evenLength := mEven*2
		if oDdLength > maxLength {
			maxStr = s[index-mOdd : index+mOdd+1]
			maxLength = oDdLength
		}
		//if evenLength > maxLength {
		//	maxStr = s[index-mEven : index+mEven+1]
		//	maxLength = evenLength
		//}



	}
	return maxStr
}

func checkLongestPalindrome(s string, i, j int) int {
	strLen := len(s)
	m := 0
	for ; i-m > 0 && j+m < strLen; m++ {
		if s[i-m] != s[i+m] {
			return m-1
		}
	}
	return m
}
