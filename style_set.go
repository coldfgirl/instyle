package instyle

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// keySizeMax is used to determine the maximum size of a named style.
	// This is important for optimization in that it allows for a map to be used to find styles by name.
	keySizeMax = 16

	// styleDepthMax is used to determine how many levels of nesting are allowed.
	styleDepthMax = 5
)

var (
	openingBegin = []rune("[~")
	openingClose = []rune("]")
	closing      = []rune("[/]")

	reset = []rune{'\033', '[', '0', 'm'}
)

type Styler interface {
	// Apply will parse and replace any valid style tags in the original rune array and return the result.
	//
	//
	// Style tags are applied to a maximum depth of 5 nested tags.
	Apply(original []rune) (output []rune)

	// Register will add a new named style which can be used in future calls of Apply.
	// The value expected is an ANSI escape code such as `31` for red.
	// To have a style map to multiple ANSI escape codes, separate them with a semicolon.
	//
	// Names have a maximum length of 16 characters.
	//
	// For example:
	//
	//    s := instyle.NewStyler()
	//    s.Register("error", "1;31")
	//    _ = s.Apply([]rune("[~error]Something unexpected happened"))
	Register(name string, value string) (self Styler)

	// RegisterLipGlossStyle will extract the text styling from a [lipgloss.Style] and register it under the name provided.
	// Specifically, the following will be captured if set on the style:
	//
	//  - Foreground color
	//  - Background color
	//  - Text styling of bold / faint / italic / underline / blink / strikethrough
	//
	// [lipgloss.Style]: https://github.com/charmbracelet/lipgloss
	RegisterLipGlossStyle(name string, value lipgloss.Style) (self Styler)
}

type styleSet struct {
	named map[[keySizeMax]rune][]rune
}

func NewStyler() Styler {
	s := new(styleSet)
	s.named = make(map[[keySizeMax]rune][]rune)

	s.Register("plain", "22")

	s.Register("reset", "0")
	s.Register("bold", "1")
	s.Register("faint", "2")
	s.Register("italic", "3")
	s.Register("underline", "4")
	s.Register("blink", "5")
	s.Register("strike", "6")

	s.Register("black", "30")
	s.Register("red", "31")
	s.Register("green", "32")
	s.Register("yellow", "33")
	s.Register("blue", "34")
	s.Register("magenta", "35")
	s.Register("cyan", "36")
	s.Register("white", "37")
	s.Register("default", "39")

	s.Register("bg-black", "40")
	s.Register("bg-red", "41")
	s.Register("bg-green", "42")
	s.Register("bg-yellow", "43")
	s.Register("bg-blue", "44")
	s.Register("bg-magenta", "45")
	s.Register("bg-cyan", "46")
	s.Register("bg-white", "47")
	s.Register("bg-default", "49")

	s.Register("light-black", "90")
	s.Register("light-red", "91")
	s.Register("light-green", "92")
	s.Register("light-yellow", "93")
	s.Register("light-blue", "94")
	s.Register("light-magenta", "95")
	s.Register("light-cyan", "96")
	s.Register("light-white", "97")

	s.Register("bg-light-black", "100")
	s.Register("bg-light-red", "101")
	s.Register("bg-light-green", "102")
	s.Register("bg-light-yellow", "103")
	s.Register("bg-light-blue", "104")
	s.Register("bg-light-magenta", "105")
	s.Register("bg-light-cyan", "106")
	s.Register("bg-light-white", "107")

	return s
}

func (s *styleSet) Register(name string, value string) Styler {
	parsed := [keySizeMax]rune{}
	for k, v := range name[:int(math.Min(keySizeMax, float64(len(name))))] {
		parsed[k] = v
	}

	s.named[parsed] = []rune(value)
	return s
}

func (s *styleSet) RegisterLipGlossStyle(name string, value lipgloss.Style) Styler {
	p := lipgloss.ColorProfile()

	var sequence []string

	if _, noColor := value.GetForeground().(lipgloss.NoColor); !noColor {
		sequence = append(sequence, p.FromColor(value.GetForeground()).Sequence(false))
	}

	if _, noColor := value.GetBackground().(lipgloss.NoColor); !noColor {
		sequence = append(sequence, p.FromColor(value.GetBackground()).Sequence(true))
	}

	if value.GetBold() {
		sequence = append(sequence, "1")
	}

	if value.GetFaint() {
		sequence = append(sequence, "2")
	}

	if value.GetItalic() {
		sequence = append(sequence, "3")
	}

	if value.GetUnderline() {
		sequence = append(sequence, "4")
	}

	if value.GetBlink() {
		sequence = append(sequence, "5")
	}

	if value.GetStrikethrough() {
		sequence = append(sequence, "6")
	}

	return s.Register(name, strings.Join(sequence, ";"))
}

func (s *styleSet) Apply(runes []rune) []rune {
	var (
		appliedStyleStack = [styleDepthMax][]rune{}
		ok                = false
		output            = make([]rune, 0, len(runes)*4/3+5) // Pre-allocate n * 1.33 + 5 the size of the passed runes.
	)

	output = append(output, reset...)

	for i, nest := 0, 0; i < len(runes); i++ {
		r := runes[i]

		if r == openingBegin[0] && nest < styleDepthMax {
			if appliedStyleStack[nest], i, ok = s.parseOpening(runes, i); ok {
				if nest = nest + 1; nest > 0 {
					output = append(output, appliedStyleStack[nest-1]...)
					continue
				}
			}
		}

		if r == closing[0] && nest > 0 {
			if i, ok = checkSequence(closing, runes, i); ok {
				if nest = nest - 1; nest >= 0 {
					output = append(output, reset...)
					appliedStyleStack[nest] = nil

					if i+1 == len(runes) {
						appliedStyleStack[0] = nil
					}

					for i := 0; i < len(appliedStyleStack) && i < nest; i++ {
						output = append(output, appliedStyleStack[i]...)
					}

					continue
				}
			}
		}

		output = append(output, r)
	}

	if appliedStyleStack[0] != nil {
		output = append(output, reset...)
	}

	return output
}

// parseOpening operates similarly to checkSequence but specifically for the opening of a style tag.
// When a valid style tag is found, the computed sequence of ANSI style runes is returned.
func (s *styleSet) parseOpening(runes []rune, idx int) ([]rune, int, bool) {
	after, ok := checkSequence(openingBegin, runes, idx)
	if !ok {
		return nil, idx, false
	}

	sequence := make([]rune, 0, 10)
	sequence = append(sequence, '\033', '[')

	first := true
	numeric := true

	key := [keySizeMax]rune{}

	for i, count := after+1, 0; i < len(runes); i++ {
		r := runes[i]

		if isClose := r == openingClose[0]; isClose || r == '+' {
			if count == 0 || count >= keySizeMax {
				return nil, idx, false
			}

			if !first {
				sequence = append(sequence, ';')
			}

			first = false

			if found, ok := s.named[key]; ok {
				sequence = append(sequence, found...)
			} else if numeric {
				sequence = append(sequence, key[:count]...)
			} else {
				return nil, idx, false
			}

			if isClose {
				var ok bool
				if after, ok = checkSequence(openingClose, runes, i); ok {
					break
				} else {
					return nil, idx, false
				}
			}

			count = 0
			numeric = true
			key = [keySizeMax]rune{}
			continue
		}

		if r < '0' || r > '9' {
			numeric = false
		}

		key[count] = r
		count++
	}

	return append(sequence, 'm'), after, true
}

// checkSequence will attempt to find a sequence of runes at a given index.
// If the sequence is found, the runes index at the end of the sequence is returned.
func checkSequence(sequence []rune, runes []rune, idx int) (int, bool) {
	lenRunes, lenSequence := len(runes), len(sequence)

	// Determine if the sequence would be impossible given current lengths:
	if lenRunes < lenSequence || lenRunes-idx < lenSequence {
		return idx, false
	}

	// Attempt to find the sequence:
	for i := 0; i < lenSequence; i++ {
		if sequence[i] != runes[idx+i] {
			return idx, false
		}
	}

	// Return the index after the sequence:
	return idx + lenSequence - 1, true
}
