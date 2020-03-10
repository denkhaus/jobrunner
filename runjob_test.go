package jobrunner

import (
	"testing"
	"time"
)

var testCount int

func testFunc() error {
	testCount++
	return nil
}

func TestDebounce(t *testing.T) {
	testCount = 0
	Start()

	job := New("Debounced", testFunc)
	Debounced(5*time.Second, job)
	time.Sleep(3 * time.Second)
	Debounced(5*time.Second, job)
	time.Sleep(6 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.FailNow()
	}
}

func TestNTimesEvery(t *testing.T) {
	testCount = 0
	Start()

	job := New("NTimesEvery", testFunc)
	NTimesEvery(5, 5*time.Second, job)
	time.Sleep(27 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestOnceNow(t *testing.T) {
	testCount = 0
	Start()

	job := New("OnceNow", testFunc)
	OnceNow(job)
	time.Sleep(5 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestEvery(t *testing.T) {
	testCount = 0
	Start()

	job := New("Every", testFunc)
	Every(2*time.Second, job)
	time.Sleep(11 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}
