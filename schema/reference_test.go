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
