package jobrunner

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var (
	testOpts []Option
)

func stateChanged(j *Job) {
	fmt.Printf("%s-%s p[%s] n[%s]\n",
		j.name, j.currentState,
		j.prev.Format(time.RFC3339),
		j.next.Format(time.RFC3339))

}

func init() {
	testOpts = []Option{
		WithStateUpdateDuration(1 * time.Second),
	}
}

type TestRunner struct {
	count int
}

func (p *TestRunner) Run() error {
	fmt.Println("callback triggered")
	p.count++
	return nil
}

func NewTestRunner() *TestRunner {
	runner := &TestRunner{
		count: 0,
	}
	return runner
}

func TestDebounce(t *testing.T) {
	OnJobChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, testOpts...)
	defer cancel()

	runner := NewTestRunner()
	job := New("Debounced", runner)
	Debounced(5*time.Second, job, BuildWrapperDefault)
	time.Sleep(3 * time.Second)
	Debounced(5*time.Second, job, BuildWrapperDefault)
	time.Sleep(12 * time.Second)

	if runner.count != 1 {
		t.Error(runner.count)
		t.FailNow()
	}
}

func TestNTimesEvery(t *testing.T) {
	OnJobChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, testOpts...)
	defer cancel()

	runner := NewTestRunner()
	job := New("NTimesEvery", runner)
	NTimesEvery(8, 2*time.Second, job, BuildWrapperDefault)
	time.Sleep(20 * time.Second)

	if runner.count != 8 {
		t.Error(runner.count)
		t.Fail()
	}
}

func TestOnceNow(t *testing.T) {
	OnJobChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, testOpts...)
	defer cancel()

	runner := NewTestRunner()
	job := New("OnceNow", runner)
	OnceNow(job, BuildWrapperDefault)
	time.Sleep(10 * time.Second)

	if runner.count != 1 {
		t.Error(runner.count)
		t.Fail()
	}
}

func TestAt(t *testing.T) {
	OnJobChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, testOpts...)
	defer cancel()

	runner := NewTestRunner()
	job := New("At", runner)
	At(Now().Add(5*time.Second), job, BuildWrapperDefault)
	time.Sleep(11 * time.Second)

	if runner.count != 1 {
		t.Error(runner.count)
		t.Fail()
	}
}

func TestEvery(t *testing.T) {
	OnJobChanged(stateChanged)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx)
	defer cancel()
	runner := NewTestRunner()
	job := New("Every", runner)
	Every(2*time.Second, job, BuildWrapperDefault)
	time.Sleep(11 * time.Second)

	if runner.count != 5 {
		t.Error(runner.count)
		t.Fail()
	}
}
