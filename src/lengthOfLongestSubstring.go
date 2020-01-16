package t66y

func LengthOfLongestSubstring(s string) int {
	tmp := make(map[int32]int)
	max := 0
	maxIndex := 0
	for index, v := range s {
		_, ok := tmp[v]
		if ok && maxIndex <= tmp[v] {
			maxIndex = tmp[v] + 1
		} else { //不在数组里面
			if max < index-maxIndex+1 {
				max = index - maxIndex + 1
			}
		}
		tmp[v] = index
	}
	return max
}
