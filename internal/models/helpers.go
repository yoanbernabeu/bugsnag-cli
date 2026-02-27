package models

import "fmt"

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func ftoa(f float64) string {
	return fmt.Sprintf("%.4f", f)
}
