package testcases

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elliotchance/orderedmap"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/testing"
	"github.com/hyperion-hyn/hyperion-tf/utils"
)

func loadTestCases() error {
	mapping, err := identifyTestCaseFiles(".yml")
	if err != nil {
		return err
	}

	for el := mapping.Front(); el != nil; el = el.Next() {
		for _, testCaseFile := range el.Value.([]string) {
			testCase := &testing.TestCase{}
			err := utils.ParseYaml(testCaseFile, testCase)

			if err == nil {
				testCase.Initialize()
				TestCases = append(TestCases, testCase)
			} else {
				fmt.Printf("Failed to parse test case file: %s - error: %s. Please make sure the test case file is valid YAML!\n", testCaseFile, err.Error())
			}
		}
	}

	fmt.Println(fmt.Sprintf("Found a total of %d test case files", len(TestCases)))

	return nil
}

func identifyTestCaseFiles(ext string) (*orderedmap.OrderedMap, error) {
	files := []string{}

	rootPath := filepath.Join(config.Configuration.Framework.BasePath, "testcases")

	if config.Configuration.Framework.Test != "all" {
		rootPath = fmt.Sprintf("%s/%s", rootPath, config.Configuration.Framework.Test)
	}

	// If a specific test case has been specified
	if strings.HasSuffix(rootPath, ext) {
		files = append(files, rootPath)
	} else {
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ext {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	dirs := []string{}
	for _, testCaseFile := range files {
		dir := path.Dir(testCaseFile)
		if !utils.StringSliceContains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}

	mapping := orderedmap.NewOrderedMap()
	for _, dir := range dirs {
		sectionFiles := []string{}

		for _, testCaseFile := range files {
			if strings.Contains(testCaseFile, dir) {
				sectionFiles = append(sectionFiles, testCaseFile)
			}

			sort.Sort(utils.ByNumericalFilename(sectionFiles))
			mapping.Set(dir, sectionFiles)
		}
	}

	return mapping, nil
}
