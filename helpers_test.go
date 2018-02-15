package schematic

import "testing"

var initialCapTests = []struct {
	Ident     string
	Depuncted string
}{
	{
		Ident:     "provider_id",
		Depuncted: "ProviderID",
	},
	{
		Ident:     "app-identity",
		Depuncted: "AppIdentity",
	},
	{
		Ident:     "uuid",
		Depuncted: "UUID",
	},
	{
		Ident:     "oauth-client",
		Depuncted: "OAuthClient",
	},
	{
		Ident:     "Dyno all",
		Depuncted: "DynoAll",
	},
}

func TestInitialCap(t *testing.T) {
	for i, ict := range initialCapTests {
		depuncted := depunct(ict.Ident, true)
		if depuncted != ict.Depuncted {
			t.Errorf("%d: wants %v, got %v", i, ict.Depuncted, depuncted)
		}
	}
}

var asCommentTests = []struct {
	Comment   string
	Commented string
}{
	{
		Comment: `This is a multi-line
comment, it contains multiple
lines.`,
		Commented: `// This is a multi-line
// comment, it contains multiple
// lines.
`,
	},
	{
		Comment: `This is a fairly long comment line, it should be over seventy characters.`,
		Commented: `// This is a fairly long comment line, it should be over seventy
// characters.
`,
	},
	{
		Comment: `散りぬべき時知りてこそ世の中の花も花なれ人も人なれ`,
		Commented: `// 散りぬべき時知りてこそ世の中の花も花なれ人も人なれ
`,
	},
}

func TestAsComment(t *testing.T) {
	for i, act := range asCommentTests {
		c := asComment(act.Comment)
		if c != act.Commented {
			t.Errorf("%d: wants %v+, got %v+", i, act.Commented, c)
		}
	}
}
