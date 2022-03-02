//go:build ignore
// +build ignore

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"go/format"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

const faceDownloadUrl = `https://down.qq.com/qqface/config/face_config_json_1026_hide.zip?mType=Other` //? 好像是会自动更新的

type ani struct {
	QSid      string
	StickerID string
}

type config struct {
	SystemFace []face `json:"sysface"`
	Stickers   []ani  `json:"-"`
}

type face struct {
	QSid         string `json:"QSid"`
	QDes         string `json:"QDes"`
	AniStickerId string `json:"AniStickerId"`
}

const codeTemplate = `// Code generated by message/generate.go DO NOT EDIT.

package message

var faceMap = map[int]string{
{{range .SystemFace}}	{{.QSid}}:	"{{.QDes}}",
{{end}}
}

var stickerMap = map[int]string{
{{range .Stickers}}	{{.QSid}}:	"{{.StickerID}}",
{{end}}
}
`

func main() {
	f, _ := os.OpenFile("face.go", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, 0o755)
	defer func() { _ = f.Close() }()
	resp, err := http.Get(faceDownloadUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	rsp, _ := io.ReadAll(resp.Body)
	reader, _ := zip.NewReader(bytes.NewReader(rsp), resp.ContentLength)
	file, _ := reader.Open("face_config.json")
	data, _ := io.ReadAll(file)
	faceConfig := config{}
	_ = json.Unmarshal(data, &faceConfig)
	for i, sysface := range faceConfig.SystemFace {
		faceConfig.SystemFace[i].QDes = strings.TrimPrefix(faceConfig.SystemFace[i].QDes, "/")
		if sysface.AniStickerId != "" {
			faceConfig.Stickers = append(faceConfig.Stickers, ani{
				QSid:      sysface.QSid,
				StickerID: sysface.AniStickerId,
			})
		}
	}
	tmpl, _ := template.New("template").Parse(codeTemplate)
	buffer := &bytes.Buffer{}
	_ = tmpl.Execute(buffer, &faceConfig)
	source, _ := format.Source(buffer.Bytes())
	f.Write(source)
}
