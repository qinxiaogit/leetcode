package t66y

import (
	"math"
)

/**
  * 给定两个大小为 m 和 n 的有序数组 nums1 和 nums2。

  * 请你找出这两个有序数组的中位数，并且要求算法的时间复杂度为 O(log(m + n))。

  * 你可以假设 nums1 和 nums2 不会同时为空。
 */
func FindMedianSortedArrays(nums1 []int, nums2 []int) float64 {
	lenNums1 := len(nums1)
	lenNums2 := len(nums2)
	l := (lenNums1+lenNums2+1)/2
	r := (lenNums1+lenNums2+2)/2
	ret1 := getKth(nums1,0,nums2,0,l)
	ret2 := getKth(nums1,0,nums2,0,r)
	return float64(ret1+ret2)/float64(2)
}

func getKth(nums1 []int,start1 int,nums2 []int,start2 int,k int)int{
	//序列1的开始位置大于序列1的长度
	if start1 > len(nums1)-1 {
		return nums2[start2+k-1]
	}
	//序列2的开始位置大于序列2的长度
	if start2 > len(nums2) -1{
		return nums1[start1+k-1]
	}

	//寻找序列中第一个元素
	if k == 1{
		return int(math.Min(float64(nums1[start1]),float64(nums2[start2])))
	}
	// 分别在两个数组中查找第k/2个元素，
	//若存在（即数组没有越界），
	//标记为找到的值；
	//若不存在，标记为整数最大值
	var nums1Mid ,nums2Mid int
	if start1+k/2-1<len(nums1) {
		nums1Mid = nums1[start1+k/2-1]
	}else{
		nums1Mid = 	math.MaxInt32
	}
	if start2+k/2-1<len(nums2) {
		nums2Mid = nums1[start2+k/2-1]
	}else{
		nums2Mid =  math.MaxInt32
	}
	if nums1Mid < nums2Mid {
		return getKth(nums1,start1+k/2,nums2,start2,k-k/2)
	}
	return getKth(nums1,start1,nums2,start2+k/2,k-k/2)
}
