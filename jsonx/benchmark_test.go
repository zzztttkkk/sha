package jsonx

import (
	stdlib "encoding/json"
	"testing"
)

var data = []byte(`{
    "glossary": {
        "title": "example glossary",
		"GlossDiv": {
            "title": "S",
			"GlossList": {
                "GlossEntry": {
                    "ID": "SGML",
					"SortAs": "SGML",
					"GlossTerm": "Standard Generalized Markup Language",
					"Acronym": "SGML",
					"Abbrev": "ISO 8879:1986",
					"GlossDef": {
                        "para": "A meta-markup language, used to create markup languages such as DocBook.",
						"GlossSeeAlso": ["GML", "XML"]
                    },
					"GlossSee": "markup"
                }
            }
        }
    }
}`)

type Data struct {
	Glossary struct {
		Title    string `json:"title"`
		GlossDiv struct {
			Title     string `json:"title"`
			GlossList struct {
				GlossEntry struct {
					ID        string `json:"ID"`
					SortAs    string `json:"SortAs"`
					GlossTerm string `json:"GlossTerm"`
					Acronym   string `json:"Acronym"`
					GlossDef  struct {
						Para         string   `json:"para"`
						GlossSeeAlso []string `json:"GlossSeeAlso"`
					} `json:"GlossDef"`
				} `json:"GlossEntry"`
			} `json:"GlossList"`
		} `json:"GlossDiv"`
	} `json:"glossary"`
}

func BenchmarkUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dist Data
		e := Unmarshal(data, &dist)
		if e != nil {
			b.Fail()
		}
	}
}

func BenchmarkStdUnmarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dist Data
		e := stdlib.Unmarshal(data, &dist)
		if e != nil {
			b.Fail()
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dist Data
		_, e := Marshal(&dist)
		if e != nil {
			b.Fail()
		}
	}
}

func BenchmarkStdMarshal(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var dist Data
		_, e := stdlib.Marshal(&dist)
		if e != nil {
			b.Fail()
		}
	}
}
