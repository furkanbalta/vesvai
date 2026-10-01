package search

import (
	"sort"
	"strings"
	"unicode"
)

const (
	scoreMatch        = 16
	scoreGapStart     = -10
	scoreGapExtension = -2
	bonusBoundary     = 8
	bonusCamel        = 6
	bonusConsecutive  = 4
	bonusFirstChar    = 10
	bonusExactMatch   = 20
	bonusPrefixMatch  = 10
)

type Match struct {
	Score     int
	Positions []int
}

func Score(query, text string) int {
	if query == "" {
		return 0
	}
	m := findBestMatch(query, text)
	if m == nil {
		return -1
	}
	return m.Score
}

func Find(query string, items []string) []Match {
	if query == "" {
		result := make([]Match, len(items))
		for i := range items {
			result[i] = Match{Score: 0}
		}
		return result
	}
	q := strings.ToLower(query)
	result := make([]Match, 0, len(items))
	for _, item := range items {
		if m := findBestMatch(q, item); m != nil {
			result = append(result, *m)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

func Filter[T any](query string, items []T, getFields func(T) []string) []T {
	if query == "" {
		out := make([]T, len(items))
		copy(out, items)
		return out
	}
	q := strings.ToLower(query)

	type entry struct {
		item  T
		tier  int
		score int
	}
	entries := make([]entry, 0, len(items))
	for _, item := range items {
		primary, secondary := -1, -1
		for i, field := range getFields(item) {
			m := findBestMatch(q, field)
			if m == nil {
				continue
			}
			if i == 0 {
				if m.Score > primary {
					primary = m.Score
				}
			} else if m.Score > secondary {
				secondary = m.Score
			}
		}
		switch {
		case primary >= 0:
			entries = append(entries, entry{item: item, tier: 0, score: primary})
		case secondary >= 0:
			entries = append(entries, entry{item: item, tier: 1, score: secondary})
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].tier != entries[j].tier {
			return entries[i].tier < entries[j].tier
		}
		return entries[i].score > entries[j].score
	})

	out := make([]T, len(entries))
	for i, e := range entries {
		out[i] = e.item
	}
	return out
}

func findBestMatch(query, text string) *Match {
	qRunes := []rune(strings.ToLower(query))
	tRunesLower := []rune(strings.ToLower(text))
	tRunesOriginal := []rune(text)
	qLen := len(qRunes)
	tLen := len(tRunesLower)

	if qLen == 0 {
		return &Match{Score: 0}
	}
	if qLen > tLen {
		return nil
	}

	bestScore := -1
	bestPositions := make([]int, 0)

	var search func(qi, ti, prevPos, score int, positions []int)
	search = func(qi, ti, prevPos, score int, positions []int) {
		if qi == qLen {
			if score > bestScore {
				bestScore = score
				bestPositions = make([]int, len(positions))
				copy(bestPositions, positions)
			}
			return
		}

		for i := ti; i < tLen; i++ {
			if tRunesLower[i] == qRunes[qi] {
				newScore := score + scoreMatch
				if qi == 0 {
					newScore += bonusFirstChar
				} else {
					newScore += gapCost(prevPos, i)
					newScore += calculateCharBonus(tRunesOriginal, i)
				}
				search(qi+1, i+1, i, newScore, append(positions, i))
			}
		}
	}

	search(0, 0, -1, 0, nil)

	if bestScore < 0 {
		return nil
	}

	m := &Match{Score: bestScore, Positions: bestPositions}

	if qLen == tLen {
		m.Score += bonusExactMatch
	} else if m.Positions[0] == 0 && m.Positions[len(m.Positions)-1] < qLen {
		m.Score += bonusPrefixMatch
	}

	return m
}

func gapCost(prevPos, currPos int) int {
	gap := currPos - prevPos - 1
	if gap <= 0 {
		return 0
	}
	penalty := scoreGapStart
	if gap > 1 {
		penalty += scoreGapExtension * (gap - 1)
	}
	return penalty
}

func calculateCharBonus(text []rune, pos int) int {
	if pos == 0 {
		return bonusFirstChar
	}

	prev := text[pos-1]
	curr := text[pos]

	if unicode.IsUpper(curr) && unicode.IsLower(prev) {
		return bonusCamel
	}

	if isDelimiter(prev) {
		return bonusBoundary
	}

	if pos >= 2 && isDelimiter(text[pos-2]) {
		return bonusBoundary
	}

	if prev == curr {
		return bonusConsecutive
	}

	return 0
}

func isDelimiter(r rune) bool {
	switch r {
	case '-', '_', '/', '\\', '.', ' ', '\t':
		return true
	}
	return false
}
