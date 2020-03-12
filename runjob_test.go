package jobrunner

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func stateChanged(j *Job) {
	fmt.Printf("%s-%s p[%s] n[%s]\n",
		j.Name, j.state,
		j.Prev.Format(time.RFC3339),
		j.Next.Format(time.RFC3339))

}

func TestDebounce(t *testing.T) {
	testCount := 0
	testFunc := func() error {
		testCount++
		return nil
	}
	OnJobStateChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx)
	defer cancel()

	job := New("Debounced", testFunc)
	Debounced(5*time.Second, job)
	time.Sleep(3 * time.Second)
	Debounced(5*time.Second, job)
	time.Sleep(10 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.FailNow()
	}
}

func TestNTimesEvery(t *testing.T) {
	testCount := 0
	testFunc := func() error {
		testCount++
		return nil
	}

	OnJobStateChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx)
	defer cancel()

	job := New("NTimesEvery", testFunc)
	NTimesEvery(5, 5*time.Second, job)
	time.Sleep(30 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestOnceNow(t *testing.T) {
	testCount := 0
	testFunc := func() error {
		testCount++
		return nil
	}

	OnJobStateChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx)
	defer cancel()

	job := New("OnceNow", testFunc)
	OnceNow(job)
	time.Sleep(10 * time.Second)

	if testCount != 1 {
		t.Error(testCount)
		t.Fail()
	}
}

func TestEvery(t *testing.T) {
	testCount := 0
	testFunc := func() error {
		testCount++
		return nil
	}
	OnJobStateChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx)
	defer cancel()

	job := New("Every", testFunc)
	Every(2*time.Second, job)
	time.Sleep(11 * time.Second)

	if testCount != 5 {
		t.Error(testCount)
		t.Fail()
	}
}
