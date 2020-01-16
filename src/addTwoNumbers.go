package t66y

import "fmt"


// Definition for singly-linked list.
type ListNode struct {
	Val int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	l1q:= []int{}
	l2q:= []int{}
	for l1 !=nil||l2 != nil  {
		if l1 !=nil {
			l1q = append(l1q, l1.Val)
			l1 = l1.Next
		}else{
			l1q = append(l1q,0)
			l1 = nil
		}
		if l2 != nil{
			l2q = append(l2q,l2.Val)
			l2 = l2.Next
		}else{
			l2q = append(l2q,0)
			l2  = nil
		}
	}
	head :=&ListNode{Val:0,Next:nil}
	current :=  head
	seq := 0
	sum :=0
	for i:=0;i<len(l2q)  ;i++  {
		tmp :=ListNode{Val:0,Next:nil}
		sum = l1q[i]+l2q[i]+seq
		fmt.Println("-:",sum)
		if (sum)/10 == 1{
			seq = 1
			tmp.Val =(sum)%10
		}else{
			tmp.Val =sum
			if seq == 1{
				seq = 0
			}
		}
		tmp.Next =nil
		current.Next = &tmp;
		current = current.Next
	}
	if seq == 1{
		tmp :=ListNode{Val:0,Next:nil}
		tmp.Next =nil
		tmp.Val = seq
		current.Next = &tmp;
	}
	return head.Next
}
func printList(node *ListNode){
	for true {
		fmt.Println(node.Val)
		node = node.Next
		if node == nil{
			break
		}
	}
}
func getList(nums []int)*ListNode{
	head :=&ListNode{Val:0,Next:nil}
	current :=  head
	for i:=0;i<len(nums);i++  {
		tmp :=ListNode{Val:0,Next:nil}
		tmp.Val  = nums[i]
		tmp.Next = nil
		current.Next = &tmp
		current = current.Next
	}
	return head.Next
}
