package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	// ReferencePrefixStockIn is the prefix for Stock IN transaction reference numbers.
	ReferencePrefixStockIn = "IN"
	// ReferencePrefixStockOut is the prefix for Stock OUT transaction reference numbers.
	ReferencePrefixStockOut = "OUT"
	// ReferencePrefixWarehouseTransfer is the prefix for Warehouse Transfer transaction reference numbers.
	ReferencePrefixWarehouseTransfer = "TR"

	referenceRandomLength = 6
	referenceCharset      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	// ErrInvalidReferencePrefix indicates an unrecognized transaction reference prefix.
	ErrInvalidReferencePrefix = errors.New("invalid reference number prefix")
)

// GenerateReferenceNumber generates a reference number with the given prefix and current date.
//
// FORMAT: PREFIX-YYYYMMDD-RANDOM (e.g. IN-20260924-7X9K2L).
//
// Uniqueness and Collision Handling:
// The random portion uses crypto/rand to generate 6 uppercase alphanumeric characters
// (~2.17 billion possible combinations per prefix per day).
// While the collision probability is statistically negligible under normal workloads,
// memory-only generation cannot mathematically guarantee zero collisions across distributed workers.
// The database UNIQUE(reference_number) constraint serves as the final barrier.
// Calling service layers must handle unique constraint collisions (e.g. retry generation)
// when persisting transactions.
func GenerateReferenceNumber(prefix string) (string, error) {
	return GenerateReferenceNumberAt(prefix, time.Now())
}

// GenerateReferenceNumberAt generates a reference number using a specific date.
func GenerateReferenceNumberAt(prefix string, t time.Time) (string, error) {
	if prefix != ReferencePrefixStockIn &&
		prefix != ReferencePrefixStockOut &&
		prefix != ReferencePrefixWarehouseTransfer {
		return "", fmt.Errorf("%w: %q", ErrInvalidReferencePrefix, prefix)
	}

	randomPart, err := cryptoRandomString(referenceRandomLength)
	if err != nil {
		return "", fmt.Errorf("generate reference number: %w", err)
	}

	datePart := t.Format("20060102")
	return fmt.Sprintf("%s-%s-%s", prefix, datePart, randomPart), nil
}

// GenerateStockInReference generates an IN-YYYYMMDD-RANDOM reference number.
func GenerateStockInReference() (string, error) {
	return GenerateReferenceNumber(ReferencePrefixStockIn)
}

// GenerateStockOutReference generates an OUT-YYYYMMDD-RANDOM reference number.
func GenerateStockOutReference() (string, error) {
	return GenerateReferenceNumber(ReferencePrefixStockOut)
}

// GenerateWarehouseTransferReference generates a TR-YYYYMMDD-RANDOM reference number.
func GenerateWarehouseTransferReference() (string, error) {
	return GenerateReferenceNumber(ReferencePrefixWarehouseTransfer)
}

func cryptoRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	charsetLen := big.NewInt(int64(len(referenceCharset)))

	for i := range bytes {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		bytes[i] = referenceCharset[num.Int64()]
	}

	return string(bytes), nil
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "1062") ||
		strings.Contains(msg, "unique constraint")
}
