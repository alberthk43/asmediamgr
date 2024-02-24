package ytsmoviedir

import (
	"fmt"

	"golang.org/x/text/language"
)

func getLanguageFromFileName(name string) (languageStr string, err error) {
	// try match special case
	if lang, ok := specials[name]; ok {
		return lang, nil
	}
	// try match with std pkg
	tag, index, confidence := langMatcher.Match(language.Make(name))
	if confidence <= language.Low && index == 0 {
		return "", fmt.Errorf("no language found")
	}
	return tag.String(), nil
}

var (
	specials = map[string]string{
		"English.eng":        language.English.String(),
		"English":            language.English.String(),
		"Simplified.chi":     language.SimplifiedChinese.String(),
		"Traditional.chi":    language.TraditionalChinese.String(),
		"Brazilian.por":      language.BrazilianPortuguese.String(),
		"Latin American.spa": language.LatinAmericanSpanish.String(),
		"SDH.eng.HI":         language.English.String() + ".sdh",
		"繁體.chi":             language.TraditionalChinese.String(),
	}

	langMatcher = language.NewMatcher([]language.Tag{
		language.Amharic,
		language.Arabic,
		language.ModernStandardArabic,
		language.Azerbaijani,
		language.Bulgarian,
		language.Bengali,
		language.Catalan,
		language.Czech,
		language.Danish,
		language.German,
		language.Greek,
		language.English,
		language.AmericanEnglish,
		language.BritishEnglish,
		language.Spanish,
		language.EuropeanSpanish,
		language.LatinAmericanSpanish,
		language.Estonian,
		language.Persian,
		language.Finnish,
		language.Filipino,
		language.French,
		language.CanadianFrench,
		language.Gujarati,
		language.Hebrew,
		language.Hindi,
		language.Croatian,
		language.Hungarian,
		language.Armenian,
		language.Indonesian,
		language.Icelandic,
		language.Italian,
		language.Japanese,
		language.Georgian,
		language.Kazakh,
		language.Khmer,
		language.Kannada,
		language.Korean,
		language.Kirghiz,
		language.Lao,
		language.Lithuanian,
		language.Latvian,
		language.Macedonian,
		language.Malayalam,
		language.Mongolian,
		language.Marathi,
		language.Malay,
		language.Burmese,
		language.Nepali,
		language.Dutch,
		language.Norwegian,
		language.Punjabi,
		language.Polish,
		language.Portuguese,
		language.BrazilianPortuguese,
		language.EuropeanPortuguese,
		language.Romanian,
		language.Russian,
		language.Sinhala,
		language.Slovak,
		language.Slovenian,
		language.Albanian,
		language.Serbian,
		language.SerbianLatin,
		language.Swedish,
		language.Swahili,
		language.Tamil,
		language.Telugu,
		language.Thai,
		language.Turkish,
		language.Ukrainian,
		language.Urdu,
		language.Uzbek,
		language.Vietnamese,
		language.Chinese,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.Zulu,
	})
)
