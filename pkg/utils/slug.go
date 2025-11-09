package utils

import (
	"regexp"
	"strings"
)

// GenerateSlug creates a URL-friendly slug from variable number of strings
// Usage: GenerateSlug("string1", "string2", ...) or GenerateSlug(singleString)
// Example: GenerateSlug("Футбольное", "Поле") returns "futbolnoe-pole"
func GenerateSlug(texts ...string) string {
	// Concatenate all input strings with spaces
	concatenated := strings.Join(texts, " ")

	// Convert to lowercase
	slug := strings.ToLower(concatenated)

	// Transliterate common Cyrillic characters to Latin
	cyrillicToLatin := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z",
		'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r",
		'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "E", 'Ж': "Zh", 'З': "Z",
		'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N", 'О': "O", 'П': "P", 'Р': "R",
		'С': "S", 'Т': "T", 'У': "U", 'Ф': "F", 'Х': "H", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var transliterated strings.Builder
	for _, r := range slug {
		if replacement, exists := cyrillicToLatin[r]; exists {
			transliterated.WriteString(replacement)
		} else {
			transliterated.WriteRune(r)
		}
	}
	slug = transliterated.String()

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9\\-]+")
	slug = reg.ReplaceAllString(slug, "_")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	return slug
}
