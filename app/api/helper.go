package api

import (
	"app/datastore"
	"context"
	"os"

	"bytes"
	"fmt"
	"html/template"
)

//Public template object
type Helper struct {
	Ctx         context.Context
	ID          string
	Manage      bool
	TemplateDto interface{}
}

func (p Helper) list(id string, num int) []datastore.Page {
	//TODO 1ページ目固定
	pages, _, err := datastore.SelectChildPages(p.Ctx, id, "", num, p.Manage)
	if err != nil {
		return make([]datastore.Page, 0)
	}
	return pages
}

var privateMark = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAENklEQVR4XuVb240cMQzLtZIKkv6LSCpIKwkmgA9arSiSsuceuPzkHrYs0RTFGdy+fPvi/16m9f/58fPv99+/Xq7/pzGcfeusfGb8/vraiXmttTe8VcG5EBVsFwQZgPcq3L3RtV4FggLACo8HrbWZlijG2lvtmxZ+7avaBMWDADiF7yR7Ym8E8IpX5Y4YUQLAbuxE0nfGQAJdgSADoPbUnYU5sdVLfABA3eQk8t5rWTtQAD7bzWfAtwD47MUvMDoQXhmQF02Kn8TIghUd36n2eRMA0ChSgOxGbrV/geYAhEA4wgB08xmULuHODFVmK8aagnztKwFQAiLD0RXJ4sabZaBW5yjxM3D/AXB7F9GJucdlUxUmVA9DHeiKdlR12gCwItktsL5lerDjVcYAdIeyOcsKrn7fxdw5bwwAK6IaZWyP2gaZUbcAkHvLofFOQgwEVrz7NggyIDsmFYCdfqzOZGJc/X4yatfZD2MwskAFgDFHSY4VHVlSGSDHFN2iAahItzBmbE602nEAYvExOHrVhWwtKx7Nf4epyO/YPmAFyu/dKuPCbk3VEOYN1IlzGwOQa1MSy8xBfZ5ZougLOj/uHTEgK7hCYZZMF4MxSQG6yvl1CjiClVW5OtzpTbU4dR3zFnEEXl9vMyA+hKhJRpp3e5DITluuEuoxAJ2A5f6sCo6TAjm6Ks4pxq3zjwNwJVhNiGo0onbKa1dMZIocQcwsGAHQjSUEQEfbSoOYr2DgVUxpx6DzYDEFoDqDxXKnQ8UWNLUepoDyRqULlBF3BHHCmmvP0hZ1ikEGrGB5RDAaMdYoDk5hQDdWHRAkANgMz73p3jQCmcVFeXXTSLnAp7fCDgBM8VHLxD7tblBVd/USItPHYxAlXIlVtxY9MbomZ1J8ZKE1BhFNOy2It8ho7hbviHLWucXCIwDkVqiMTFwT20wRyqqXq6kzad8RADsF5iRdEVPAQGsqc2UDoKgxKlJ5I9QZmZ3ic7vYIrgELSYRf6Yqdt5fFcWofAKIFcNiQNV3OzfmmBg00x2wqks6BkAsRknqhPhVrOzYccQJIgorRXf0Z7b6JPsiE47/fQADYtf753nOzqtAjw9+2wDklx8sodyHCiDRue2MTakFHFGrpgDaX83gyp11PeyMVySa8eev7wOQTWTjpkK0MyGqD0AxKrY54xcywHmgcOc4U2q1BXLxuS3U1ruVAU77ZGfG2LZin7qsByd4ogXYDSgFdpqCNESNm9c9ATABwdEAJdE71yAAv8wfSyOj9fR5gWmP3Xl709hKLfADE67BmSZ5175MeTQu7c8MnRC7u4pmxqrK3QYgjqOPCIZrlduPzXXBVAOTjUelxtU0OXU2u6Ttzw26dEbvE904ynpW/BWDAlA5ts6Wsj5UEs+O0mWbUvjKQwYgJ37K4yNAYhEMgPh8rwK8DYB70Edd/w+BjJCMxE5c7wAAAABJRU5ErkJggg==`

func (p Helper) mark() template.HTML {
	src := ""
	if p.Manage {
		src = `<a href="/manage/page/` + p.ID + `"><img src="` + privateMark + `" style="position: fixed; display: block; right: 0; bottom: 0; margin-right: 40px; margin-bottom: 40px; z-index: 900;" /></a>`
	}
	return template.HTML(src)
}

func (p Helper) FuncMap() template.FuncMap {
	return template.FuncMap{
		"html":              ConvertHTML,
		"eraseBR":           EraseBR,
		"plane":             ConvertString,
		"convertDate":       ConvertDate,
		"convertDateFormat": ConvertDateFormat,
		"list":              p.list,
		"mark":              p.mark,
		"templateContent":   p.ConvertTemplate,
		"variable":          p.getVariable,
		"variableHTML":      p.getVariableHTML,
		"variableCSS":       p.getVariableCSS,
		"variableJS":        p.getVariableJS,
		"variableJSStr":     p.getVariableJSStr,
		"env":               os.Getenv,
	}
}

//Contentの変換時にテンプレートを実現する
func (p Helper) ConvertTemplate(data string) template.HTML {

	tmpl, err := template.New("").Funcs(p.FuncMap()).Parse(data)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Parse Error[%s]", err))
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(data)+500))
	err = tmpl.Execute(buf, p.TemplateDto)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Execute Error[%s]", err))
	}

	return template.HTML(buf.String())
}

func (p Helper) getVariable(key string) string {
	val, err := datastore.GetVariable(p.Ctx, key)
	if err != nil {
		return err.Error()
	}
	return val
}

func (p Helper) getVariableHTML(key string) template.HTML {
	return template.HTML(p.getVariable(key))
}

func (p Helper) getVariableCSS(key string) template.CSS {
	return template.CSS(p.getVariable(key))
}

func (p Helper) getVariableJS(key string) template.JS {
	return template.JS(p.getVariable(key))
}

func (p Helper) getVariableJSStr(key string) template.JSStr {
	return template.JSStr(p.getVariable(key))
}
