package samples

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupSuite() {
	log.Println("this is a log from the suite setup")
}

func (s *TestSuite) SetupTest() {
	log.Println("this is a log from the test setup")
}

func (s *TestSuite) TearDownTest() {
	log.Println("this is a log from the test teardown")
}

func (s *TestSuite) TearDownSuite() {
	log.Println("this is a log from the suite teardown")
}

func (s *TestSuite) TestExample() {
	log.Println("this is a log from the test example")
	s.Equal(1, 1, "they should be equal")
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
