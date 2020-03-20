package fileinput

import (
	"bufio"
	"regexp"
)

// NewLineStartSplitFunc creates a bufio.SplitFunc that splits an incoming stream into
// tokens that start with a match to the regex pattern provided
func NewLineStartSplitFunc(re *regexp.Regexp) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		firstLoc := re.FindIndex(data)
		if firstLoc == nil {
			// TODO what to do with starting lines that don't match?
			return 0, nil, nil // read more data and try again.
		}
		firstMatchStart := firstLoc[0]
		firstMatchEnd := firstLoc[1]

		secondLocOffset := firstMatchEnd + 1
		secondLoc := re.FindIndex(data[secondLocOffset:])
		if secondLoc == nil {
			if atEOF {
				return len(data), data[firstMatchStart:], nil // return the rest of the file and advance to end
			}
			return 0, nil, nil // read more data and try again
		}
		secondMatchStart := secondLoc[0] + secondLocOffset

		advance = secondMatchStart                     // start scanning at the beginning of the second match
		token = data[firstMatchStart:secondMatchStart] // the token begins at the first match, and ends at the beginning of the second match
		err = nil
		return
	}
}

// NewLineEndSplitFunc creates a bufio.SplitFunc that splits an incoming stream into
// tokens that end with a match to the regex pattern provided
func NewLineEndSplitFunc(re *regexp.Regexp) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		loc := re.FindIndex(data)
		if loc == nil {
			return 0, nil, nil // read more data and try again
		}

		// If the match goes up to the end of the current buffer, do another
		// read until we can capture the entire match
		// TODO figure out how to test this
		if loc[1] == len(data)-1 && !atEOF {
			return 0, nil, nil
		}

		advance = loc[1]
		token = data[:loc[1]]
		err = nil
		return
	}
}