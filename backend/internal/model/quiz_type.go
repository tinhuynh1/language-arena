package model

type QuizType string

const (
	QuizTypeMeaningToWord QuizType = "meaning_to_word"
	QuizTypeWordToMeaning QuizType = "word_to_meaning"
	QuizTypeWordToIPA     QuizType = "word_to_ipa"
	QuizTypeWordToPinyin  QuizType = "word_to_pinyin"
	QuizTypeWordToTone        QuizType = "word_to_tone"
	QuizTypeDefinitionToWord  QuizType = "definition_to_word"
)
