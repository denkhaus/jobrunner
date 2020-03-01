package jobrunner

import (
	"testing"
	"time"
)

type TestJob struct {
}

var testCount int

func (p TestJob) Run() {
	testCount++
}

func TestDebounce(t *testing.T) {
	testCount = 0
	Start()

	job := New("Debounced", TestJob{})
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

	job := New("NTimesEvery", TestJob{})
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

	job := New("OnceNow", TestJob{})
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

	job := New("Every", TestJob{})
	Every(2*time.Second, job)
	time.Sleep(11 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}
