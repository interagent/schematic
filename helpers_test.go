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
		Comment: `日本では、フランスはファッションや美術、料理など、文化的に高い評価を受ける国として有名であり、毎年多数の日本人観光客が高級ブランドや美術館巡り、グルメツアーなどを目的にフランスを訪れている。また、音楽、美術、料理などを学ぶためにフランスに渡る日本人も多く、在仏日本人は3万5千人に及ぶ`,
		Commented: `// 日本では、フランスはファッションや美術、料理など、文化的に高い評価を受ける国として有名であり、毎年多数の日本人観光客が高級ブランドや美術館巡
// り、グルメツアーなどを目的にフランスを訪れている。また、音楽、美術、料理などを学ぶためにフランスに渡る日本人も多く、在仏日本人は3万5千人に
// 及ぶ
`,
	},
	{
		Comment: `a value that Heroku will use to sign all webhook notification requests (the signature is included in the request’s "Heroku-Webhook-Hmac-SHA256" header)`,
		Commented: `// a value that Heroku will use to sign all webhook notification
// requests (the signature is included in the request’s
// "Heroku-Webhook-Hmac-SHA256" header)
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
