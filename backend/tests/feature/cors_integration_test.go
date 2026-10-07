package feature

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/tests"
)

func TestAPICorsPreflightAllowsFrontend(t *testing.T) {
	response, err := (&tests.TestCase{}).Http(t).
		WithHeaders(map[string]string{
			"Origin":                         "http://localhost:5173",
			"Access-Control-Request-Method":  "POST",
			"Access-Control-Request-Headers": "authorization,content-type",
		}).
		Options("/api/login")
	require.NoError(t, err)
	response.AssertNoContent()
	response.AssertHeader("Access-Control-Allow-Origin", "http://localhost:5173")
	response.AssertHeader("Access-Control-Allow-Methods", "POST")
	response.AssertHeader("Access-Control-Allow-Headers", "authorization,content-type")
}
