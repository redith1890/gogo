package handlers

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
)

type UserReq struct {
	Name string
}

func ParseForm(r *http.Request, dest interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		formTag := field.Tag.Get("form")
		if formTag == "" {
			formTag = strings.ToLower(field.Name)
		}

		val := r.FormValue(formTag)
		if val == "" {
			return errors.New("missing form field: " + formTag)
		}

		if field.Type.Kind() == reflect.String {
			v.Field(i).SetString(val)
		}
	}

	return nil
}
