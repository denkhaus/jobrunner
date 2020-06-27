// Code generated by "stringer -type=JobState"; DO NOT EDIT.

package jobrunner

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[JobStateInitializing-1]
	_ = x[JobStateIdle-2]
	_ = x[JobStateRunning-3]
	_ = x[JobStateFinished-4]
}

const _JobState_name = "JobStateInitializingJobStateIdleJobStateRunningJobStateFinished"

var _JobState_index = [...]uint8{0, 20, 32, 47, 63}

func (i JobState) String() string {
	i -= 1
	if i < 0 || i >= JobState(len(_JobState_index)-1) {
		return "JobState(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _JobState_name[_JobState_index[i]:_JobState_index[i+1]]
}