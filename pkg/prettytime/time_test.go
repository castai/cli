package prettytime

import (
	"fmt"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{name: "Just now", t: time.Now(), want: "just now"},
		{name: "Second", t: time.Now().Add(
			time.Hour*time.Duration(0) +
				time.Minute*time.Duration(0) +
				time.Second*time.Duration(1)),
			want: "1 second from now",
		},
		{name: "SecondAgo", t: time.Now().Add(
			time.Hour*time.Duration(0) +
				time.Minute*time.Duration(0) +
				time.Second*time.Duration(-1)),
			want: "1 second ago"},
		{name: "Minutes", t: time.Now().Add(time.Hour*time.Duration(0) +
			time.Minute*time.Duration(59) +
			time.Second*time.Duration(59)), want: "60 minutes from now"},
		{name: "Tomorrow", t: time.Now().AddDate(0, 0, 1), want: "tomorrow"},
		{name: "Yesterday", t: time.Now().AddDate(0, 0, -1), want: "yesterday"},
		{name: "Days", t: time.Now().AddDate(0, 0, 5), want: "5 days from now"},
		{name: "Month", t: time.Now().AddDate(0, 1, 2), want: "1 month from now"},
		{name: "MonthAgo", t: time.Now().AddDate(0, -1, 2), want: "4 weeks ago"},
		{name: "Year", t: time.Now().AddDate(50, 0, 0), want: "50 years from now"},
		{name: "YearAgo", t: time.Now().AddDate(-2, 0, 0), want: "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTimeSince := Format(tt.t); gotTimeSince != tt.want {
				t.Errorf("expected=%v, got=%v", gotTimeSince, tt.want)
			}
		})
	}
}

func ExampleFormat() {
	timeSlots := []struct {
		name string
		t    time.Time
	}{
		{name: "Just now", t: time.Now()},
		{name: "Second", t: time.Now().Add(
			time.Hour*time.Duration(0) +
				time.Minute*time.Duration(0) +
				time.Second*time.Duration(1)),
		},
		{name: "SecondAgo", t: time.Now().Add(
			time.Hour*time.Duration(0) +
				time.Minute*time.Duration(0) +
				time.Second*time.Duration(-1)),
		},
		{name: "Minutes", t: time.Now().Add(time.Hour*time.Duration(0) +
			time.Minute*time.Duration(59) +
			time.Second*time.Duration(59))},
		{name: "Tomorrow", t: time.Now().AddDate(0, 0, 1)},
		{name: "Yesterday", t: time.Now().AddDate(0, 0, -1)},
		{name: "Week", t: time.Now().AddDate(0, 0, 7)},
		{name: "WeekAgo", t: time.Now().AddDate(0, 0, -7)},
		{name: "Month", t: time.Now().AddDate(0, 1, 0)},
		{name: "MonthAgo", t: time.Now().AddDate(0, -1, 0)},
		{name: "Year", t: time.Now().AddDate(2, 0, 0)},
		{name: "YearAgo", t: time.Now().AddDate(-2, 0, 0)},
	}

	for _, timeSlot := range timeSlots {
		fmt.Printf("%s = %v\n", timeSlot.name, Format(timeSlot.t))
	}
}

func TestFormatYear(t *testing.T) {
	now := time.Now()

	oneYearFromNow := now.AddDate(1, 0, 0)
	gotTimeSince := Format(oneYearFromNow)
	expected := "1 year from now"
	if gotTimeSince != expected {
		t.Errorf("got %v, want %v", expected, gotTimeSince)
	}
}
