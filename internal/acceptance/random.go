package acceptance

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Helpers for generating random tidbits for use in identifiers to prevent
// collisions in acceptance tests.

// RandInt generates a random integer
func RandInt() int {
	// #nosec G404
	return rand.Int()
}

// RandIntRange returns a random integer between min (inclusive) and max (exclusive)
func RandIntRange(min int, max int) int {
	// #nosec G404
	return rand.Intn(max-min) + min
}

func RandTimeInt() int {
	// acctest.RantInt() returns a value of size:
	// 000000000000000000
	// YYMMddHHmmsshhRRRR

	// go format: 2006-01-02 15:04:05.00

	timeStr := strings.Replace(time.Now().Local().Format("060102150405.00"), ".", "", 1) // no way to not have a .?
	postfix := RandStringFromCharSet(4, "0123456789")

	i, err := strconv.Atoi(timeStr + postfix)
	if err != nil {
		panic(err)
	}

	return i
}

// RandStringFromCharSet generates a random string by selecting characters from
// the charset provided
func RandStringFromCharSet(strlen int, charSet string) string {
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = charSet[RandIntRange(0, len(charSet))]
	}
	return string(result)
}
