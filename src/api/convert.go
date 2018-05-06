package api

import (
	"html/template"
	"time"
)

func ConvertString(data []byte) string {
	return string(data)
}

func ConvertHTML(data []byte) template.HTML {
	return template.HTML(data)
}

func ConvertDate(t time.Time) string {
	if t.IsZero() {
		return "None"
	}
	jst, _ := time.LoadLocation("Asia/Tokyo")
	jt := t.In(jst)
	return jt.Format("2006/01/02 15:04")
}
