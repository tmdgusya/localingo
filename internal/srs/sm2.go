package srs

import (
	"math"
	"time"
)

// Item represents the SRS state of a review item
type Item struct {
	Interval   int       `json:"interval"`    // Days until next review
	Repetition int       `json:"repetition"`  // Consecutive correct answers
	EaseFactor float64   `json:"ease_factor"` // Difficulty multiplier (default 2.5)
	DueDate    time.Time `json:"due_date"`    // Next review date
}

// NewItem creates a fresh SRS item
func NewItem() Item {
	return Item{
		Interval:   0,
		Repetition: 0,
		EaseFactor: 2.5,
		DueDate:    time.Now(),
	}
}

// Grade represents the user's performance on a card (0-5)
// For TUI simplicity, we might map Pass/Fail to these values.
// 5: Perfect response
// 3: Correct response with hesitation
// 0: Incorrect response
type Grade int

const (
	GradePerfect   Grade = 5
	GradePass      Grade = 3
	GradeIncorrect Grade = 0
)

// CalculateNextReview updates the item based on the performance grade using SM-2 algorithm
func (i Item) CalculateNextReview(grade Grade) Item {
	next := i

	if grade >= 3 {
		if next.Repetition == 0 {
			next.Interval = 1
		} else if next.Repetition == 1 {
			next.Interval = 6
		} else {
			next.Interval = int(math.Round(float64(next.Interval) * next.EaseFactor))
		}
		next.Repetition++
	} else {
		next.Repetition = 0
		next.Interval = 1
	}

	// Update Ease Factor (EF)
	// EF' = EF + (0.1 - (5-q) * (0.08 + (5-q)*0.02))
	// Minimum EF is 1.3
	q := float64(grade)
	next.EaseFactor = next.EaseFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if next.EaseFactor < 1.3 {
		next.EaseFactor = 1.3
	}

	// Calculate Due Date
	next.DueDate = time.Now().AddDate(0, 0, next.Interval)

	return next
}
