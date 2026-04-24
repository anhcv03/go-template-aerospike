package util

import "reflect"

/***************************************************************************************************************/

/* Check is equal */
func IsEqual[T any, U any](elm1 T, elm2 U) bool {
	elm1Val := reflect.ValueOf(elm1)
	elm2Val := reflect.ValueOf(elm2)

	if elm1Val.Kind() == reflect.Ptr {
		elm1Val = elm1Val.Elem()
	}

	if elm2Val.Kind() == reflect.Ptr {
		elm2Val = elm2Val.Elem()
	}

	if reflect.DeepEqual(elm1Val.Interface(), elm2Val.Interface()) {
		return true
	}

	return false
}

/* Compare */
func Compare(num1 float64, num2 float64) int {
	if num1 > num2 {
		return 1
	} else if num1 < num2 {
		return -1
	} else {
		return 0
	}
}

/***************************************************************************************************************/
