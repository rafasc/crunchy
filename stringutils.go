/*
 * crunchy - find common flaws in passwords
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package crunchy

import (
	"bufio"
	"hash"
	"os"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// countUniqueChars returns the amount of unique runes in a string
func countUniqueChars(s string) int {
	m := make(map[rune]struct{})

	for _, c := range s {
		c = unicode.ToLower(c)
		if _, ok := m[c]; !ok {
			m[c] = struct{}{}
		}
	}

	return len(m)
}

// countSystematicChars returns how many runes in a string are part of a sequence ('abcdef', '654321')
func countSystematicChars(s string) int {
	var x int
	rs := []rune(s)

	for i, c := range rs {
		if i == 0 {
			continue
		}
		if c == rs[i-1]+1 || c == rs[i-1]-1 {
			x++
		}
	}

	return x
}

// reverse returns the reversed form of a string
func reverse(s string) string {
	var rs []rune
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-size]

		rs = append(rs, r)
	}

	return string(rs)
}

// normalize returns the trimmed and lowercase version of a string
func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func hashsum(hashers []hash.Hash, words <-chan []byte, hashes chan<- []byte, wg *sync.WaitGroup) {
	//make a copy of each hasher
	hh := make([]hash.Hash, len(hashers))
	for i, hasher := range hashers {
		hh[i] = reflect.New(reflect.TypeOf(hasher).Elem()).Interface().(hash.Hash)
	}

	for word := range words {
		for _, h := range hh {
			h.Reset()
			_, _ = h.Write(word)
			hashes <- h.Sum(nil)
		}
	}
	wg.Done()
}

// returns the lines of a file as a channel
func linesFromFiles(filenames []string) <-chan string {
	out := make(chan string, 500)
	go func() {
		for _, filename := range filenames {
			file, err := os.Open(filename)
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				out <- scanner.Text()
			}

		}
		close(out)
	}()
	return out
}

func lineCount(filename string) (linecount uint, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linecount++
	}
	return linecount, scanner.Err()
}
