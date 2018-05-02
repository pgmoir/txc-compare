package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	st := time.Now()

	LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr, os.Stdout)

	files, err := GetFilesToCompare()
	if err != nil {
		fmt.Println("Failed to identify files to compare", err)
		return
	}

	Compare.Println("Comparing transxchange data")
	Compare.Println("")
	Compare.Println("The purpose of this log file is to highlight the changes that have occurred from one version of the transxchange data to another.")
	Compare.Println("This will more than likely be a comparison of the latest with the previous, but it can be run against any versions.")
	Compare.Println("It only compares changes in data elements that are used for importing so NptgLocalities, JourneyPatternSections and VehicleJourneys wont be included.")
	Compare.Println("Where a whole file is added or removed, then only the core service description is included.")
	Compare.Println("Where a file is altered, then changes down to property level are described.")
	Compare.Println("Any suggestions for change should be emailed to philmoir@tfl.gov.uk.")
	Compare.Println("")
	Compare.Printf("Compare of %s with %s on %s", files[0].name, files[1].name, time.Now().Format("02-01-2006 15:04:05"))
	Compare.Println("==========================================")
	Compare.Println("")

	ExtractFiles(files)

	CompareFiles(files)

	Compare.Println("")
	Compare.Printf("Comparison complete")
	Compare.Println("==========================================")
	Compare.Println("")

	ft := time.Now()
	fmt.Println("Complete", int(ft.Sub(st).Seconds()))
}

func GetAllFolder(n string) string {
	return strings.Join([]string{"extract-", n, "/all"}, "")
}

// GetFilesToCompare returns latest file and option to extract, previous file to compare against and option to extract
func GetFilesToCompare() ([]TxcFile, error) {
	l, el, p, ep := GetLatestPrevious()

	zips, _ := FilterZips("data/")

	li := len(zips) - l - 1
	pi := len(zips) - p - 1
	if li < 0 || pi < 0 {
		return nil, errors.New("File indicated not available")
	}

	txcFiles := []TxcFile{
		TxcFile{name: zips[li], extract: el},
		TxcFile{name: zips[pi], extract: ep},
	}

	fmt.Println("latest", txcFiles[0].name, txcFiles[0].extract)
	fmt.Println("previous", txcFiles[1].name, txcFiles[1].extract)

	return txcFiles, nil
}

// GetLatestPrevious identify files to compare and extract from command line, use default if command line params missing
func GetLatestPrevious() (int, bool, int, bool) {
	l, el, p, ep := 0, true, 1, false

	if len(os.Args) > 2 {
		l, _ = strconv.Atoi(os.Args[1])
		p, _ = strconv.Atoi(os.Args[2])
	}

	if len(os.Args) > 3 {
		el, _ = strconv.ParseBool(os.Args[3])
	}

	if len(os.Args) > 4 {
		ep, _ = strconv.ParseBool(os.Args[4])
	}

	return l, el, p, ep
}

// FilterZips returns list of all zip files in directory
func FilterZips(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	zips := []string{}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), "zip") {
			zips = append(zips, f.Name())
		}
	}
	return zips, nil
}

// ExtractFiles bundles async calls to process the Txc zip file
func ExtractFiles(files []TxcFile) {
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i := 0; i < len(files); i++ {
		go func(f TxcFile) {
			defer wg.Done()
			ProcessTxc(f)
		}(files[i])
	}
	wg.Wait()
	return
}

// ProcessTxc extracts and transforms a single Txc zip file
func ProcessTxc(t TxcFile) {

	fmt.Println("processing", t.name, t.extract)
	if !t.extract {
		return
	}

	dir := "extract-" + t.name + "/"
	os.RemoveAll(dir)
	os.MkdirAll(filepath.Dir(dir), 0777)
	Unzip("data/"+t.name, dir)

	// extract sub zip files and convert to txt format
	var wg sync.WaitGroup
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(files))
	wg.Add(len(files))
	for _, f := range files {
		fmt.Println("extracting", f.Name())
		go func(filename string) {
			defer wg.Done()
			UnzipAndConvert(dir, filename)
		}(f.Name())
	}
	wg.Wait()

	return
}

func UnzipAndConvert(srcfolder, filename string) error {

	extension := filepath.Ext(filename)
	name := filename[0 : len(filename)-len(extension)]
	src := srcfolder + filename
	dest := srcfolder + name

	fmt.Println("unzip of " + src + " has now started")

	err := Unzip(src, dest)
	if err != nil {
		return err
	}

	fmt.Println("unzip of " + src + " has now completed")

	var wgx sync.WaitGroup
	files, err := ioutil.ReadDir(dest)
	if err != nil {
		log.Fatal(err)
	}
	wgx.Add(len(files))
	fmt.Println(src, " contains ", len(files), " files")
	dirall := srcfolder + "all/"
	dirallxml := srcfolder + "allxml/"
	os.MkdirAll(filepath.Dir(dirall), 0777)
	os.MkdirAll(filepath.Dir(dirallxml), 0777)
	for _, f := range files {
		var filename = f.Name()
		var extension = filepath.Ext(filename)
		var name = filename[0 : len(filename)-len(extension)]
		go func(src, dest, destxml, folder string) {
			defer wgx.Done()
			fmt.Println("unconverting ", folder, " file ", src)
			ConvertXmlToTxt(src, dest, destxml)
		}(dest+"/"+filename, dirall+name+".txt", dirallxml+"/"+filename, srcfolder)
	}
	wgx.Wait()

	return nil
}

func ConvertXmlToTxt(src, dest, destxml string) error {
	os.Link(src, destxml)
	d1, err := xmlTxcToBytes(src)
	if err != nil {
		return err
	}
	err1 := ioutil.WriteFile(dest, d1, 0644)
	if err1 != nil {
		return err1
	}
	return nil
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
