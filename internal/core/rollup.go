package core

// CalculateProgress calculates the deterministic progress percentage (0 - 100)
// for a task based on its status and any child subtasks.
//
// Mathematical Rules:
// 1. If len(subtasks) == 0:
//   - If task.Status == StatusDone: 100%
//   - Otherwise: preserves assigned manual progress (task.Progress, 0-99).
//
// 2. If len(subtasks) > 0:
//   - Floor average: floor((1/N) * sum(subtask_i.Progress))
//   - Subtask progress is normalized: negative values clamp to 0; corrupt non-done values > 100 clamp to 99.
//   - If all direct subtasks are StatusDone, progress is strictly 100%.
//   - If all direct subtasks are complete (each is either StatusDone or an intermediate parent with rolled-up 100%),
//     progress is strictly 100%, preserving nested completion without degradation.
//   - Otherwise, the integer floor average is clamped to a maximum of 99% when any subtask remains incomplete.
func CalculateProgress(task Task, subtasks []Task) int {
	if task.Status == StatusDone {
		return 100
	}
	if len(subtasks) == 0 {
		if task.Progress >= 100 {
			return 99
		}
		if task.Progress < 0 {
			return 0
		}
		return task.Progress
	}

	allDone := true
	hasComplete := false
	allComplete := true
	sum := 0
	for _, subtask := range subtasks {
		if subtask.Status != StatusDone {
			allDone = false
		}
		p := subtask.Progress
		if subtask.Status == StatusDone {
			p = 100
			hasComplete = true
		} else {
			if p < 0 {
				p = 0
			} else if p > 100 {
				p = 99 // Clamp corrupt values > 100 to 99
			}
			if p == 100 {
				hasComplete = true
			}
			if p < 100 {
				allComplete = false
			}
		}
		sum += p
	}

	if allDone {
		return 100
	}
	if allComplete && hasComplete {
		return 100
	}
	avg := sum / len(subtasks)
	if avg < 0 {
		return 0
	}
	if avg >= 100 {
		return 99
	}
	return avg
}
