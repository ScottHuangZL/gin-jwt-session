package model

import (
	"fmt"
	"time"
)

//FormatAsDate return a yyyy-mm-dd date
func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}
