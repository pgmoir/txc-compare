package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

func xmlTxcToBytes(filepath string) ([]byte, error) {
	xmlFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	var txc TransXChange
	decoder := xml.NewDecoder(xmlFile)
	decoder.CharsetReader = makeCharsetReader
	if err := decoder.Decode(&txc); err != nil {
		return nil, err
	}

	excludeOperators := strings.Split("Northern Rail,Southern,c2c,South Western Railway,East Midlands Trains,Greater Anglia,Great Western Railway,Arriva Trains Wales,Chiltern Railways,Merseyrail,Island Line,Great Northern,Southeastern,ScotRail,Heathrow Express,Gatwick Express,Virgin Trains East Coast,Grand Central,First Hull Trains,First TransPennine Expr,Cross Country,Virgin Trains,Thameslink,West Midlands Trains,London Midland", ",")

	if contains(excludeOperators, txc.Operators.OperatorList[0].OperatorShortName) {
		return nil, errors.New("ignore national rail")
	}

	return txcToBytes(filepath, txc)
}

func txcToBytes(filepath string, txc TransXChange) ([]byte, error) {
	var buffer bytes.Buffer

	sort.Sort(StopPointById(txc.StopPoints.StopPointList))
	for _, sp := range txc.StopPoints.StopPointList {
		buffer.WriteString(fmt.Sprintf("%s\n", sp))
	}

	sort.Sort(RouteSectionById(txc.RouteSections.RouteSectionList))
	for _, rs := range txc.RouteSections.RouteSectionList {
		buffer.WriteString(fmt.Sprintf("%s\n", rs))
		sort.Sort(RouteLinkById(rs.RouteLinkList))
		for _, rsl := range rs.RouteLinkList {
			buffer.WriteString(fmt.Sprintf("%s\n", rsl))
		}
	}

	sort.Sort(RouteById(txc.Routes.RouteList))
	for _, r := range txc.Routes.RouteList {
		buffer.WriteString(fmt.Sprintf("%s\n", r))
	}

	sort.Sort(OperatorById(txc.Operators.OperatorList))
	for _, op := range txc.Operators.OperatorList {
		buffer.WriteString(fmt.Sprintf("%s\n", op))
	}

	sort.Sort(ServiceById(txc.Services.ServiceList))
	for _, s := range txc.Services.ServiceList {
		buffer.WriteString(fmt.Sprintf("%s\n", s))
		sort.Sort(LineById(s.Lines.LineList))
		for _, l := range s.Lines.LineList {
			buffer.WriteString(fmt.Sprintf("%s\n", l))
		}
	}

	fmt.Println("txtToText-", filepath)
	return buffer.Bytes(), nil
}
