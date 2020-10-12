package schematic

import "testing"

var initialCapTests = []struct {
	In  string
	Out string
}{
	{
		In:  "provider_id",
		Out: "ProviderID",
	},
	{
		In:  "app-identity",
		Out: "AppIdentity",
	},
	{
		In:  "uuid",
		Out: "UUID",
	},
	{
		In:  "oauth-client",
		Out: "OAuthClient",
	},
	{
		In:  "Dyno all",
		Out: "DynoAll",
	},
}

func TestInitialCap(t *testing.T) {
	for i, ict := range initialCapTests {
		depuncted := depunct(ict.In, true)
		if depuncted != ict.Out {
			t.Errorf("%d: wants %v, got %v", i, ict.Out, depuncted)
		}
	}
}

var fieldNameTests = append([]struct {
	In  string
	Out string
}{
	{
		In:  "ca_signed?",
		Out: "IsCaSigned",
	},
},
	initialCapTests...,
)

func TestFieldName(t *testing.T) {
	for _, tt := range fieldNameTests {
		t.Run(tt.In, func(t *testing.T) {
			got := fieldName(tt.In)
			if got != tt.Out {
				t.Fatalf("got %q want %q", got, tt.Out)
			}
		})
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
