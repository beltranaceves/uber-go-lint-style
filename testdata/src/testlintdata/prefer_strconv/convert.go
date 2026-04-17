package prefer_strconv

import (
	"fmt"
	"strconv"
)

func badSprint(i int) {
	_ = fmt.Sprint(i) // want "prefer strconv functions for primitive-to-string conversions instead of fmt.Sprint"
}

func badSprintf(i int) {
	_ = fmt.Sprintf("%d", i) // want "prefer strconv functions for primitive-to-string conversions instead of fmt.Sprintf"
}

func goodItoa(i int) {
	_ = strconv.Itoa(i)
}

func goodFormatInt(i int64) {
	_ = strconv.FormatInt(i, 10)
}
