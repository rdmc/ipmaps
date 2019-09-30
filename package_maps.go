// Read LDAP Template to SCE Packages Map file
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

// TemplatePackageMap - LDAP Template -> SCE package mapping
type TemplatePackageMap map[int]int

// readTemplatePackage() as a csv file
// the encoding/csv is not the most usefull for this task
// poor error report on parsing file for end user
// we shuld write a new simple parse for the maping file.
func readTemplateToPackage() (TemplatePackageMap, error) {
	var tp TemplatePackageMap
	tp = make(map[int]int)

	// load package map file
	csvF, err := os.Open(Conf.TemplateToPackageFile)
	if err != nil {
		return nil, fmt.Errorf("readTemplateToPackage open error: %s", err)
	}
	defer csvF.Close()

	csvR := csv.NewReader(csvF)
	csvR.FieldsPerRecord = 2
	csvR.Comment = '#'
	csvR.TrimLeadingSpace = true
	csvR.ReuseRecord = true

	for {
		rec, err := csvR.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("readTemplateToPackage error: %s", err)
		}
		templ, err := strconv.Atoi(rec[0])
		if err != nil {
			return nil, fmt.Errorf("readTemplateToPackage, template strconv.Atoi(%s) error : %s", rec[0], err)
		}
		pack, err := strconv.Atoi(rec[1])
		if err != nil {
			return nil, fmt.Errorf("readTemplateToPackage, packageid strconv.Atoi(%s) error : %s", rec[1], err)
		}
		// check for duplicates
		if _, ok := tp[templ]; ok {
			return nil, fmt.Errorf("readTemplateToPackage, duplicate template %s", rec[0])
		}
		tp[templ] = pack
	}

	return tp, nil
}

func displaySortedMap(m map[int]int) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("%d: %d\n", k, m[k])
	}
}
