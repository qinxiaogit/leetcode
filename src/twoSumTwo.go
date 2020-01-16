package t66y

import "fmt"

func twoSum(nums []int, target int) []int {
	result :=[]int{1,3}
	for i := 0; i < len(nums); i++ {
		for j := len(nums) - 1; j > i; j-- {
			fmt.Println(i,j)
			if nums[i]+nums[j] == target {
				result[0] = i;
				result[1] = j;
				return result
			}
		}
	}
	return result
}

func twoSumTwo(nums []int,target int)[]int{
	var m = make(map[int]int)
	for i := 0; i < len(nums); i++ {
		let := target-nums[i]
		_,ok := m[let]
		if ok{
			return []int{m[let],i}
		}
		m[nums[i]] = i
	}
	return []int{0,0}
}