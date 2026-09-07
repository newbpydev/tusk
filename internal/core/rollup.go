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
//   - Clamped between 0 and 100.
//   - If all direct subtasks are StatusDone, progress is strictly 100%.
func CalculateProgress(task Task, subtasks []Task) int {
	if len(subtasks) == 0 {
		if task.Status == StatusDone {
			return 100
		}
		if task.Progress >= 100 {
			return 99
		}
		if task.Progress < 0 {
			return 0
		}
		return task.Progress
	}

	allDone := true
	sum := 0
	for _, subtask := range subtasks {
		if subtask.Status != StatusDone {
			allDone = false
		}
		p := subtask.Progress
		if subtask.Status == StatusDone {
			p = 100
		} else {
			if p < 0 {
				p = 0
			} else if p >= 100 {
				p = 99
			}
		}
		sum += p
	}

	if allDone {
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
