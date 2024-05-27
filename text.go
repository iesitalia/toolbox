package toolbox

import (
	"bytes"
	"encoding/json"
	"github.com/gosimple/unidecode"
	"regexp"
	"strings"
)

// regexpFileNonAuthorizedChars is a regular expression that matches any character that is not a letter, digit, hyphen, underscore, period, or slash.
// regexpSlugNonAuthorizedChars is a regular expression that matches any character that is not a letter, digit, hyphen, or underscore.
// regexpSlugMultipleDashes is a regular expression that matches one or more consecutive hyphens.
var (
	regexpFileNonAuthorizedChars = regexp.MustCompile(`[^a-zA-Z0-9-_.]`)
	regexpSlugNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")
	regexpSlugMultipleDashes     = regexp.MustCompile("-+")
)

// slugSub is a map that holds the mapping of specific characters to their corresponding substitutions in a slug.
var slugSub = map[rune]string{
	'"':  "",
	'\'': "",
	'’':  "",
	'‒':  "-", // figure dash
	'–':  "-", // en dash
	'_':  "-", // en dash
	'—':  "-", // em dash
	'―':  "-", // horizontal bar
	'β':  "v",
	'Β':  "V",
	'η':  "i",
	'Η':  "I",
	'ή':  "i",
	'Ή':  "I",
	'ι':  "i",
	'Ι':  "I",
	'ί':  "i",
	'Ί':  "I",
	'ϊ':  "i",
	'Ϊ':  "I",
	'ΐ':  "i",
	'ξ':  "x",
	'Ξ':  "X",
	'υ':  "y",
	'Υ':  "Y",
	'ύ':  "y",
	'Ύ':  "Y",
	'ϋ':  "y",
	'Ϋ':  "Y",
	'ΰ':  "y",
	'φ':  "f",
	'Φ':  "F",
	'χ':  "ch",
	'Χ':  "Ch",
	'ω':  "o",
	'Ω':  "O",
	'ώ':  "o",
	'Ώ':  "O",
	'á':  "a",
	'Á':  "A",
	'é':  "e",
	'É':  "E",
	'í':  "i",
	'Í':  "I",
	'ó':  "o",
	'Ó':  "O",
	'ő':  "o",
	'Ő':  "O",
	'ú':  "u",
	'Ú':  "U",
	'ű':  "u",
	'Ű':  "U",
	'ә':  "a",
	'ғ':  "g",
	'қ':  "q",
	'ң':  "n",
	'ө':  "o",
	'ұ':  "u",
	'Ә':  "A",
	'Ғ':  "G",
	'Қ':  "Q",
	'Ң':  "N",
	'Ө':  "O",
	'Ұ':  "U",
	'æ':  "ae",
	'ø':  "oe",
	'å':  "aa",
	'Æ':  "Ae",
	'Ø':  "Oe",
	'Å':  "Aa",
	'Ă':  "A",
	'ă':  "a",
	'Â':  "A",
	'â':  "a",
	'Î':  "I",
	'î':  "i",
	'Ș':  "S",
	'ș':  "s",
	'Ț':  "T",
	'ț':  "t",
	'Đ':  "DZ",
	'đ':  "dz",
	'ş':  "s",
	'Ş':  "S",
	'ü':  "u",
	'Ü':  "U",
	'ö':  "o",
	'Ö':  "O",
	'İ':  "I",
	'ı':  "i",
	'ğ':  "g",
	'Ğ':  "G",
	'ç':  "c",
	'Ç':  "C",
	'А':  "A",
	'Б':  "B",
	'В':  "V",
	'Г':  "G",
	'Д':  "D",
	'Е':  "E",
	'Ж':  "Zh",
	'З':  "Z",
	'И':  "I",
	'Й':  "Y",
	'К':  "K",
	'Л':  "L",
	'М':  "M",
	'Н':  "N",
	'О':  "O",
	'П':  "P",
	'Р':  "R",
	'С':  "S",
	'Т':  "T",
	'У':  "U",
	'Ф':  "F",
	'Х':  "H",
	'Ц':  "Ts",
	'Ч':  "Ch",
	'Ш':  "Sh",
	'Щ':  "Sh",
	'Ъ':  "A",
	'Ь':  "Y",
	'Ю':  "Yu",
	'Я':  "Ya",
	'а':  "a",
	'б':  "b",
	'в':  "v",
	'г':  "g",
	'д':  "d",
	'е':  "e",
	'ж':  "zh",
	'з':  "z",
	'и':  "i",
	'й':  "y",
	'к':  "k",
	'л':  "l",
	'м':  "m",
	'н':  "n",
	'о':  "o",
	'п':  "p",
	'р':  "r",
	'с':  "s",
	'т':  "t",
	'у':  "u",
	'ф':  "f",
	'х':  "h",
	'ц':  "ts",
	'ч':  "ch",
	'ш':  "sh",
	'щ':  "sht",
	'ъ':  "a",
	'ь':  "y",
	'ю':  "yu",
	'я':  "ya",
}

// NormalizeFilename takes an input string and returns a normalized filename.
// The normalization process involves converting non-ASCII characters to their closest ASCII equivalents,
// converting all characters to lowercase, removing or substituting certain characters, and trimming leading/trailing dashes and underscores.
func NormalizeFilename(in string) string {
	// Process all non ASCII symbols
	slug := unidecode.Unidecode(in)
	slug = strings.ToLower(slug)

	slug = SubstituteRune(slug, slugSub)
	slug = regexpFileNonAuthorizedChars.ReplaceAllString(slug, "-")
	slug = regexpSlugMultipleDashes.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-_")

	return slug
}

// Slugify takes an input string and converts it into a slug format.
// It replaces non-ASCII symbols, converts to lowercase, substitutes specific runes,
// replaces non-authorized characters with dashes, removes consecutive dashes,
// and trims dashes and underscores from both ends.
func Slugify(in string) string {
	// Process all non ASCII symbols
	slug := unidecode.Unidecode(in)
	slug = strings.ToLower(slug)

	slug = SubstituteRune(slug, slugSub)
	slug = regexpSlugNonAuthorizedChars.ReplaceAllString(slug, "-")
	slug = regexpSlugMultipleDashes.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-_")

	return slug
}

// SubstituteRune substitutes characters in a string with corresponding values from a provided map.
//
// The function expects a string `s` and a map `sub` of type `map[rune]string` as parameters.
// It iterates over each character in the string and checks if it exists in the map.
// If a match is found, the corresponding value is appended to a buffer.
// If a match is not found, the character is directly appended to the buffer.
// Finally, the function returns the contents of the buffer as a string.
//
// Example usage:
//
//	sub := map[rune]string{
//	    'a': "A",
//	    'b': "B",
//	}
//	result := SubstituteRune("abc", sub) // result will be "ABc"
//
// Note: This function is typically used to substitute non-ASCII characters with ASCII representations.
// It can be utilized for tasks like normalizing filenames or generating slugs.
func SubstituteRune(s string, sub map[rune]string) string {
	var buf bytes.Buffer
	for _, c := range s {
		if d, ok := sub[c]; ok {
			buf.WriteString(d)
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

// ToJSON converts a value to its JSON representation.
func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// StripHTMLTags removes all HTML tags from the given string and returns the result.
//
// Example usage:
//
//	s := "<p>This is some <em>example</em> text.</p>"
//	result := StripHTMLTags(s) // result: "This is some example text."
func StripHTMLTags(s string) string {
	r := regexp.MustCompile(`<.*?>`)
	return r.ReplaceAllString(s, "")
}

// TruncateText truncates a given string to a maximum length if it exceeds the maximum length.
func TruncateText(s string, max int) string {
	if len(s) > max {
		r := 0
		for i := range s {
			r++
			if r > max {
				return s[:i]
			}
		}
	}
	return s
}
