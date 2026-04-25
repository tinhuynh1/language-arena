package pinyin

import (
	"math/rand"
	"strings"
	"unicode/utf8"
)

// Tone diacritics mapping: vowel → [tone1, tone2, tone3, tone4]
var toneMap = map[rune][4]rune{
	'a': {'ā', 'á', 'ǎ', 'à'},
	'e': {'ē', 'é', 'ě', 'è'},
	'i': {'ī', 'í', 'ǐ', 'ì'},
	'o': {'ō', 'ó', 'ǒ', 'ò'},
	'u': {'ū', 'ú', 'ǔ', 'ù'},
	'ü': {'ǖ', 'ǘ', 'ǚ', 'ǜ'},
}

// Reverse mapping: toned vowel → (base vowel, tone number 1-4)
var reverseToneMap map[rune]struct {
	base rune
	tone int
}

func init() {
	reverseToneMap = make(map[rune]struct {
		base rune
		tone int
	})
	for base, tones := range toneMap {
		for i, toned := range tones {
			reverseToneMap[toned] = struct {
				base rune
				tone int
			}{base, i + 1}
		}
	}
}

// SplitSyllables splits a multi-syllable pinyin string by whitespace.
func SplitSyllables(py string) []string {
	parts := strings.Fields(py)
	if len(parts) == 0 {
		return []string{py}
	}
	return parts
}

// StripTone removes the tone diacritic from a pinyin syllable
// and returns the base syllable + tone number (1-4, 0 for neutral).
func StripTone(syllable string) (string, int) {
	var result []rune
	tone := 0

	for _, r := range syllable {
		if info, ok := reverseToneMap[r]; ok {
			result = append(result, info.base)
			tone = info.tone
		} else {
			result = append(result, r)
		}
	}

	return string(result), tone
}

// ApplyTone applies a tone number (1-4) to a pinyin syllable.
// It follows standard pinyin tone placement rules:
// 1. If there's an 'a' or 'e', it takes the tone mark.
// 2. If there's 'ou', the 'o' takes the mark.
// 3. Otherwise, the second vowel takes the mark.
func ApplyTone(base string, tone int) string {
	if tone < 1 || tone > 4 {
		return base
	}

	// First strip any existing tones
	stripped, _ := StripTone(base)
	runes := []rune(stripped)

	// Find the vowel to apply the tone to
	vowelIdx := findToneVowel(runes)
	if vowelIdx < 0 {
		return stripped
	}

	v := runes[vowelIdx]
	if tones, ok := toneMap[v]; ok {
		runes[vowelIdx] = tones[tone-1]
	}

	return string(runes)
}

// findToneVowel determines which vowel gets the tone mark.
func findToneVowel(runes []rune) int {
	// Rule 1: 'a' or 'e' always gets the mark
	for i, r := range runes {
		if r == 'a' || r == 'e' {
			return i
		}
	}

	// Rule 2: 'ou' → mark goes on 'o'
	for i, r := range runes {
		if r == 'o' && i+1 < len(runes) && runes[i+1] == 'u' {
			return i
		}
	}

	// Rule 3: second vowel gets the mark
	vowels := "iouü"
	count := 0
	for i, r := range runes {
		if strings.ContainsRune(vowels, r) {
			count++
			if count == 2 {
				return i
			}
		}
	}

	// Fallback: first vowel
	for i, r := range runes {
		if strings.ContainsRune("aiouüe", r) {
			return i
		}
	}

	return -1
}

// GenerateToneVariants generates 4 tone variants for a pinyin string.
// For single-syllable: returns the 4 tonal variants (mā, má, mǎ, mà).
// For multi-syllable: generates 3 wrong variants by shuffling tones across syllables.
func GenerateToneVariants(py string) []string {
	syllables := SplitSyllables(py)

	if len(syllables) == 1 {
		return generateSingleSyllableVariants(syllables[0])
	}
	return generateMultiSyllableVariants(syllables, py)
}

func generateSingleSyllableVariants(syllable string) []string {
	base, _ := StripTone(syllable)
	variants := make([]string, 4)
	for i := 1; i <= 4; i++ {
		variants[i-1] = ApplyTone(base, i)
	}
	return variants
}

func generateMultiSyllableVariants(syllables []string, original string) []string {
	// Extract base + tone for each syllable
	type syllableInfo struct {
		base string
		tone int
	}

	infos := make([]syllableInfo, len(syllables))
	for i, s := range syllables {
		base, tone := StripTone(s)
		infos[i] = syllableInfo{base, tone}
		if tone == 0 {
			infos[i].tone = 1 // default neutral to tone 1 for variation
		}
	}

	variants := []string{original}
	used := map[string]bool{original: true}

	for len(variants) < 4 {
		// Create a wrong variant by randomly changing tones
		parts := make([]string, len(infos))
		for i, info := range infos {
			newTone := rand.Intn(4) + 1
			// Ensure at least one syllable differs
			parts[i] = ApplyTone(info.base, newTone)
		}
		candidate := strings.Join(parts, " ")

		if !used[candidate] {
			used[candidate] = true
			variants = append(variants, candidate)
		}
	}

	return variants
}

// StripAllTones removes all tone diacritics from a full pinyin string.
func StripAllTones(py string) string {
	var result []rune
	for _, r := range py {
		if info, ok := reverseToneMap[r]; ok {
			result = append(result, info.base)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// HasTone returns true if the pinyin string contains any tone diacritics.
func HasTone(py string) bool {
	for _, r := range py {
		if _, ok := reverseToneMap[r]; ok {
			return true
		}
	}
	return false
}

// Len returns the number of runes in a string (useful for pinyin with multi-byte chars).
func Len(s string) int {
	return utf8.RuneCountInString(s)
}
