package jav

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMetadataLanguageDefaultsToChinese(t *testing.T) {
	SetMetadataLanguage("")
	t.Cleanup(func() { SetMetadataLanguage("zh") })

	if got := CurrentMetadataLanguage(); got != MetadataLanguageChinese {
		t.Fatalf("CurrentMetadataLanguage() = %q, want %q", got, MetadataLanguageChinese)
	}
	if got := PreferredProvider(); got != ProviderJavBus {
		t.Fatalf("PreferredProvider() = %s, want %s", got.String(), ProviderJavBus.String())
	}
}

func TestMetadataLanguageEnglishUsesThePornDB(t *testing.T) {
	SetMetadataLanguage("en")
	t.Cleanup(func() { SetMetadataLanguage("zh") })

	if got := CurrentMetadataLanguage(); got != MetadataLanguageEnglish {
		t.Fatalf("CurrentMetadataLanguage() = %q, want %q", got, MetadataLanguageEnglish)
	}
	if got := PreferredProvider(); got != ProviderJavDatabase {
		t.Fatalf("PreferredProvider() = %s, want %s", got.String(), ProviderJavDatabase.String())
	}
}

func TestProviderLanguageHelpers(t *testing.T) {
	tests := []struct {
		provider  Provider
		language  MetadataLanguage
		isEnglish bool
	}{
		{ProviderJavBus, MetadataLanguageChinese, false},
		{ProviderJavDB, MetadataLanguageChinese, false},
		{ProviderAvmoo, MetadataLanguageChinese, false},
		{ProviderAvsox, MetadataLanguageChinese, false},
		{ProviderJavMenu, MetadataLanguageChinese, false},
		{ProviderJavDatabase, MetadataLanguageEnglish, true},
		{ProviderThePornDB, MetadataLanguageEnglish, true},
		{ProviderUser, MetadataLanguageChinese, false},
		{ProviderUnknown, MetadataLanguageChinese, false},
	}

	for _, tt := range tests {
		if got := ProviderMetadataLanguage(tt.provider); got != tt.language {
			t.Fatalf("ProviderMetadataLanguage(%s) = %q, want %q", tt.provider.String(), got, tt.language)
		}
		if got := ProviderIsEnglish(tt.provider); got != tt.isEnglish {
			t.Fatalf("ProviderIsEnglish(%s) = %t, want %t", tt.provider.String(), got, tt.isEnglish)
		}
	}
}

func TestLookupProvidersByProviderIncludesMetadataProviders(t *testing.T) {
	tests := []Provider{
		ProviderJavBus,
		ProviderJavDatabase,
		ProviderJavDB,
		ProviderAvmoo,
		ProviderAvsox,
		ProviderJavMenu,
		ProviderThePornDB,
		ProviderJavModel,
	}

	for _, provider := range tests {
		got, ok := lookupProvidersByProvider[provider]
		if !ok {
			t.Fatalf("lookup provider missing for %s", provider.String())
		}
		if got == nil {
			t.Fatalf("lookup provider for %s is nil", provider.String())
		}
	}
}

func TestJavInfoOmitsUncensoredWhenUnset(t *testing.T) {
	data, err := json.Marshal(JavInfo{Code: "ABC-001", Provider: ProviderAvmoo})
	if err != nil {
		t.Fatalf("marshal jav info: %v", err)
	}
	if strings.Contains(string(data), "IsUncensored") {
		t.Fatalf("unexpected IsUncensored field: %s", data)
	}
}

func TestParseMetadataLanguageRejectsUnknownValues(t *testing.T) {
	if _, ok := ParseMetadataLanguage("system"); ok {
		t.Fatal("ParseMetadataLanguage(system) accepted an unknown language")
	}
}
