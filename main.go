package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

type employeeInput struct {
	ShiftID    int64
	EmployeeID int64
	StartTime  string
	EndTime    string
}

type employeeOutput struct {
	EmployeeID    int64
	StartOfWeek   string
	RegularHours  int
	OverTimeHours int
	InvalidShifts []int64
}

func main() {

	file, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Println("There is an error while reading the file", err)
	}
	var data []employeeInput
	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("An error occurred during unmarshalling json", err)
	}
	results := addTime(data)
	res, err := PrettyStruct(results)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}

func addTime(data []employeeInput) (employeeData []employeeOutput) {
	employeeShifts := make(map[int64][]employeeInput)

	// Loop through the input and employee them by employeeID
	for _, id := range data {
		employeeShifts[id.EmployeeID] = append(employeeShifts[id.EmployeeID], id)
	}

	for id, employee := range employeeShifts {
		output := employeeOutput{
			EmployeeID:    id,
			InvalidShifts: []int64{},
		}

		// Sort the shifts by start time
		sort.Slice(employee, func(i int, j int) bool {
			startTime1, _ := time.Parse(time.RFC3339Nano, employee[i].StartTime)
			startTime2, _ := time.Parse(time.RFC3339Nano, employee[j].StartTime)
			return startTime1.Before(startTime2)
		})

		// Iterate the shifts
		lastShift := employeeInput{}
		for _, shift := range employee {

			// Check if current shift overlaps the previous shift.. if so, add both shifts to invalid slice
			shiftStart, _ := time.Parse(time.RFC3339Nano, shift.StartTime)
			lastShiftEnd, _ := time.Parse(time.RFC3339Nano, lastShift.EndTime)
			if !lastShiftEnd.IsZero() && shiftStart.Before(lastShiftEnd) {
				// Mark invalid
				output.InvalidShifts = append(output.InvalidShifts, shift.ShiftID, lastShift.ShiftID)
				lastShift = shift
				continue
			}

			// Figure out the week of the shift
			output.StartOfWeek = startOfWeek(shiftStart).String()
			// Check if the current shift starts before sunday and ends after sunday... if so divide into two results

			// Add hours to regular time, if over 40, add extra time to overtime
			output.RegularHours = int(shiftStart.Hour()) + int(lastShiftEnd.Hour())
			if output.RegularHours > 40 {
				output.OverTimeHours = output.RegularHours - 40
			}
			// Add week to results
			lastShift = shift
		}
		employeeData = append(employeeData, output)
	}
	return

}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func startOfWeek(date time.Time) time.Time {
	// Calculate the difference in days from the current weekday to Sunday (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
	daysUntilSunday := int(time.Sunday - date.Weekday())

	// Adjust the date to the most recent Sunday
	startOfWeek := date.AddDate(0, 0, daysUntilSunday)

	// Set the time to midnight
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	return startOfWeek
}
