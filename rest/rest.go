package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v2"
)

const (
	FieldTypeString = FieldType("string")
	FieldTypeInt    = FieldType("int")
	FieldTypeFloat  = FieldType("float")
)

type FieldType string

func (f FieldType) Value(val interface{}) (interface{}, error) {
	if f == FieldTypeString {
		if val, ok := val.(string); ok {
			return val, nil
		}
		if val, ok := val.(fmt.Stringer); ok {
			return val.String(), nil
		}
	}
	if f == FieldTypeInt {
		if isInt(val) {
			return val, nil
		}
	}
	if f == FieldTypeFloat {
		if isFloat(val) {
			return val, nil
		}
	}
	return nil, fmt.Errorf("bad value for type %s", f)
}
func (f FieldType) Zero() interface{} {
	if f == FieldTypeString {
		return ""
	}
	if f == FieldTypeInt {
		return 0
	}
	if f == FieldTypeFloat {
		return 0.0
	}
	panic(fmt.Errorf("unsupported FieldType %s", f))
}

func isInt(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return true
	default:
		return false
	}
}
func isFloat(val interface{}) bool {
	switch val.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

type Field struct {
	Name     string    `yaml:"name"`
	Type     FieldType `yaml:"type"`
	Nullable bool      `yaml:"nullable"`
}
type Resource struct {
	Name   string   `yaml:"name"`
	Fields []*Field `yaml:"fields"`

	db *sqlx.DB
}

func NewResource(db *sql.DB) (*Resource, error) {
	res := &Resource{
		db: sqlx.NewDb(db, "sqlite3"),
	}
	return res, nil
}

func LoadResource(db *sql.DB, b []byte) (*Resource, error) {
	res, err := NewResource(db)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (res *Resource) Route(router *mux.Router) {
	router.HandleFunc("", res.index).Methods("GET").Name(res.Name + ".index")
	router.HandleFunc("/{id}", res.show).Methods("GET").Name(res.Name + ".show")
	router.HandleFunc("/{id}", res.create).Methods("POST").Name(res.Name + ".create")
	router.HandleFunc("/{id}", res.update).Methods("PUT").Name(res.Name + ".update")
	router.HandleFunc("/{id}", res.delete).Methods("DELETE").Name(res.Name + ".delete")
}

func (res *Resource) cleanMap(m map[string]interface{}) map[string]interface{} {
	var val interface{}
	var err error
	newMap := map[string]interface{}{}
	for _, field := range res.Fields {
		val, err = field.Type.Value(m[field.Name])
		if err != nil {
			if field.Nullable {
				val = nil
			} else {
				val = field.Type.Zero()
			}
		}
		newMap[field.Name] = val
	}
	return newMap
}

type RestError struct {
	Message string `json:"message"`
}

func errorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&RestError{
		Message: err.Error(),
	})
}

func (res *Resource) index(w http.ResponseWriter, r *http.Request) {
	rows, err := res.db.QueryxContext(r.Context(), "select * from "+res.Name)
	if err != nil {
		errorResponse(w, err)
		return
	}

	list := []map[string]interface{}{}
	m := map[string]interface{}{}
	for rows.Next() {
		err = rows.MapScan(m)
		m = res.cleanMap(m)
		list = append(list, m)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

func (res *Resource) show(w http.ResponseWriter, r *http.Request) {
	rows, err := res.db.QueryxContext(r.Context(), "select * from "+res.Name+" where id=? limit 1", "1a849fe7-621c-4942-a674-0a8c3d9a191c")
	if err != nil {
		errorResponse(w, err)
		return
	}

	if !rows.Next() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nil)
		return
	}
	m := map[string]interface{}{}
	err = rows.MapScan(m)

	m = res.cleanMap(m)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

func (res *Resource) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"route": "create"})
}

func (res *Resource) update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"route": "update"})
}

func (res *Resource) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"route": "delete"})
}
