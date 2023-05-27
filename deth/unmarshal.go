package deth

import (
//"github.com/k0kubun/pp/v3"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"reflect"
	"strings"
	"unicode"
)

// Unmarshal unmarshals HCL data with interfaces determined by Determined.
//
//   - dat: Hcl data
//   - current: object as interface
//   - endpoint: Determined
//   - ref: struct map, with key being string name and value pointer to struct
//   - optional label_values: fields' values of labels
func Unmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}, label_values ...string) error {
fmt.Printf("start unmarshal .... %s\nspec ...%s\n\n", dat, endpoint.String())
	if endpoint == nil {
fmt.Printf("stop 1\n")
		return unplain(dat, current, label_values...)
	}
	objectMap := endpoint.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
fmt.Printf("stop 2\n")
		return unplain(dat, current, label_values...)
	}

	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line:1,Column:1})
	if diags.HasErrors() { return diags }
//pp.Println(file.Body)

	t := reflect.TypeOf(current).Elem()
	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	n := t.NumField()

	var newFields []reflect.StructField
	tagref := make(map[string]bool)
	for i := 0; i < n; i++ {
		field := t.Field(i)
		name := field.Name
		if unicode.IsUpper([]rune(name)[0]) && field.Tag == "" {
			return fmt.Errorf("missing tag for %s", name)
		}
		if _, ok := objectMap[name]; ok {
			two := tag2(field.Tag)
			tagref[two[0]] = true
		} else {
			newFields = append(newFields, field)
		}
	}
	if newFields != nil && len(newFields) == n {
fmt.Printf("stop 3\n")
		return unplain(dat, current, label_values...)
	}

	body := &hclsyntax.Body{
		Attributes: file.Body.(*hclsyntax.Body).Attributes,
		SrcRange: file.Body.(*hclsyntax.Body).SrcRange,	
		EndRange: file.Body.(*hclsyntax.Body).EndRange}
	blockref := make(map[string][]*hclsyntax.Block)
	for _, block := range file.Body.(*hclsyntax.Body).Blocks {
		if tagref[block.Type] {
			blockref[block.Type] = append(blockref[block.Type], block)
		} else {
			body.Blocks = append(body.Blocks, block)
		}
	}
	newType := reflect.StructOf(newFields)
	//raw := reflect.New(newType).Elem().Addr().Interface()
	raw := reflect.New(newType).Interface()
fmt.Printf("1001: %#v\n", raw)
	diags = gohcl.DecodeBody(body, nil, raw)
	if diags.HasErrors() { return diags }
	rawValue := reflect.ValueOf(raw).Elem()
fmt.Printf("33333: %#v\n", rawValue)

	m := 0
	if label_values != nil {
		m = len(label_values)
	}
	k := 0

	j := 0
	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		name := field.Name
		two := tag2(field.Tag)
		f := tmp.Elem().Field(i)
		result, ok := objectMap[name]
		if ok {
			two := tag2(field.Tag)
			blocks := blockref[two[0]]
//pp.Println(blocks) 
//x, _, _ := getBlockBytes(blocks  file)
fmt.Printf("AAAA %s=>%#v\n", name, objectMap)
fmt.Printf("BBB %s\n", result.String())
			if x := result.GetListStruct(); x != nil {
fmt.Printf("x %#v\n", x.GetListFields())
				nextListStructs := x.GetListFields()
				n := len(nextListStructs)
				if n == 0 {
					return fmt.Errorf("missing list struct for %s", name)
				}

				var fSlice, fMap reflect.Value
				if fieldType.Kind()==reflect.Map {
					fMap = reflect.MakeMapWithSize(fieldType, n)
				} else {
					fSlice = reflect.MakeSlice(fieldType, n, n)
				}
				for k := 0; k < n; k++ {
					nextStruct := nextListStructs[k]
					trial := ref[nextStruct.ClassName]
fmt.Printf("1111 %#v\n", ref)
fmt.Printf("2222 bbbbbb %s, %s=>%v\n", nextStruct.String(), nextStruct.ClassName, trial)
					if trial == nil {
						return fmt.Errorf("ref not found for %s", name)
					}
					trial = clone(trial)
					s, labels, err := getBlockBytes(blocks[k], file)
					if err != nil {
						return err
					}
					err = Unmarshal(s, trial, nextStruct, ref, labels...)
					if err != nil {
						return err
					}
					if fieldType.Kind()==reflect.Map {
						fMap.SetMapIndex(reflect.ValueOf(labels[0]), reflect.ValueOf(trial))
					} else {
						fSlice.Index(k).Set(reflect.ValueOf(trial))
					}
				}
				if fieldType.Kind()==reflect.Map {
					f.Set(fMap)
				} else {
					f.Set(fSlice)
				}
			} else if x := result.GetSingleStruct(); x != nil {
				trial := ref[x.ClassName]
				if trial == nil {
					return fmt.Errorf("class ref not found for %s", x.ClassName)
				}
fmt.Printf("3111 bbbbbb %#v\n", trial)
				trial = clone(trial)
				s, labels, err := getBlockBytes(blocks[0], file)
				if err != nil {
					return err
				}
				err = Unmarshal(s, trial, x, ref, labels...)
				if err != nil {
					return err
				}
				if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(trial))
				} else {
					f.Set(reflect.ValueOf(trial).Elem())
				}
			}
		} else if unicode.IsUpper([]rune(name)[0]) {
			if strings.ToLower(two[1]) == "label" && k<m {
				f.Set(reflect.ValueOf(label_values[k]))
				k++
			} else {
				rawField := rawValue.Field(j)
				j++
				f.Set(rawField)
			}
		}
	}

	oriValue.Set(tmp)

	return nil
}

func debugBody(x *hclsyntax.Body, file *hcl.File) {
	fmt.Printf("700 Body %v\n", x)
	for k, v := range x.Attributes {
		c, d := v.Expr.Value(nil)
		fmt.Printf("701 Attr %s => %s => %#v\n", k, v.Name, v.Expr)
		fmt.Printf("702 ctyValue %#v => %s => %#v\n", c, c.GoString(), d)
	}
   	for _, block := range x.Blocks {
		fmt.Printf("703 block %#v\n", block)
				rng1 := block.OpenBraceRange
				rng2 := block.CloseBraceRange
				bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
fmt.Printf("801 %s\n", bs)
	}
	fmt.Printf("704 range start %#v\n", x.SrcRange.Start)
	fmt.Printf("705 range end %#v\n", x.SrcRange.End)
	fmt.Printf("706 filename %#v\n", x.SrcRange.Filename)
	fmt.Printf("707 range %#v\n", x.SrcRange.String())
}

func getBodyBytes(body *hclsyntax.Body, file *hcl.File) ([]byte, []string, error) {
	if body == nil {
		return nil, nil, fmt.Errorf("body not found")
	}
	rng1 := body.SrcRange
	rng2 := body.EndRange
	bs := file.Bytes[rng1.Start.Byte:rng2.Start.Byte]
fmt.Printf("801 %s\n", bs)
	return bs, nil, nil
}
func getBlockBytes(block *hclsyntax.Block, file *hcl.File) ([]byte, []string, error) {
	if block == nil {
		return nil, nil, fmt.Errorf("block not found")
	}
	rng1 := block.OpenBraceRange
	rng2 := block.CloseBraceRange
	bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
fmt.Printf("802 %s\n", bs)
	return bs, block.Labels, nil
}
