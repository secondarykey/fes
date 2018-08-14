package main

import (
	"time"
	"io"
	"os"
	"fmt"
)

type Stage struct {
	Name string
	Speakers []Artist
}

type Artist struct {
	Name string
	StartTime time.Time
	EndTime   time.Time
	Color string
}

func main() {

	startTime := time.Date(2018, time.October, 20, 9, 0, 0, 0, time.Local)
	endTime   := time.Date(2018, time.October, 20, 20, 0, 0, 0, time.Local)

	stages := make([]Stage,0)
	stage1 := Stage{
		Name: "Stage1",
	}
	stage2 := Stage{
		Name: "Stage2",
	}
	stage3 := Stage{
		Name: "Stage3",
	}
	stage4 := Stage{
		Name: "Stage4",
	}

	stage1.Speakers = make([]Artist,2)
	stage1.Speakers[0] = Artist{
		Name:"CHOP STICK",
		StartTime: time.Date(2018, time.October, 20, 9, 15, 0, 0, time.Local),
		EndTime: time.Date(2018, time.October, 20, 9, 40, 0, 0, time.Local),
		Color: "pink",
	}
	stage1.Speakers[1] = Artist{
		Name:"Who",
		StartTime: time.Date(2018, time.October, 20, 10, 0, 0, 0, time.Local),
		EndTime: time.Date(2018, time.October, 20, 10, 25, 0, 0, time.Local),
		Color: "pink",
	}

	stage2.Speakers = make([]Artist,2)
	stage2.Speakers[0] = Artist{
		Name:"YABU",
		StartTime: time.Date(2018, time.October, 20, 9, 35, 0, 0, time.Local),
		EndTime: time.Date(2018, time.October, 20, 10, 10, 0, 0, time.Local),
		Color: "green",
	}
	stage2.Speakers[1] = Artist{
		Name:"誰？",
		StartTime: time.Date(2018, time.October, 20, 10, 25, 0, 0, time.Local),
		EndTime: time.Date(2018, time.October, 20, 10, 55, 0, 0, time.Local),
		Color: "green",
	}

	stages = append(stages,stage1)
	stages = append(stages,stage2)
	stages = append(stages,stage3)
	stages = append(stages,stage4)

	fp,_ := os.Create("../app/templates/manage/table/view.tmpl")
	writeTable(fp,startTime,endTime,stages)
	defer fp.Close()
}

func writeTable(w io.Writer,startTime,endTime time.Time,stages []Stage) {

	writeln(w,`{{define "title"}}
Table View
{{end}}
{{define "page_template"}}`)

	writeln(w,`<table class="mdl-shadow--2dp tt-table">`)
	writeln(w,`<thead>`)
	//時刻
	writeln(w,`<th style="min-width:25px;max-width:25px;width:25px;">` + `</th>`)

	for _,stage := range stages {
		writeln(w,`<th>`)
		writeln(w,stage.Name)
		writeln(w,`</th>`)
	}

	writeln(w,`</tr>`)
	writeln(w,`</thead>`)
	writeln(w,`<tbody>`)

	for {
		if startTime.Unix() > endTime.Unix() {
			break
		}
		writeLine(w,startTime,stages)
		startTime = startTime.Add(time.Minute * 5)
	}

	writeln(w,`</tbody>`)
	writeln(w,`</table>`)
	writeln(w,`{{ end }}`)
}

func writeLine(w io.Writer,t time.Time,stages []Stage) {

	min := t.Unix()

	hour := ""
	m := t.Minute()
	clazz := ""

	if m == 0 {
		hour = fmt.Sprintf("%d:00",t.Hour())
		clazz = "tt-hour"
	} else if m == 30 {
		hour = "30"
		clazz = "tt-half"
	}

	writeln(w,"<!--" + t.Format("15:04") + "-->")
	speakers := make([]*Artist,len(stages))
	for idx,stage := range stages {
		for _,speak := range stage.Speakers {
			s := speak.StartTime
			e := speak.EndTime
			if s.Unix() <= min && e.Unix() >= min {
				speakers[idx] = &speak
				break
			}
		}
	}

	writeln(w,`<tr class="tt-row">`)
	writeln(w,`<td class="tt-time" rowspan="2">` + hour + `</td>`)

	for _,speaker := range speakers {
		if speaker == nil {
			writeln(w, `<td class="tt-artist"></td>`)
		} else if speaker.StartTime.Unix() == min {
			writeln(w, `<td class="tt-artist"></td>`)
		}
	}
	writeln(w,`</tr>`)
	writeln(w,`<tr>`)

	for _,speaker := range speakers {
		if speaker == nil {
			writeln(w, `<td class="tt-artist ` + clazz + `"></td>`)
		} else if speaker.EndTime.Unix() == min {
			writeln(w, `<td class="tt-artist ` + clazz + `"></td>`)
		} else if speaker.StartTime.Unix() == min {
		  d := speaker.EndTime.Sub(speaker.StartTime)
		  sub := d.Minutes()

		  five := sub/5
		  span := int(five * 2)
		  // 5分コマ数 * 2
		  num := fmt.Sprintf("%d",span)

		  st := speaker.StartTime.Format("15:04")
		  et := speaker.EndTime.Format("15:04")

		  t := st + "-" + et + "<br>" + speaker.Name
		  line := `<td class="tt-artist tt-artist-name" style="background-color:` + speaker.Color + `;" rowspan="` + num + `">` + t + `</td>`
		  writeln(w,line)
		}
	}

	writeln(w,`</tr>`)
}

func writeln(w io.Writer,line string) error {
	_,err := w.Write([]byte(line + "\n"))
	return err
}

