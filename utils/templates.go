package utils

import (
	"html/template"
	"log"
	"net/http"
	. "gogo/strings"
	"path/filepath"
	"io/fs"
)

var ParsedTemplates = map[string]*template.Template{}

func LoadTemplates() {
	err := filepath.WalkDir("templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".html" {
			return nil
		}

		tmpl, err := template.New(d.Name()).Funcs(template.FuncMap{
			"getLit": func(key string, args ...interface{}) string {
				return GetLit("es", key, args...)
			},
		}).ParseFiles(path)
		if err != nil {
			log.Printf("Error cargando template %s: %v", d.Name(), err)
			return nil
		}

		for _, t := range tmpl.Templates() {
			ParsedTemplates[t.Name()] = t
		}
		return nil
	})

	if err != nil {
		log.Fatal("Error leyendo templates:", err)
	}
}

func RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := ParsedTemplates[name]
	if !ok {
		http.Error(w, "Template no encontrada: "+name, http.StatusInternalServerError)
		return
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error ejecutando template:", err)
	}
}


func Template(templateName string, data map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RenderTemplate(w, templateName, data)
	}
}
