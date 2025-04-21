package load

import (
	"bufio"
	"bytes"
	"math/rand/v2"
	"strings"

	"github.com/klauspost/compress/zstd"
)

func joiner(maxLeft int, maxRight int) func(trg [][2]int) {
	used := map[[2]int]bool{}
	return func(trg [][2]int) {
		// fill with random pairs
		for i := range len(trg) {
			found := true
			for found {
				trg[i][0] = rand.Int() % maxLeft
				trg[i][1] = rand.Int() % maxRight
				_, found = used[trg[i]]
			}
			used[trg[i]] = true
		}
	}
}

func ComparePair(a, b [2]int) int {
	if a[0] == b[0] {
		return a[1] - b[1]
	}
	return a[0] - b[0]
}

// fillWords will fill a string slice with random dash-delimited word triplets
func fillWords(trg []string) {
	parts := make([]string, 3)
	for i := range len(trg) {
		for j := range len(parts) {
			parts[j] = words[rand.Int()%len(words)]
		}
		trg[i] = strings.Join(parts, "-")
	}
}

func readWords() []string {
	f, err := wordsfs.Open("data/words.zst")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	zf, err := zstd.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer zf.Close()
	scn := bufio.NewScanner(zf)
	var words []string
	for scn.Scan() {
		words = append(words, scn.Text())
	}
	return words
}

// ScanQueries is a bufio.SplitFunc that splits on semicolons. This is used to
// split indiviual queries from a file of queries. Note that the results are not
// trimmed and this does not exclude empty queries.
func ScanQueries(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, ';'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
