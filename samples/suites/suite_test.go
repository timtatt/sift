package samples

import (
	"crypto/rand"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupSuite() {
	log.Println("this is a log from the suite setup")

	go func() {
		tick := time.NewTicker(time.Millisecond * 50)

		for {
			select {
			case <-s.T().Context().Done():
				log.Println("stopping application")
				return
			case <-tick.C:
				log.Println("some application log", rand.Text())
			}
		}
	}()

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

	time.Sleep(time.Second * 1)
}

func (s *TestSuite) TestPanic() {
	log.Println("this test is about to panic")
	panic("something went terribly wrong")
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
