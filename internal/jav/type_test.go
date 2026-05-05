package jav

import "testing"

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

func TestMetadataLanguageEnglishUsesJavDatabase(t *testing.T) {
	SetMetadataLanguage("en")
	t.Cleanup(func() { SetMetadataLanguage("zh") })

	if got := CurrentMetadataLanguage(); got != MetadataLanguageEnglish {
		t.Fatalf("CurrentMetadataLanguage() = %q, want %q", got, MetadataLanguageEnglish)
	}
	if got := PreferredProvider(); got != ProviderJavDatabase {
		t.Fatalf("PreferredProvider() = %s, want %s", got.String(), ProviderJavDatabase.String())
	}
}

func TestParseMetadataLanguageRejectsUnknownValues(t *testing.T) {
	if _, ok := ParseMetadataLanguage("system"); ok {
		t.Fatal("ParseMetadataLanguage(system) accepted an unknown language")
	}
}
