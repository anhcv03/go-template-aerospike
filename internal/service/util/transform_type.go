package util

import (
	"strconv"
	"strings"
)

/***************************************************************************************************************/

/* Convert String to Float array */
func StringToFloats(stringComma string, s string) ([]float64, error) {
	stringArr := strings.Split(stringComma, s)

	floatArr := []float64{}
	for idx := range stringArr {
		floatArrElm, err := strconv.ParseFloat(strings.TrimSpace(stringArr[idx]), 64)
		if err != nil {
			return nil, err
		}

		floatArr = append(floatArr, floatArrElm)
	}

	return floatArr, nil
}

/***************************************************************************************************************/
