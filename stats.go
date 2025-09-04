package main

import (
	"fmt"
	"sort"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const outOfRange = 99999

type column []int

// the function given a user email returns the commits made in the last 6 months
func processRepositories(email string, months int) (map[int]int, error) {
	filePath := getDotFilePath()
	repos := parseFileLinesToSlice(filePath)
	daysInMap := months * 30
	commits := make(map[int]int, daysInMap)

	for i := daysInMap; i > 0; i-- {
		commits[i] = 0
	}

	for _, path := range repos {
		if err := fillCommits(email, path, commits, daysInMap); err != nil {
			fmt.Printf("Skip repo %s: %v \n", path, err)
		}
	}
	return commits, nil
}

// the function normalizes a time to midnight
func getBeginningOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// the function returns how many days have passed since date if more than max days return outOfRange
func countDaysSinceDate(date time.Time, maxDays int) int {
	days := 0
	now := getBeginningOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(24 * time.Hour)
		days++
		if days > maxDays {
			return outOfRange
		}
	}
	return days
}

// the function walks through all the local branches of the repo at path
// and fills the commits map with counts for the given email
func fillCommits(email string, path string, commits map[int]int, maxDays int) error {
	// instantiate a git repo object from path
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	refs, err := repo.References()
	if err != nil {
		return err
	}

	// track seen commits to avoid double counting
	seen := make(map[string]bool)

	offset := calcOffset()

	// iterate over all local branches refs
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsBranch() {
			return nil
		}

		// walk through the commits of the branch
		iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return nil
		}

		_ = iter.ForEach(func(c *object.Commit) error {
			hash := c.Hash.String()
			if seen[hash] {
				return nil
			}
			seen[hash] = true

			if c.Author.Email != email {
				return nil
			}

			// number of days ago the commit happened
			daysAgo := countDaysSinceDate(c.Author.When, maxDays) + offset
			if daysAgo != outOfRange {
				commits[daysAgo]++
			}
			return nil
		})
		return nil
	})

	return err
}

// the function aligns the commits to match a github-style contribution graph
func calcOffset() int {
	switch time.Now().Weekday() {
	case time.Sunday:
		return 7
	case time.Monday:
		return 6
	case time.Tuesday:
		return 5
	case time.Wednesday:
		return 4
	case time.Thursday:
		return 3
	case time.Friday:
		return 2
	case time.Saturday:
		return 1
	}
	return 0
}

// the function prints the commit graph
func printCommitsStats(commits map[int]int, months int) {
	keys := sortMapIntoSlice(commits)
	cols := buildCols(keys, commits)
	printCells(cols, months)
}

// the function returns a sorted slice of map keys
func sortMapIntoSlice(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// the function organizes commits into column -> weeks and rows -> days
func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}
	for _, k := range keys {
		week := int(k / 7)
		dayInWeek := k % 7

		if dayInWeek == 0 {
			col = column{} // start new week
		}
		col = append(col, commits[k])

		if dayInWeek == 6 {
			cols[week] = col // store completed week
		}
	}
	return cols
}

// the function prints month headers at the top of the graph
func printMonths(months int) {
	week := getBeginningOfDay(time.Now()).AddDate(0, -months, 0)
	month := week.Month()
	fmt.Print("  ")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("  ")
		}
		week = week.Add(7 * 24 * time.Hour)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

// the function prints weekdays labels
func printDayCol(day int) {
	switch day {
	case 1:
		fmt.Printf(" Mon ")
	case 3:
		fmt.Printf(" Wed ")
	case 5:
		fmt.Printf(" Fri ")
	default:
		fmt.Printf("     ")
	}
}

// the function renders the grid row by row
func printCells(cols map[int]column, months int) {
	printMonths(months)
	weeks := (months * 30 / 7)
	for j := 6; j >= 0; j-- {
		for i := weeks + 1; i >= 0; i-- {
			if i == weeks+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
				} else if len(col) > j {
					printCell(col[j], false)
				} else {
					printCell(0, false)
				}
			} else {
				printCell(0, false)
			}
		}
		fmt.Printf("\n")
	}
}

// the function prints a single cell color based on the commit count
func printCell(val int, today bool) {
	escape := "\033[0;37;30m"

	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m" // light-gray
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m" // yellow
	case val >= 10:
		escape = "\033[1;30;42m" // green
	}
	if today {
		escape = "\033[1;37;45m" // highlight today
	}

	if val == 0 {
		fmt.Printf(escape + " - " + "\033[0m")
		return
	}
	fmt.Printf(escape+" %d "+"\033[0m", val)
}

// stats function calculates and prints the stats
func stats(email string, months int) {
	commits, err := processRepositories(email, months)
	if err != nil {
		fmt.Println("error processing repositories", err)
		return
	}
	printCommitsStats(commits, months)
}
