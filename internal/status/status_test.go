package status

import (
	"testing"
	"time"

	"license-expiry/internal/models"
)

func TestComputeCancelled(t *testing.T) {
	c := &models.Certificate{
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		IsCancelled: true,
	}
	s, _ := Compute(c, 180)
	if s != StatusCancelled {
		t.Fatalf("got %s", s)
	}
}

func TestComputeSoon(t *testing.T) {
	c := &models.Certificate{
		ExpiryDate:  time.Now().AddDate(0, 0, 10),
		IsCancelled: false,
	}
	s, days := Compute(c, 180)
	if s != StatusSoon {
		t.Fatalf("status got %s", s)
	}
	if days != 10 {
		t.Fatalf("days got %d", days)
	}
}

func TestComputeValid(t *testing.T) {
	c := &models.Certificate{
		ExpiryDate:  time.Now().AddDate(0, 0, 200),
		IsCancelled: false,
	}
	s, _ := Compute(c, 180)
	if s != StatusValid {
		t.Fatalf("got %s", s)
	}
}
