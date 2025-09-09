package samples

import (
	"log"
	"testing"
)

func TestBabushkaTests(t *testing.T) {
	t.Run("big doll", func(t *testing.T) {
		t.Run("medium doll", func(t *testing.T) {
			log.Println("middle doll 1")
			t.Run("little doll", func(t *testing.T) {
				log.Println("little doll")
			})
			t.Run("little doll", func(t *testing.T) {
				log.Println("little doll")
			})
		})
		t.Run("medium/doll", func(t *testing.T) {
			log.Println("middle doll 2")
			t.Run("little doll", func(t *testing.T) {
				log.Println("little doll")
			})
		})

	})
}
