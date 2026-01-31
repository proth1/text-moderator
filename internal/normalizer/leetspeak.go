package normalizer

// defaultLeetspeak returns a mapping of common leetspeak substitutions.
func defaultLeetspeak() map[rune]rune {
	return map[rune]rune{
		'0': 'o',
		'1': 'l',
		'3': 'e',
		'4': 'a',
		'5': 's',
		'7': 't',
		'@': 'a',
		'$': 's',
		'!': 'i',
	}
}
