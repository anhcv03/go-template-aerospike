package util

import (
	"reflect"
	"regexp"
)

/***************************************************************************************************************/

/* Check elm is in arr */
func IsInclude[T any](arr []T, elm T) bool {
	elmVal := reflect.ValueOf(elm)

	if elmVal.Kind() == reflect.Ptr {
		elmVal = elmVal.Elem()
	}

	for idx := range arr {
		arrVal := reflect.ValueOf(arr[idx])

		if arrVal.Kind() == reflect.Ptr {
			arrVal = arrVal.Elem()
		}

		if reflect.DeepEqual(arrVal.Interface(), elmVal.Interface()) {
			return true
		}
	}

	return false
}

/* Search elms in arr */
func Search(arr []string, pattern string) []string {
	matchings := []string{}

	r, err := regexp.Compile(pattern)
	if err != nil {
		return matchings
	}

	for idx := range arr {
		if r.MatchString(arr[idx]) {
			matchings = append(matchings, arr[idx])
		}
	}

	return matchings
}

/* Check matching pattern */
func Check(str string, pattern string) bool {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	if r.MatchString(str) {
		return true
	}

	return false
}

/* Average */
func Average(arr []float64) float64 {
	arrLength := float64(len(arr))
	if arrLength == 0 {
		return 0.0
	}

	sum := 0.0
	for idx := range arr {
		sum = sum + arr[idx]
	}

	return sum / arrLength
}

/* Replace */
func Replace[T any](arr []T, old T, new T) []T {
	oldVal := reflect.ValueOf(old)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}

	for idx := range arr {
		arrVal := reflect.ValueOf(arr[idx])

		if arrVal.Kind() == reflect.Ptr {
			arrVal = arrVal.Elem()
		}

		if reflect.DeepEqual(oldVal.Interface(), arrVal.Interface()) {
			arr[idx] = new
		}
	}

	return arr
}

/* Count */
func Count[T any](arr []T, elm T) uint32 {
	var count uint32 = 0

	elmVal := reflect.ValueOf(elm)

	if elmVal.Kind() == reflect.Ptr {
		elmVal = elmVal.Elem()
	}

	for idx := range arr {
		arrVal := reflect.ValueOf(arr[idx])

		if arrVal.Kind() == reflect.Ptr {
			arrVal = arrVal.Elem()
		}

		if reflect.DeepEqual(elmVal.Interface(), arrVal.Interface()) {
			count++
		}
	}

	return count
}

/* Remove duplicated elms of arr */
func Unique[T any](arr []T) []T {
	res := []T{}

	for idx := range arr {
		if !IsInclude(res, arr[idx]) {
			res = append(res, arr[idx])
		}
	}

	return res
}

/* Loop over each element of array and no return */
func ForEach[T any, U any](arr []T, f func(int, T) U) {
	for idx := range arr {
		f(idx, arr[idx])
	}
}

/* Loop over each element of array */
func Map[T any, U any](arr []T, f func(int, T) U, onlyOne bool) []U {
	res := []U{}

	for idx := range arr {
		elm := f(idx, arr[idx])

		if !onlyOne || (onlyOne && !IsInclude(res, elm)) {
			res = append(res, elm)
		}
	}

	return res
}

/* Loop over each element and return array includes elements if func is true */
func Filter[T any](arr []T, f func(int, T) bool) []T {
	res := []T{}

	for idx := range arr {
		if f(idx, arr[idx]) {
			res = append(res, arr[idx])
		}
	}

	return res
}

/* Loop over each element and return true if any func return true */
func Any[T any](arr []T, f func(int, T) bool) bool {
	for idx := range arr {
		if f(idx, arr[idx]) {
			return true
		}
	}

	return false
}

/* Copy array and add elms*/
func CopyArray[T any](arr []T, elms ...T) []T {
	return append(append([]T{}, arr...), elms...)
}

/* Append elms to arr if elms is not in arr */
func AppendUnique[T any](arr []T, elms ...T) []T {
	res := append([]T{}, arr...)

	for idx := range elms {
		if !IsInclude(res, elms[idx]) {
			res = append(res, elms[idx])
		}
	}

	return res
}

/* Remove elms of arr */
func Remove[T any](arr []T, elms ...T) []T {
	res := []T{}

	for idx1 := range arr {
		shouldRemove := false

		for idx2 := range elms {
			arrVal := reflect.ValueOf(arr[idx1])
			elmVal := reflect.ValueOf(elms[idx2])

			if arrVal.Kind() == reflect.Ptr {
				arrVal = arrVal.Elem()
			}

			if elmVal.Kind() == reflect.Ptr {
				elmVal = elmVal.Elem()
			}

			if reflect.DeepEqual(arrVal.Interface(), elmVal.Interface()) {
				shouldRemove = true

				break
			}
		}

		if !shouldRemove {
			res = append(res, arr[idx1])
		}
	}

	return res
}

/* Min */
func Min(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	min := arr[0]
	for idx := range arr {
		if arr[idx] < min {
			min = arr[idx]
		}
	}

	return min
}

/* Max */
func Max(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	max := arr[0]
	for idx := range arr {
		if arr[idx] > max {
			max = arr[idx]
		}
	}

	return max
}

/***************************************************************************************************************/
