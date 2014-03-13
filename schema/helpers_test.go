package schema

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
