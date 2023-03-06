package utils

import (
	"math/rand"
	"time"
)

func RandomNumberGenerator(min int, max int) int {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Intn(max-min)
}

func DeleteItemInArray(array []interface{}, item interface{}) []interface{} {

	found := false
	var i int
	for k, value := range array {
		if equal(value, item) {
			i = k
			break
		}
	}

	if found {
		data := append(array[:i], array[i+1:]...)
		return data
	}

	return nil
}

func DeleteItemByIndexInArray(array []int, i int) []int {

	data := append(array[:i], array[i+1:]...)

	return data
}

func equal(a interface{}, b interface{}) bool {

	// int
	ai, ok1 := a.(int)
	bi, ok2 := b.(int)
	if ok1 && ok2 && ai == bi {
		return true
	}

	return false
}
