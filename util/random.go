package util

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random int64 between min and max
func RandomInt(min, max int32) int32 {
	return min + rand.Int31n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	l := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(l)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomGuest generates a random guest name
func RandomGuestName() string {
	return RandomString(5)
}

// RandomGuestSize generates a random guest size amount
func RandomGuestSize() int32 {
	return RandomInt(0, 20)
}

// RandomGuestArrivalTime returns a random time forward in the current day/next day
func RandomGuestArrivalTime() sql.NullTime {
	current := time.Now().UTC().Truncate(time.Second)
	current.Add(time.Hour*time.Duration(RandomInt(0, 10)) +
		time.Minute*time.Duration(RandomInt(0, 60)))
	return sql.NullTime{
		Time:  current,
		Valid: true,
	}
}

// RandomTableSize returns a random table size amount
func RandomTableSize() int32 {
	return RandomInt(2, 50)
}
