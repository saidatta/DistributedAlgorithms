package main

import (
	"fmt"
	"sort"
)

// Task represents a periodic task with a period and a deadline.
type Task struct {
	period   int
	deadline int
}

// ByPeriod is a type that implements sort.Interface for []Task based on the
// period field.
type ByPeriod []Task

func (a ByPeriod) Len() int           { return len(a) }
func (a ByPeriod) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPeriod) Less(i, j int) bool { return a[i].period < a[j].period }

// Scheduler represents a rate-monotonic scheduler.
type Scheduler struct {
	tasks []Task
}

// NewScheduler creates a new Scheduler.
func NewScheduler(tasks []Task) *Scheduler {
	// Sort the tasks by period in ascending order.
	sort.Sort(ByPeriod(tasks))

	return &Scheduler{tasks: tasks}
}

// Schedule schedules the periodic tasks according to the rate-monotonic
// algorithm.
func (s *Scheduler) Schedule() {
	for i, task := range s.tasks {
		// Execute the task.
		execute(task)

		// Check if there are any new tasks with a shorter period than the
		// current task.
		for j := i + 1; j < len(s.tasks); j++ {
			if s.tasks[j].period < task.period {
				// Execute the new task and update its deadline.
				execute(s.tasks[j])
				s.tasks[j].deadline += s.tasks[j].period
			}
		}
	}
}

func execute(task Task) {
	fmt.Printf("Executing task with period %d and deadline %d\n", task.period, task.deadline)
}

func main() {
	// Create a new scheduler with some tasks.
	scheduler := NewScheduler([]Task{
		{period: 2, deadline: 2},
		{period: 3, deadline: 3},
		{period: 5, deadline: 5},
	})

	// Schedule the tasks.
	scheduler.Schedule()
}

