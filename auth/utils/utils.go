package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomNumberString(length int) string {

	var result strings.Builder
	for n := 0; n < length; n++ {
		result.WriteString(strconv.Itoa(rand.Intn(9)))
	}
	return result.String()

}
