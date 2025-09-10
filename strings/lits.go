package strings

import "fmt"

var LangsName = []string{
	"es",
	"en",
}

var Langs = make(map[string]map[string]string)

func RegisterLang(lang string, lits map[string]string) {
	Langs[lang] = lits
}

func GetLit(lang, key string, args ...interface{}) string {
	langMap, ok := Langs[lang]
	if !ok {
		langMap = Langs["es"]
	}

	lit, ok := langMap[key]
	if !ok {
		if esMap, ok := Langs["es"]; ok {
			lit = esMap[key]
		}

		if lit == "" {
			return fmt.Sprintf("[MISSING: %s]", key)
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(lit, args...)
	}
	return lit
}
