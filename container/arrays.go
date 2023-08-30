package container

import "reflect"

func Contains[T any](array []T, elemToFind T) bool {
	return ContainsMatch(array, func(elem T) bool {
		return reflect.DeepEqual(elem, elemToFind)
	})
}

func ContainsMatch[T any](array []T, matchToFind func(elem T) bool) bool {
	for _, elem := range array {
		if matchToFind(elem) {
			return true
		}
	}
	return false
}

func Eq[T any](arr1, arr2 []T) bool {
	return reflect.DeepEqual(arr1, arr2)
}
