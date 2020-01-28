package api

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

func ConvertString(data []byte) string {
	return string(data)
}

func ConvertSize(size int64) string {

	unit := ""
	s := float64(size)

	if s > 1024.0 {
		s = s / 1024
		unit = "k"
	}

	if s > 1024.0 {
		s = s / 1024
		unit = "M"
	}

	if s > 1024.0 {
		s = s / 1024
		unit = "G"
	}

	return fmt.Sprintf("%0.1f%s", s, unit)
}

func ConvertHTML(data string) template.HTML {
	return template.HTML(data)
}

func EraseBR(data string) string {
	return strings.Replace(data, "<br>", " ", -1)
}

func ConvertDate(t time.Time) string {
	if t.IsZero() {
		return "None"
	}
	jst, _ := time.LoadLocation("Asia/Tokyo")
	jt := t.In(jst)
	return jt.Format("2006/01/02 15:04:05")
}
