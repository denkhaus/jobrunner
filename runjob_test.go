package jobrunner

import (
	"testing"
	"time"
)

type TestJob struct {
}

var testCount int

func (p *TestJob) Run() {
	testCount++
}

func newTestJob() *TestJob {
	return &TestJob{}
}
func TestDebounce(t *testing.T) {
	testCount = 0
	Start()

	Debounced(5*time.Second, newTestJob())
	time.Sleep(3 * time.Second)
	Debounced(5*time.Second, newTestJob())
	time.Sleep(6 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.FailNow()
	}
}

func TestNTimesEvery(t *testing.T) {
	testCount = 0
	Start()

	NTimesEvery(5, 5*time.Second, newTestJob())
	time.Sleep(27 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestOnceNow(t *testing.T) {
	testCount = 0
	Start()

	OnceNow(newTestJob())
	time.Sleep(5 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestEvery(t *testing.T) {
	testCount = 0
	Start()

	Every(2*time.Second, newTestJob())
	time.Sleep(11 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}
