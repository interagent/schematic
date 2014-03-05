package schema

import (
	"reflect"
	"testing"
)

var resolveTests = []struct {
	Ref      string
	Schema   *Schema
	Resolved *Schema
}{
	{
		Ref: "#/definitions/uuid",
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"uuid": &Schema{
					Title: "Identifier",
				},
			},
		},
		Resolved: &Schema{
			Title: "Identifier",
		},
	},
	{
		Ref: "#/definitions/struct/definitions/uuid",
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"struct": &Schema{
					Definitions: map[string]*Schema{
						"uuid": &Schema{
							Title: "Identifier",
						},
					},
				},
			},
		},
		Resolved: &Schema{
			Title: "Identifier",
		},
	},
}

func TestReferenceResolve(t *testing.T) {
	for i, rt := range resolveTests {
		ref := NewReference(rt.Ref)
		rsl := ref.Resolve(rt.Schema)
		if !reflect.DeepEqual(rsl, rt.Resolved) {
			t.Errorf("%i: resolved schema don't match, got %v, wants %v", i, rsl, rt.Resolved)
		}
	}
}

var hrefTests = []struct {
	HRef     string
	Schema   *Schema
	Resolved map[string]*Schema
}{
	{
		HRef: "/edit/{(%23%2Fdefinitions%2Fstruct%2Fdefinitions%2Fuuid)}",
		Schema: &Schema{
			Definitions: map[string]*Schema{
				"struct": &Schema{
					Definitions: map[string]*Schema{
						"uuid": &Schema{
							Title: "Identifier",
						},
					},
				},
			},
		},
		Resolved: map[string]*Schema{
			"uuid": &Schema{
				Title: "Identifier",
			},
		},
	},
}

func TestHREfResolve(t *testing.T) {
	for i, ht := range hrefTests {
		href := NewHRef(ht.HRef)
		rsl := href.Resolve(ht.Schema)
		if !reflect.DeepEqual(rsl, ht.Resolved) {
			t.Errorf("%i: resolved schemas don't match, got %v, wants %v", i, rsl, ht.Resolved)
		}
	}
}
