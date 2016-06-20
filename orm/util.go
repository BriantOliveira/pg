package orm

import (
	"reflect"

	"gopkg.in/pg.v4/types"
)

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func indirectNew(v reflect.Value, set bool) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if set {
				v.Set(reflect.New(v.Type().Elem()))
			} else {
				v = reflect.New(v.Type().Elem())
			}

		}
		v = v.Elem()
	}
	return v
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, x := range index {
		if t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		t = t.Field(x).Type
	}
	return indirectType(t)
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for i, x := range index {
		if i > 0 {
			v = indirectNew(v, true)
		}
		v = v.Field(x)
	}
	return v
}

func columns(table types.Q, prefix string, fields []*Field) []byte {
	var b []byte
	for i, f := range fields {
		if len(table) > 0 {
			b = append(b, table...)
			b = append(b, '.')
		}
		b = types.AppendField(b, prefix+f.SQLName, 1)
		if i != len(fields)-1 {
			b = append(b, ", "...)
		}
	}
	return b
}

func values(v reflect.Value, index []int, fields []*Field) []byte {
	var b []byte
	walk(v, index, func(v reflect.Value) {
		b = append(b, '(')
		for i, field := range fields {
			b = field.AppendValue(b, v, 1)
			if i != len(fields)-1 {
				b = append(b, ", "...)
			}
		}
		b = append(b, "), "...)
	})
	if len(b) > 0 {
		b = b[:len(b)-2] // trim ", "
	}
	return b
}

func walk(v reflect.Value, index []int, fn func(reflect.Value)) {
	v = reflect.Indirect(v)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			visitStruct(v.Index(i), index, fn)
		}
	} else {
		visitStruct(v, index, fn)
	}
}

func visitStruct(strct reflect.Value, index []int, fn func(reflect.Value)) {
	if len(index) > 0 {
		strct = strct.Field(index[0])
		walk(strct, index[1:], fn)
	} else {
		fn(strct)
	}
}

func appendColumnAndValue(b []byte, v reflect.Value, table *Table, fields []*Field) []byte {
	for i, f := range fields {
		b = append(b, table.Alias...)
		b = append(b, '.')
		b = append(b, f.ColName...)
		b = append(b, " = "...)
		b = f.AppendValue(b, v, 1)
		if i != len(fields)-1 {
			b = append(b, " AND "...)
		}
	}
	return b
}

func modelId(b []byte, v reflect.Value, fields []*Field) []byte {
	for _, f := range fields {
		b = f.AppendValue(b, v, 0)
		b = append(b, ',')
	}
	return b
}

func modelIdMap(b []byte, m map[string]string, prefix string, fields []*Field) []byte {
	for _, f := range fields {
		b = append(b, m[prefix+f.SQLName]...)
		b = append(b, ',')
	}
	return b
}

func dstValues(root reflect.Value, path []int, fields []*Field) map[string][]reflect.Value {
	mp := make(map[string][]reflect.Value)
	var id []byte
	walk(root, path[:len(path)-1], func(v reflect.Value) {
		id = modelId(id[:0], v, fields)
		mp[string(id)] = append(mp[string(id)], v.Field(path[len(path)-1]))
	})
	return mp
}

func appendSep(b []byte, sep string) []byte {
	if len(b) > 0 {
		b = append(b, sep...)
	}
	return b
}
