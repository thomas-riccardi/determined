package dethcl

import (
	"testing"

	"github.com/genelet/determined/utils"
)

func TestHclSimple(t *testing.T) {
	data1 := `
	radius = 1.234
`
	c := new(circle)
	err := UnmarshalSpec([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}

/*
	func TestHclArray(t *testing.T) {
		type datacenter struct {
			Alias  string `json:"alias" hcl:"alias,label"`
			Region string `json:"region" hcl:"region"`
		}
		type city struct {
			Google []*datacenter `json:"google" hcl:"google,block"`
		}

		type provider struct {
			Provider map[string]*city `json:"provider" hcl:"provider,block"`
		}

		data2 := `
		   	# default configuration
		   	provider "google" {
		   	#	alias  = "us"
		   	  region = "us-central1"
		   	}

		   	# alternate configuration, whose alias is "europe"
		   	provider "google" {
		   	  alias  = "europe"
		   	  region = "europe-west1"
		   	}

`

	p := new(provider)
	err := UnmarshalSpec([]byte(data2), p, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Errorf("%#v", p.Provider["google"])

}
*/
func TestHclShape1(t *testing.T) {
	data1 := `
	name = "peter shape"
	shape {
		radius = 1.234
	}
`
	g := &geo{}
	c := &circle{}
	ref := map[string]interface{}{"circle": c}
	spec, err := utils.NewStruct(
		"geo", map[string]interface{}{"Shape": "circle"})
	err = UnmarshalSpec([]byte(data1), g, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "peter shape" || g.Shape.(*circle).Radius != 1.234 {
		t.Errorf("%#v", g)
	}

	data2 := `
	name = "peter shape"
	shape {
    	sx = 5
    	sy = 6
	}
`
	g = &geo{}
	s := &square{}
	ref = map[string]interface{}{"circle": c, "square": s}
	spec, err = utils.NewStruct(
		"geo", map[string]interface{}{"Shape": "square"})
	err = UnmarshalSpec([]byte(data2), g, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "peter shape" || g.Shape.(*square).SX != 5 {
		t.Errorf("%#v", g)
	}
}

func TestHclShape2(t *testing.T) {
	data4 := `
	name = "peter drawings"
	drawings {
		sx=55
		sy=66
	}
	drawings {
		sx=7
		sy=8
	}
`
	p := &picture{}
	spec, err := utils.NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		t.Fatal(err)
	}

	g := &geo{}
	s := &square{}
	c := &circle{}
	ref := map[string]interface{}{"geo": g, "circle": c, "square": s}
	err = UnmarshalSpec([]byte(data4), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings := p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*square).SX != 55 ||
		drawings[1].(*square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*square))
		t.Errorf("%#v", drawings[1].(*square))
	}

	data5 := `
    name = "peter drawings"
    drawings "abc1" def1 {
        sx=5
        sy=6
    }
    drawings abc2 "def2" {
        sx=7
        sy=8
    }
`
	p = &picture{}
	spec, err = utils.NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"moresquare", "moresquare"}})
	if err != nil {
		t.Fatal(err)
	}
	ref["moresquare"] = &moresquare{}
	err = UnmarshalSpec([]byte(data5), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings = p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*moresquare).Morename1 != "abc1" ||
		drawings[0].(*moresquare).SX != 5 ||
		drawings[1].(*moresquare).Morename2 != "def2" ||
		drawings[1].(*moresquare).SX != 7 {
		t.Errorf("%#v", drawings[0].(*moresquare))
		t.Errorf("%#v", drawings[1].(*moresquare))
	}

	bs, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter drawings"
  drawings "abc1" "def1" {
    sx = 5
    sy = 6
  }
  drawings "abc2" "def2" {
    sx = 7
    sy = 8
  }` {
		t.Errorf("%s", bs)
	}
}

func TestHclShape3(t *testing.T) {
	data4 := `
	name = "peter drawings"
	drawings {
		sx=55
		sy=66
	}
	drawings {
		sx=7
		sy=8
	}
`
	p := &Painting{}
	err := Unmarshal([]byte(data4), p)
	if err != nil {
		t.Fatal(err)
	}
	drawings := p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*square).SX != 55 ||
		drawings[1].(*square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*square))
		t.Errorf("%#v", drawings[1].(*square))
	}
}

func TestHclShape4(t *testing.T) {
	data := `
city = "Chicago"
makers "x" {
	name = "marcus drawings"
	drawings {
		sx=11
		sy=22
	}
	drawings {
		sx=3
		sy=4
	}
}

makers "y" {
	name = "peter drawings"
	drawings {
		sx=55
		sy=66
	}
	drawings {
		sx=7
		sy=8
	}
}
`
	g := &gallery{}

	err := Unmarshal([]byte(data), g)
	if err != nil {
		t.Fatal(err)
	}
	p := g.Makers["y"]
	drawings := p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*square).SX != 55 ||
		drawings[1].(*square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*square))
		t.Errorf("%#v", drawings[1].(*square))
	}
}

func TestHash(t *testing.T) {
	data3 := `
	name = "peter shapes"
	shapes obj5 {
		sx = 5
		sy = 6
	}
	shapes obj7 {
		sx = 7
		sy = 8
	}
`
	g := &geometry{}
	spec, err := utils.NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": []string{"square", "square"}})
	//"Shapes": []string{"square", "square"}})
	ref := map[string]interface{}{"square": new(square)}
	err = UnmarshalSpec([]byte(data3), g, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	shapes := g.Shapes
	if g.Name != "peter shapes" ||
		shapes["obj5"].(*square).SX != 5 ||
		shapes["obj7"].(*square).SX != 7 {
		t.Errorf("%#v", shapes["obj5"].(*square))
		t.Errorf("%#v", shapes["obj7"].(*square))
	}
}

func TestHclOld(t *testing.T) {
	data1 := `
description = "here is detailed description"
y7 {
  many = 3
  why = "national day"
}
y10 {
  many = 4
  why = "labor day"
}
y10 {
  many = 5
  why = "holiday day"
}
y11 k6 {
  many = 6
  why = "memorial day"
}
y11 k7 {
  many = 7
  why = "new day"
}
y12 k8 {
  many = 8
  why = "christmas day"
}
y12 k9 {
  many = 9
  why = "new year day"
}
y13 = {
  str131 = "la"
  str132 = "nyc"
}
y14 = {
  str141 = 141
  str142 = 142
}
y15 = {
  str151 = true
  str152 = false
}
`
	f0 := new(frame0)
	err := UnmarshalSpec([]byte(data1), f0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f0.Y10) != 2 || len(f0.Y11) != 2 || len(f0.Y12) != 2 || len(f0.Y13) != 2 || len(f0.Y14) != 2 || len(f0.Y15) != 2 {
		t.Errorf("%#v", f0)
	}
}

func TestHclFrame(t *testing.T) {
	data1 := `
description = "here is detailed description"
x1 {
  name = "x1 shape"
  shape {
    radius = 1.111
  }
}
x2 {
  name = "x2 shape"
  shape {
    radius = 2.222
  }
}
x3 {
  name = "x3 1 shape"
  shape {
    radius = 3.111
  }
}
x3 {
  name = "x3 2 shape"
  shape {
    radius = 3.222
  }
}
x4 {
  name = "x4 1 shape"
  shape {
    radius = 4.111
  }
}
x4 {
  name = "x4 2 shape"
  shape {
    radius = 4.222
  }
}
x5 k51 {
  name = "x5 1 shape"
  shape {
    radius = 5.111
  }
}
x5 k52 {
  name = "x5 2 shape"
  shape {
    radius = 5.222
  }
}
x6 k61 {
  name = "x6 1 shape"
  shape {
    radius = 6.111
  }
}
x6 k62 {
  name = "x6 2 shape"
  shape {
    radius = 6.222
  }
}

y7 {
  many = 3
  why = "national day"
}
number = 4
what = "flags"

y10 {
  many = 4
  why = "labor day"
}
y10 {
  many = 5
  why = "holiday day"
}

y11 k6 {
  many = 6
  why = "memorial day"
}
y11 k7 {
  many = 7
  why = "new day"
}

`
	spec, err := utils.NewStruct(
		"frame", map[string]interface{}{
			"X1": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X2": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X3": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X4": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X5": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X6": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "frame": &frame{}}

	f := new(frame)
	err = UnmarshalSpec([]byte(data1), f, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if f.Description != "here is detailed description" {
		t.Errorf("description: %#v", f)
	}
	if f.Y7.Many != 3 || f.Y7.Why != "national day" {
		t.Errorf("x7: %#v", f.Y7)
	}
	if f.X8.Number != 4 || f.X8.What != "flags" {
		t.Errorf("x8: %#v", f.X8)
	}
	if len(f.Y10) != 2 || len(f.Y11) != 2 {
		t.Errorf("%#v", f.Y10)
		t.Errorf("%#v", f.Y11)
	}
}

func TestHclChild(t *testing.T) {
	data1 := `
age = 5
brand {
	toy_name = "roblox"
	price = 99.9
	geo {
		name = "peter shape"
		shape {
    		radius = 1.234
		}
	}
}
`
	spec, err := utils.NewStruct(
		"child1", map[string]interface{}{
			"Brand": [2]interface{}{
				"toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"geo", map[string]interface{}{"Shape": "circle"}}}}})
	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "toy": &toy{}}

	c := new(child)
	err = UnmarshalSpec([]byte(data1), c, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if c.Age != 5 || c.Brand.Geo.Shape.(*circle).Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}

func TestHclChildMore(t *testing.T) {
	data1 := `
age = 5
brand = {
	toy_name = "roblox"
	price = 99.9
	geo = {
		name = "peter shape"
		shape = {
    		radius = 1.234
		}
	}
}
`
	spec, err := utils.NewStruct(
		"child1", map[string]interface{}{
			"Brand": [2]interface{}{
				"toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"geo", map[string]interface{}{"Shape": "circle"}}}}})
	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "toy": &toy{}}

	c := new(child)
	err = UnmarshalSpec([]byte(data1), c, spec, ref)
	str := err.Error()
	expected := `An argument named "brand" is not expected here.`
	if str[len(str)-len(expected):] != expected {
		t.Errorf("'%s'", str[len(str)-len(expected):])
	}
}

// treat object as map. add = in front of {}
func TestHclChildMap(t *testing.T) {
	data1 := `{
age = 5
brand = {
	toy_name = "roblox"
	price = 99.9
	geo = {
		name = "peter shape"
		shape = {
    		radius = 1.234
		}
	}
}
}`

	c := map[string]interface{}{}
	err := Unmarshal([]byte(data1), &c)
	if err != nil {
		t.Fatal(err)
	}

	if c["age"].(int64) != 5 || c["brand"].(map[string]interface{})["geo"].(map[string]interface{})["shape"].(map[string]interface{})["radius"].(float64) != 1.234 {
		t.Errorf("%#v", c)
	}
}
