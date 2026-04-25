package pinyin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripTone(t *testing.T) {
	tests := []struct {
		input    string
		wantBase string
		wantTone int
	}{
		{"mā", "ma", 1},
		{"má", "ma", 2},
		{"mǎ", "ma", 3},
		{"mà", "ma", 4},
		{"ma", "ma", 0},
		{"xué", "xue", 2},
		{"zhuāng", "zhuang", 1},
		{"nǚ", "nü", 3},
		{"lǜ", "lü", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			base, tone := StripTone(tt.input)
			assert.Equal(t, tt.wantBase, base)
			assert.Equal(t, tt.wantTone, tone)
		})
	}
}

func TestApplyTone(t *testing.T) {
	tests := []struct {
		base string
		tone int
		want string
	}{
		{"ma", 1, "mā"},
		{"ma", 2, "má"},
		{"ma", 3, "mǎ"},
		{"ma", 4, "mà"},
		{"ma", 0, "ma"},   // neutral
		{"ma", 5, "ma"},   // out of range
		{"xue", 2, "xué"},
		{"zhuang", 1, "zhuāng"},
		{"gui", 4, "guì"},
		{"liu", 2, "liú"},
		{"nü", 3, "nǚ"},
	}

	for _, tt := range tests {
		t.Run(tt.base+"_"+string(rune('0'+tt.tone)), func(t *testing.T) {
			got := ApplyTone(tt.base, tt.tone)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStripTone_ApplyTone_Roundtrip(t *testing.T) {
	originals := []string{"mā", "xué", "hǎo", "shì", "nǚ", "zhuāng"}
	for _, orig := range originals {
		base, tone := StripTone(orig)
		rebuilt := ApplyTone(base, tone)
		assert.Equal(t, orig, rebuilt, "roundtrip failed for %s", orig)
	}
}

func TestSplitSyllables(t *testing.T) {
	assert.Equal(t, []string{"xué", "xí"}, SplitSyllables("xué xí"))
	assert.Equal(t, []string{"mā"}, SplitSyllables("mā"))
	assert.Equal(t, []string{"nǐ", "hǎo"}, SplitSyllables("nǐ hǎo"))
	assert.Equal(t, []string{""}, SplitSyllables(""))
}

func TestGenerateToneVariants_SingleSyllable(t *testing.T) {
	variants := GenerateToneVariants("mā")
	assert.Len(t, variants, 4)
	assert.Contains(t, variants, "mā")
	assert.Contains(t, variants, "má")
	assert.Contains(t, variants, "mǎ")
	assert.Contains(t, variants, "mà")
}

func TestGenerateToneVariants_MultiSyllable(t *testing.T) {
	variants := GenerateToneVariants("xué xí")
	assert.Len(t, variants, 4)
	assert.Contains(t, variants, "xué xí") // correct one is always included

	// All variants should be unique
	unique := make(map[string]bool)
	for _, v := range variants {
		unique[v] = true
	}
	assert.Len(t, unique, 4)
}

func TestStripAllTones(t *testing.T) {
	assert.Equal(t, "xue xi", StripAllTones("xué xí"))
	assert.Equal(t, "ni hao", StripAllTones("nǐ hǎo"))
	assert.Equal(t, "ma", StripAllTones("ma")) // no tone
}

func TestHasTone(t *testing.T) {
	assert.True(t, HasTone("mā"))
	assert.True(t, HasTone("xué xí"))
	assert.False(t, HasTone("ma"))
	assert.False(t, HasTone("hello"))
}
