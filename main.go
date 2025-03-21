package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type City struct {
	Code     string
	Name     string
	Province string
	Country  string
}

type DistributorPermissions struct {
	DistributorName string
	IncludeRegion   []string
	ExcludeRegion   []string
	Parent          *DistributorPermissions // Added: Parent field to form a hierarchy
}

var cities []City
var distributorPermissions []*DistributorPermissions // Change: Store pointers to DistributorPermissions

// Load cities from the CSV file
func loadCities(inputFile string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error while opening the file %s\n", inputFile)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error while reading records")
		return err
	}

	for _, record := range records {
		// Skip the header or irrelevant rows
		if record[0] == "City Code" {
			continue
		}
		city := City{
			Code:     record[0],
			Name:     record[3],
			Province: record[4],
			Country:  record[5],
		}
		cities = append(cities, city)
	}
	return nil
}

// Check if a region is contained in a list of regions (INCLUDE or EXCLUDE), using case-insensitive matching
func isContained(list []string, item string, isExclusion bool) bool {
	item = strings.ToLower(item) // Convert to lowercase for case-insensitive matching
	for _, s := range list {
		s = strings.ToLower(s)
		if isExclusion {
			// Strict matching for exclusions
			if s == item {
				fmt.Printf("Exclusion: Matching '%s' with '%s' - Found!\n", item, s) // Debug print
				return true
			}
		} else {
			// Matching for INCLUDE regions
			if strings.Contains(s, item) {
				fmt.Printf("Inclusion: Matching '%s' with '%s' - Found!\n", item, s) // Debug print
				return true
			}
		}
	}
	return false
}

// Check if the distributor has permission based on INCLUDE and EXCLUDE lists
// Check if the distributor has permission based on INCLUDE and EXCLUDE lists
func hasPermissions(dp *DistributorPermissions, city City) string {
	// Debugging: Print the distributor's include and exclude regions
	fmt.Printf("Checking permissions for Distributor %s\n", dp.DistributorName)
	fmt.Printf("INCLUDE regions: %v\n", dp.IncludeRegion)
	fmt.Printf("EXCLUDE regions: %v\n", dp.ExcludeRegion)

	// Check if the distributor has a parent and if the parent grants permission
	if dp.Parent != nil {
		if hasPermissions(dp.Parent, city) == "NO" {
			return "NO" // If parent doesn't grant permission, deny access
		}
	}

	// Check if the city is in the exclude region first
	if isContained(dp.ExcludeRegion, city.Country, true) || isContained(dp.ExcludeRegion, city.Province, true) || isContained(dp.ExcludeRegion, city.Name, true) {
		fmt.Printf("City %s-%s-%s is in the EXCLUDE list\n", city.Name, city.Province, city.Country) // Debug print
		return "NO"                                                                                  // If excluded, deny access
	}

	// Then, check if the city is in the include region
	if isContained(dp.IncludeRegion, city.Country, false) || isContained(dp.IncludeRegion, city.Province, false) || isContained(dp.IncludeRegion, city.Name, false) {
		return "YES" // Otherwise, grant permission
	}

	return "NO" // If not included, deny access
}

// Check permissions for each city and distributor
func checkPermissions() {
	for _, city := range cities {
		for _, dp := range distributorPermissions {
			result := hasPermissions(dp, city) // Check if the distributor has permission for the city
			fmt.Printf("Distributor %s has permission to distribute in %s-%s-%s: %s\n", dp.DistributorName, city.Name, city.Province, city.Country, result)
		}
	}
}

// Get distributor permissions from user input
func getDistributorPermissionsFromUser() (*DistributorPermissions, error) {
	var distributorName string
	fmt.Println("Please enter Distributor Name:")
	_, err := fmt.Scanln(&distributorName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Permissions for %s\n", distributorName)

	var include, exclude string
	fmt.Println("INCLUDE:")
	_, err = fmt.Scanln(&include)
	if err != nil {
		return nil, err
	}

	fmt.Println("EXCLUDE:")
	_, err = fmt.Scanln(&exclude)
	if err != nil {
		return nil, err
	}

	// Split and trim spaces from include and exclude strings
	splittedIncludeStr := strings.Split(include, ",")
	splittedExcludeStr := strings.Split(exclude, ",")
	for i := range splittedIncludeStr {
		splittedIncludeStr[i] = strings.Trim(splittedIncludeStr[i], " ")
	}

	for j := range splittedExcludeStr {
		splittedExcludeStr[j] = strings.Trim(splittedExcludeStr[j], " ")
	}

	// Create and return DistributorPermissions for the entered distributor
	dp := &DistributorPermissions{
		DistributorName: distributorName,
		IncludeRegion:   splittedIncludeStr,
		ExcludeRegion:   splittedExcludeStr,
	}
	return dp, nil
}

func main() {
	// Load cities from CSV file
	err := loadCities("cities.csv")
	if err != nil {
		fmt.Println("Error loading the input file:", err)
		return
	}

	// Get the first distributor's permissions
	dp1, err := getDistributorPermissionsFromUser()
	if err != nil {
		fmt.Println("Error in getting user input:", err)
		return
	}
	distributorPermissions = append(distributorPermissions, dp1)

	// Get the second distributor's permissions
	dp2, err := getDistributorPermissionsFromUser()
	if err != nil {
		fmt.Println("Error in getting user input:", err)
		return
	}
	// Link dp2 as a sub-distributor of dp1
	dp2.Parent = dp1
	distributorPermissions = append(distributorPermissions, dp2)

	// Check permissions for each distributor and city
	checkPermissions()
}
