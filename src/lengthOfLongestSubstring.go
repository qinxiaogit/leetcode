package t66y

func LengthOfLongestSubstring(s string) int {
  	tmp:= make(map[int32]int)
  	max := 0
  	bewttenLen := 0
  	for index,v:=range s{
  		_,ok := tmp[v]
  		if ok{
			bewttenLen = index-tmp[v]
  			if bewttenLen > max {
  				max = bewttenLen
			}
		}
  		tmp[v] = index
	}
	if max == 0{
		max = len(s)
	}
	return max
}
