package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CompareFiles takes two txc files are compare file existence and file contents
func CompareFiles(txcs []TxcFile) {
	ltxc, ptxc := txcs[0].name, txcs[1].name
	lfs, _ := ioutil.ReadDir(GetAllFolder(ltxc))
	pfs, _ := ioutil.ReadDir(GetAllFolder(ptxc))

	ct, lfi, pfi := len(lfs)+len(pfs), 0, 0
	for ct > 0 {
		lf := lfs[lfi]
		pf := pfs[pfi]
		if lf.Name() < pf.Name() {
			FileAddedToTxc(ltxc, lf.Name(), &ct, &lfi)
		} else if lf.Name() > pf.Name() {
			FileDroppedFromTxc(ptxc, pf.Name(), &ct, &pfi)
		} else {
			CompareMatchedFilesInTxc(ltxc, ptxc, lf, pf, &ct, &lfi, &pfi)
		}
	}
	fmt.Println("Finished")
}

// FileAddedToTxc details files added to transxchange data
func FileAddedToTxc(d, f string, ct, lfi *int) {
	*ct = *ct - 1
	*lfi = *lfi + 1
	LogUniqueFile("added", d, f)
}

// FileDroppedFromTxc details files dropped from transxchange data
func FileDroppedFromTxc(d, f string, ct, pfi *int) {
	*ct = *ct - 1
	*pfi = *pfi + 1
	LogUniqueFile("dropped", d, f)
}

// LogUniqueFile includes log of files added and dropped
func LogUniqueFile(s, d, f string) {
	Compare.Printf("%s has been %s", f, s)
	FileDetails(d, f)
	Compare.Println("")
	Compare.Println("=====================================")
	Compare.Println("")
}

// CompareMatchedFilesInTxc checks for differences between files with matching names in different txc
func CompareMatchedFilesInTxc(ld, pd string, lf, pf os.FileInfo, ct, lfi, pfi *int) {
	*ct = *ct - 2
	*lfi = *lfi + 1
	*pfi = *pfi + 1

	if FileEquals(ld, pd, lf) {
		return
	}

	fmt.Println(lf.Name())

	ltxc := GetTxc(ld, lf)
	ptxc := GetTxc(pd, pf)

	Compare.Printf("%s has changed (size and content: from %d to %d bytes)", lf.Name(), pf.Size(), lf.Size())
	Compare.Println("")

	CompareStopPoints(ptxc.StopPoints.StopPointList, ltxc.StopPoints.StopPointList)
	CompareRouteSections(ptxc.RouteSections.RouteSectionList, ltxc.RouteSections.RouteSectionList)
	CompareRoutes(ptxc.Routes.RouteList, ltxc.Routes.RouteList)
	CompareOperators(ptxc.Operators.OperatorList, ltxc.Operators.OperatorList)
	CompareServices(ptxc.Services.ServiceList, ltxc.Services.ServiceList)

	Compare.Println("=====================================")
	Compare.Println("")
}

func GetTxc(d string, f os.FileInfo) TransXChange {
	filename := f.Name()
	extension := filepath.Ext(filename)
	name := filename[0 : len(filename)-len(extension)]
	fn := "extract-" + d + "/allxml/" + name + ".xml"
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var txc TransXChange
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	if err := decoder.Decode(&txc); err != nil {
		log.Fatal(err)
	}
	txc.UpdateAndOrdered()
	return txc
}

// CompareStopPoints compares one stoppoint against another stoppoint object based on sort by ID
func CompareStopPoints(a, b []StopPoint) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem StopPoint
		var bItem StopPoint
		if aIndex >= len(a) {
			aItem = StopPoint{ID: "ZZZZZZZZZZZ"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = StopPoint{ID: "ZZZZZZZZZZZ"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

// CompareRouteSections compares one RouteSection against another RouteSection object based on sort by ID
func CompareRouteSections(a, b []RouteSection) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem RouteSection
		var bItem RouteSection
		if aIndex >= len(a) {
			aItem = RouteSection{ID: "ZZZZZZZZZZZ-99"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = RouteSection{ID: "ZZZZZZZZZZZ-99"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
				CompareRouteLinks(aItem.RouteLinkList, bItem.RouteLinkList)
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

// CompareRouteLinks compares one RouteLink against another RouteLink object based on sort by ID
func CompareRouteLinks(a, b []RouteLink) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem RouteLink
		var bItem RouteLink
		if aIndex >= len(a) {
			aItem = RouteLink{Base: "ZZZZZZZZZZZ-99"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = RouteLink{Base: "ZZZZZZZZZZZ-99"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

// CompareRoutes compares one Route against another Route object based on sort by ID
func CompareRoutes(a, b []Route) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem Route
		var bItem Route
		if aIndex >= len(a) {
			aItem = Route{ID: "ZZZZZZZZZZZ"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = Route{ID: "ZZZZZZZZZZZ"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

// CompareOperators compares one operator against another operator object based on sort by ID
func CompareOperators(a, b []Operator) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem Operator
		var bItem Operator
		if aIndex >= len(a) {
			aItem = Operator{ID: "ZZZZZZZZZZZ"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = Operator{ID: "ZZZZZZZZZZZ"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

// CompareServices compares one service against another service object based on sort by ID
func CompareServices(a, b []Service) {
	ct, aIndex, bIndex := len(a)+len(b), 0, 0
	for ct > 0 {
		var aItem Service
		var bItem Service
		if aIndex >= len(a) {
			aItem = Service{ID: "ZZZZZZZZZZZ"}
		} else {
			aItem = a[aIndex]
		}
		if bIndex >= len(b) {
			bItem = Service{ID: "ZZZZZZZZZZZ"}
		} else {
			bItem = b[bIndex]
		}
		comparison, comments := aItem.ComparedTo(bItem)
		if comparison == Removed {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			aIndex = aIndex + 1
		} else if comparison == Added {
			Compare.Println(comments[0])
			Compare.Println("")
			ct = ct - 1
			bIndex = bIndex + 1
		} else {
			if comparison == Different {
				for _, s := range comments {
					Compare.Println(s)
				}
				Compare.Println("")
			}
			ct = ct - 2
			aIndex = aIndex + 1
			bIndex = bIndex + 1
		}
	}
	return
}

func FileDetails(d, name string) {
	fn := "extract-" + d + "/all/" + name
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		t := scanner.Text()
		if strings.HasPrefix(t, "Service") {
			Compare.Println(ServiceDetail(t))
		}
	}
}

// ServiceDetail formats into friendlier explanation of service data for compare log purposes
func ServiceDetail(s string) string {
	sd := strings.Split(s, ",")
	op := strings.Split(sd[2], "|")
	svc := Service{
		ID:              sd[0],
		PrivateCode:     sd[1],
		OperatingPeriod: OperatingPeriod{StartDate: op[0], EndDate: op[1]},
		Mode:            sd[4],
		Description:     sd[5],
	}
	return fmt.Sprintf("%s for %s (route: %s) => %s to %s", svc.Mode, svc.PrivateCode, svc.Description, svc.OperatingPeriod.StartDate, svc.OperatingPeriod.EndDate)
}

func FileEquals(ld, pd string, lf os.FileInfo) bool {
	lfn := "extract-" + ld + "/all/" + lf.Name()
	l, _ := ioutil.ReadFile(lfn)
	pfn := "extract-" + pd + "/all/" + lf.Name()
	p, _ := ioutil.ReadFile(pfn)

	return bytes.Equal(l, p)
}
