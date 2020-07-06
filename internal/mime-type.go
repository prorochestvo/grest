package internal

import (
	"encoding/json"
	"encoding/xml"
)

var JSON = MimeType{Format: "application/json", Marshal: json.Marshal, Unmarshal: json.Unmarshal}
var XML = MimeType{Format: "application/xml", Marshal: xml.Marshal, Unmarshal: xml.Unmarshal}

type marshal func(v interface{}) ([]byte, error)

type unmarshal func(data []byte, v interface{}) error

type MimeType struct {
	Format    string
	Marshal   marshal
	Unmarshal unmarshal
}
