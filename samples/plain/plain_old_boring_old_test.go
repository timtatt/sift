package samples

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAFailingUnitTest(t *testing.T) {
	t.Log("this is a log from the plain old boring old test")

	t.Error("this is an error from the plain old boring old test")

	t.Fail() // Mark the test as failed
}

func TestAPassingUnitTest(t *testing.T) {
	t.Log("this is a log from the plain old boring old test")
}

func TestAFailingTestifyTest(t *testing.T) {
	assert.Equal(t, 1, 2, "they should be equal")
}

func TestWithADelay(t *testing.T) {
	t.Log("this is a log from the plain old boring old test with a delay")

	// Simulate a delay
	time.Sleep(2 * time.Second)
}

func TestSkippedTest(t *testing.T) {
	t.Skip("this test is not required")

}
