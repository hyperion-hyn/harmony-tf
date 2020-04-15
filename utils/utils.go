package utils

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// PrefixAddress - generates a string in the format of e.g. Testnet_one1wxt77vendr827zm00tfw2fh46ll7qlz5ey62yg - this is intended to be used for naming keys in the keystore
func PrefixAddress(prefix string, address string) string {
	return fmt.Sprintf("%s_%s", strings.Title(prefix), address)
}

// NetworkTimeoutAdjustment - adjusts the wait time based on different networks
func NetworkTimeoutAdjustment(networkName string, currentTimeout int) int {
	if currentTimeout > 0 {
		switch networkName {
		case "localnet", "staking", "stress":
			currentTimeout = int(math.RoundToEven(float64(currentTimeout) * 1.5))
		}
	}

	return currentTimeout
}

// FileToLines - parse a given text file and split it into a new line delimited slice
func FileToLines(filePath string) (lines []string, err error) {
	data, err := ReadFileToString(filePath)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		lines = strings.Split(string(data), "\n")
		// Remove extra line introduced by strings.Split - see https://play.golang.org/p/sNsAc2xVDT
		if strings.Contains(data, "\n") {
			lines = lines[:len(lines)-1]
		}
	}

	return lines, nil
}

// ParseYaml - parses yaml into a specific type
func ParseYaml(path string, entity interface{}) error {
	yamlData, err := ReadFileToString(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(yamlData), entity)
	if err != nil {
		return err
	}

	return nil
}

// RandomItemFromMap - select a random item from a map
func RandomItemFromMap(itemMap map[string]string) (string, string) {
	var keys []string

	for key := range itemMap {
		keys = append(keys, key)
	}

	randKey := RandomItemFromSlice(keys)
	randItem := itemMap[randKey]

	return randKey, randItem
}

// RandomItemFromSlice - select a random item from a slice
func RandomItemFromSlice(items []string) string {
	rand.Seed(time.Now().UnixNano())
	item := items[rand.Intn(len(items))]

	return item
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ReadFileToString - check if a file exists, proceed to read it to memory if it does
func ReadFileToString(filePath string) (string, error) {
	if fileExists(filePath) {
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return "", err
		}

		return string(data), nil
	} else {
		return "", nil
	}
}

// GlobFiles - find a set of files matching a specific pattern
func GlobFiles(pattern string) ([]string, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// StringSliceContains - checks if a string slice contains a given string
func StringSliceContains(slice []string, str string) bool {
	for _, item := range slice {
		if str == item {
			return true
		}
	}
	return false
}
