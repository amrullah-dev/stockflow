package feature

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goravel/app/services"
)

func TestReferenceNumberGeneratorPrefixesAndFormat(t *testing.T) {
	testCases := []struct {
		name       string
		prefix     string
		generateFn func() (string, error)
	}{
		{
			name:       "Stock IN",
			prefix:     services.ReferencePrefixStockIn,
			generateFn: services.GenerateStockInReference,
		},
		{
			name:       "Stock OUT",
			prefix:     services.ReferencePrefixStockOut,
			generateFn: services.GenerateStockOutReference,
		},
		{
			name:       "Warehouse Transfer",
			prefix:     services.ReferencePrefixWarehouseTransfer,
			generateFn: services.GenerateWarehouseTransferReference,
		},
	}

	today := time.Now().Format("20060102")
	pattern := regexp.MustCompile(`^(IN|OUT|TR)-\d{8}-[0-9A-Z]{6}$`)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := tc.generateFn()
			require.NoError(t, err)
			require.NotEmpty(t, ref)

			// Overall format
			require.True(t, pattern.MatchString(ref), "reference %q must match PREFIX-YYYYMMDD-RANDOM", ref)

			// Prefix
			require.True(t, strings.HasPrefix(ref, tc.prefix+"-"), "reference %q must start with prefix %s", ref, tc.prefix)

			// Date portion
			parts := strings.Split(ref, "-")
			require.Len(t, parts, 3)
			require.Equal(t, tc.prefix, parts[0])
			require.Equal(t, today, parts[1])
			require.Len(t, parts[2], 6)

			// Length constraint (fits within varchar 50)
			require.LessOrEqual(t, len(ref), 50)
		})
	}
}

func TestReferenceNumberGeneratorAtDeterministicDate(t *testing.T) {
	fixedTime := time.Date(2026, 9, 24, 15, 4, 5, 0, time.UTC)

	ref, err := services.GenerateReferenceNumberAt(services.ReferencePrefixStockIn, fixedTime)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(ref, "IN-20260924-"))

	parts := strings.Split(ref, "-")
	require.Len(t, parts, 3)
	require.Equal(t, "IN", parts[0])
	require.Equal(t, "20260924", parts[1])
	require.Len(t, parts[2], 6)
}

func TestReferenceNumberGeneratorRejectsInvalidPrefix(t *testing.T) {
	invalidPrefixes := []string{"", "INVALID", "STOCK", "PO", "SO"}
	for _, p := range invalidPrefixes {
		ref, err := services.GenerateReferenceNumber(p)
		require.Error(t, err)
		require.ErrorIs(t, err, services.ErrInvalidReferencePrefix)
		require.Empty(t, ref)
	}
}

func TestReferenceNumberGeneratorRepeatedCallsDiverge(t *testing.T) {
	const iterations = 50
	seen := make(map[string]struct{}, iterations)

	for i := 0; i < iterations; i++ {
		ref, err := services.GenerateStockInReference()
		require.NoError(t, err)

		_, exists := seen[ref]
		require.False(t, exists, "generated reference numbers must differ across repeated calls: %s", ref)
		seen[ref] = struct{}{}
	}
}
