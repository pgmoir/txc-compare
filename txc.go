package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Comparison int

// const for Comparison type values
const (
	Match     Comparison = 0
	Added     Comparison = 1
	Removed   Comparison = 2
	Different Comparison = 3
)

type TxcFile struct {
	name    string
	extract bool
}

type TransXChange struct {
	StopPoints    StopPoints
	RouteSections RouteSections
	Routes        Routes
	Operators     Operators
	Services      Services
}

type StopPoints struct {
	StopPointList []StopPoint `xml:"StopPoint"`
}

type StopPoint struct {
	ID                    string `xml:"AtcoCode"`
	Descriptor            Descriptor
	Place                 Place
	StopClassification    StopClassification
	AdministrativeAreaRef string
}

func (s StopPoint) String() string {
	return fmt.Sprintf("StopPoint:%s,%s,%s-E%s:N%s,%s,%s", s.ID, s.Descriptor.CommonName, s.Place.NptgLocalityRef,
		s.Place.Location.Easting, s.Place.Location.Northing, s.StopClassification.StopType, s.AdministrativeAreaRef)
}

// ComparedTo returns the result of a comparison between two objects of StopPoint type
func (s StopPoint) ComparedTo(sc StopPoint) (Comparison, []string) {
	typename := "StopPoint"
	comments := []string{}
	if s.ID < sc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been added (%s)", s.ID, typename, s.Descriptor.CommonName))
		return Added, comments
	} else if s.ID > sc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been dropped (%s)", sc.ID, typename, sc.Descriptor.CommonName))
		return Removed, comments
	} else {
		if s.IsSame(sc) {
			//comments = append(comments, fmt.Sprintf("%s : %s is same", s, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("%s : %s is different (%s)", s.ID, typename, s.Descriptor.CommonName))
		comments = CheckDifference(s.Descriptor.CommonName, sc.Descriptor.CommonName, "CommonName", comments)
		comments = CheckDifference(s.Place.NptgLocalityRef, sc.Place.NptgLocalityRef, "NptgLocalityRef", comments)
		comments = CheckDifference(s.Place.Location.Easting, sc.Place.Location.Easting, "Easting", comments)
		comments = CheckDifference(s.Place.Location.Northing, sc.Place.Location.Northing, "Northing", comments)
		comments = CheckDifference(s.StopClassification.StopType, sc.StopClassification.StopType, "StopType", comments)
		comments = CheckDifference(s.AdministrativeAreaRef, sc.AdministrativeAreaRef, "AdministrativeAreaRef", comments)
		return Different, comments
	}
}

// IsSame returns the result of direct field to field match or not
func (s StopPoint) IsSame(sc StopPoint) bool {
	return (s.ID == sc.ID &&
		s.Descriptor.CommonName == sc.Descriptor.CommonName &&
		s.Place.NptgLocalityRef == sc.Place.NptgLocalityRef &&
		s.Place.Location.Easting == sc.Place.Location.Easting &&
		s.Place.Location.Northing == sc.Place.Location.Northing &&
		s.StopClassification.StopType == sc.StopClassification.StopType &&
		s.AdministrativeAreaRef == sc.AdministrativeAreaRef)
}

type Descriptor struct {
	CommonName string
}

type Place struct {
	NptgLocalityRef string
	Location        Location
}

type Location struct {
	Easting  string
	Northing string
}

type StopClassification struct {
	StopType string
}

type OffStreet struct {
	Rail Rail
}

type Rail struct {
	Platform string
}

type RouteSections struct {
	RouteSectionList []RouteSection `xml:"RouteSection"`
}

type RouteSection struct {
	ID            string      `xml:"id,attr"`
	RouteLinkList []RouteLink `xml:"RouteLink"`
}

func (rs RouteSection) String() string {
	return fmt.Sprintf("RouteSection:%s", rs.ID)
}

func (rs RouteSection) GetID() string {
	i := strings.LastIndex(rs.ID, "-")
	n, _ := strconv.Atoi(rs.ID[i+1:])
	return fmt.Sprintf("%s%02d", rs.ID[:i+1], n)
}

// ComparedTo returns the result of a comparison between two objects of RouteSection type
func (rs RouteSection) ComparedTo(rsc RouteSection) (Comparison, []string) {
	typename := "RouteSection"
	comments := []string{}
	if rs.GetID() < rsc.GetID() {
		comments = append(comments, fmt.Sprintf("%s : %s has been added", rs.ID, typename))
		return Added, comments
	} else if rs.GetID() > rsc.GetID() {
		comments = append(comments, fmt.Sprintf("%s : %s has been dropped", rsc.ID, typename))
		return Removed, comments
	} else {

		if rs.IsSame(rsc) {
			comments = append(comments, fmt.Sprintf("%s : %s is same", rs, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("%s : %s is different", rs.GetID(), typename))
		return Different, comments
	}
}

// ComparedTo returns the result of a comparison between two objects of RouteLink type
func (rl RouteLink) ComparedTo(rlc RouteLink) (Comparison, []string) {
	typename := "RouteLink"
	comments := []string{}
	if rl.GetID() < rlc.GetID() {
		comments = append(comments, fmt.Sprintf("   %s : %s has been added (from \"%s\" to \"%s\")", rl.Base, typename, rl.From.StopPointRef, rl.To.StopPointRef))
		return Added, comments
	} else if rl.GetID() > rlc.GetID() {
		comments = append(comments, fmt.Sprintf("   %s : %s has been dropped (from \"%s\" to \"%s\")", rlc.Base, typename, rlc.From.StopPointRef, rlc.To.StopPointRef))
		return Removed, comments
	} else {

		if rl.IsSame(rlc) {
			comments = append(comments, fmt.Sprintf("   %s : %s is same", rl, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("   %s : %s is different", rl.GetID(), typename))
		comments = CheckDifference(rl.Base, rlc.Base, "   ID", comments)
		comments = CheckDifference(rl.From.StopPointRef, rlc.From.StopPointRef, "   From", comments)
		comments = CheckDifference(rl.To.StopPointRef, rlc.To.StopPointRef, "   To", comments)
		comments = CheckDifference(rl.Distance, rlc.Distance, "   Distance", comments)
		comments = CheckDifference(rl.Direction, rlc.Direction, "   Direction", comments)
		return Different, comments
	}
}

// IsSame returns the result of direct field to field match or not
func (rs RouteSection) IsSame(rsc RouteSection) bool {
	if rs.ID != rsc.ID || len(rs.RouteLinkList) != len(rsc.RouteLinkList) {
		return false
	}

	for i := 0; i < len(rs.RouteLinkList); i++ {
		if !rs.RouteLinkList[i].IsSame(rsc.RouteLinkList[i]) {
			return false
		}
	}
	return true
}

// IsSame returns the result of direct field to field match or not
func (rl RouteLink) IsSame(rlc RouteLink) bool {
	return (rl.Base == rlc.Base &&
		rl.From.StopPointRef == rlc.From.StopPointRef &&
		rl.To.StopPointRef == rlc.To.StopPointRef &&
		rl.Distance == rlc.Distance &&
		rl.Direction == rlc.Direction)
}

type RouteLink struct {
	Base      string `xml:"id,attr"`
	From      StopPointReference
	To        StopPointReference
	Distance  string
	Direction string
}

func (rl RouteLink) GetID() string {
	i := strings.LastIndex(rl.Base, "-")
	n, _ := strconv.Atoi(rl.Base[i+1:])
	return fmt.Sprintf("%s%02d", rl.Base[:i+1], n)
}

func (rl RouteLink) String() string {
	return fmt.Sprintf("RouteLink:%s,%s-%s,%s,%s", rl.GetID(), rl.From.StopPointRef, rl.To.StopPointRef, rl.Distance, rl.Direction)
}

// func (a RouteLink) IsEqual(b RouteLink) bool {

// }

type StopPointReference struct {
	StopPointRef string
}

type Routes struct {
	RouteList []Route `xml:"Route"`
}

type Route struct {
	ID              string `xml:"id,attr"`
	PrivateCode     string
	Description     string
	RouteSectionRef string
}

func (r Route) String() string {
	return fmt.Sprintf("Route:%s,%s,%s:%s", r.ID, r.PrivateCode, r.Description, r.RouteSectionRef)
}

// ComparedTo returns the result of a comparison between two objects of Route type
func (r Route) ComparedTo(rc Route) (Comparison, []string) {
	typename := "Route"
	comments := []string{}
	if r.ID < rc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been added (%s)", r.ID, typename, r.Description))
		return Added, comments
	} else if r.ID > rc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been dropped (%s)", rc.ID, typename, rc.Description))
		return Removed, comments
	} else {

		if r.IsSame(rc) {
			comments = append(comments, fmt.Sprintf("%s : %s is same", r, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("%s : %s is different", r.ID, typename))
		comments = CheckDifference(r.PrivateCode, rc.PrivateCode, "PrivateCode", comments)
		comments = CheckDifference(r.Description, rc.Description, "Description", comments)
		comments = CheckDifference(r.RouteSectionRef, rc.RouteSectionRef, "RouteSectionRef", comments)
		return Different, comments
	}
}

// IsSame returns the result of direct field to field match or not
func (r Route) IsSame(rc Route) bool {
	return (r.ID == rc.ID &&
		r.PrivateCode == rc.PrivateCode &&
		r.Description == rc.Description &&
		r.RouteSectionRef == rc.RouteSectionRef)
}

type Operators struct {
	OperatorList []Operator `xml:"Operator"`
}

type Operator struct {
	ID                    string `xml:"id,attr"`
	OperatorCode          string
	OperatorShortName     string
	OperatorNameOnLicence string
	TradingName           string
}

func (o Operator) String() string {
	return fmt.Sprintf("Operator:%s,%s,%s,%s,%s", o.ID, o.OperatorCode, o.OperatorShortName, o.OperatorNameOnLicence, o.TradingName)
}

// ComparedTo returns the result of a comparison between two objects of Operator type
func (o Operator) ComparedTo(oc Operator) (Comparison, []string) {
	typename := "Operator"
	comments := []string{}
	if o.ID < oc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been added (%s)", o.ID, typename, o.OperatorNameOnLicence))
		return Added, comments
	} else if o.ID > oc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been dropped (%s)", oc.ID, typename, oc.OperatorNameOnLicence))
		return Removed, comments
	} else {

		if o.IsSame(oc) {
			//comments = append(comments, fmt.Sprintf("%s : %s is same", o, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("%s : %s is different", o.ID, typename))
		comments = CheckDifference(o.OperatorCode, oc.OperatorCode, "OperatorCode", comments)
		comments = CheckDifference(o.OperatorShortName, oc.OperatorShortName, "OperatorShortName", comments)
		comments = CheckDifference(o.OperatorNameOnLicence, oc.OperatorNameOnLicence, "OperatorNameOnLicence", comments)
		comments = CheckDifference(o.TradingName, oc.TradingName, "TradingName", comments)
		return Different, comments
	}
}

// IsSame returns the result of direct field to field match or not
func (o Operator) IsSame(oc Operator) bool {
	return (o.ID == oc.ID &&
		o.OperatorCode == oc.OperatorCode &&
		o.OperatorShortName == oc.OperatorShortName &&
		o.OperatorNameOnLicence == oc.OperatorNameOnLicence &&
		o.TradingName == oc.TradingName)
}

type Services struct {
	ServiceList []Service `xml:"Service"`
}

type Service struct {
	ID                    string `xml:"ServiceCode"`
	PrivateCode           string
	Lines                 Lines
	OperatingPeriod       OperatingPeriod
	RegisteredOperatorRef string
	Mode                  string
	Description           string
}

func (s Service) String() string {
	return fmt.Sprintf("Service:%s,%s,%s,%s,%s,%s", s.ID, s.PrivateCode, s.OperatingPeriod, s.RegisteredOperatorRef, s.Mode, s.Description)
}

// ComparedTo returns the result of a comparison between two objects of Service type
func (s Service) ComparedTo(sc Service) (Comparison, []string) {
	typename := "Service"
	comments := []string{}
	if s.ID < sc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been added (%s)", s.ID, typename, s.Description))
		return Added, comments
	} else if s.ID > sc.ID {
		comments = append(comments, fmt.Sprintf("%s : %s has been dropped (%s)", sc.ID, typename, sc.Description))
		return Removed, comments
	} else {
		s.OperatingPeriod.StartDate = sb(s.OperatingPeriod.StartDate)
		sc.OperatingPeriod.StartDate = sb(sc.OperatingPeriod.StartDate)

		if s.IsSame(sc) {
			//comments = append(comments, fmt.Sprintf("%s : %s is same", s, typename))
			return Match, comments
		}

		comments = append(comments, fmt.Sprintf("%s : %s is different", s.ID, typename))
		comments = CheckDifference(s.PrivateCode, sc.PrivateCode, "PrivateCode", comments)
		comments = CheckDifference(s.OperatingPeriod.StartDate, sc.OperatingPeriod.StartDate, "StartDate", comments)
		comments = CheckDifference(s.OperatingPeriod.EndDate, sc.OperatingPeriod.EndDate, "EndDate", comments)
		comments = CheckDifference(s.RegisteredOperatorRef, sc.RegisteredOperatorRef, "RegisteredOperatorRef", comments)
		comments = CheckDifference(s.Mode, sc.Mode, "Mode", comments)
		comments = CheckDifference(s.Description, sc.Description, "Description", comments)
		return Different, comments
	}
}

// IsSame returns the result of direct field to field match or not
func (s Service) IsSame(sc Service) bool {
	return (s.ID == sc.ID &&
		s.PrivateCode == sc.PrivateCode &&
		s.OperatingPeriod.StartDate == sc.OperatingPeriod.StartDate &&
		s.OperatingPeriod.EndDate == sc.OperatingPeriod.EndDate &&
		s.RegisteredOperatorRef == sc.RegisteredOperatorRef &&
		s.Mode == sc.Mode &&
		s.Description == sc.Description)
}

type Lines struct {
	LineList []Line `xml:"Line"`
}

type Line struct {
	ID       string `xml:"id,attr"`
	LineName string
}

func (l Line) String() string {
	return fmt.Sprintf("Line:%s-%s", l.ID, l.LineName)
}

type OperatingPeriod struct {
	StartDate string
	EndDate   string
}

func sb(start string) string {
	layout := "2006-01-02"
	startdate := start
	currentdate := time.Now().Local()
	t, _ := time.Parse(layout, startdate)
	if t.Before(currentdate) {
		return "before"
	}
	return start
}

func (op OperatingPeriod) String() string {
	return fmt.Sprintf("%s|%s", sb(op.StartDate), op.EndDate)
}

type StandardService struct {
	Origin      string
	Destination string
}

func (ss StandardService) String() string {
	return fmt.Sprintf("%s-%s", ss.Origin, ss.Destination)
}

type StopPointById []StopPoint

func (a StopPointById) Len() int           { return len(a) }
func (a StopPointById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StopPointById) Less(i, j int) bool { return a[i].ID < a[j].ID }

type RouteSectionById []RouteSection

func (a RouteSectionById) Len() int           { return len(a) }
func (a RouteSectionById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RouteSectionById) Less(i, j int) bool { return a[i].GetID() < a[j].GetID() }

type RouteLinkById []RouteLink

func (a RouteLinkById) Len() int           { return len(a) }
func (a RouteLinkById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RouteLinkById) Less(i, j int) bool { return a[i].GetID() < a[j].GetID() }

type RouteById []Route

func (a RouteById) Len() int           { return len(a) }
func (a RouteById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RouteById) Less(i, j int) bool { return a[i].ID < a[j].ID }

type OperatorById []Operator

func (a OperatorById) Len() int           { return len(a) }
func (a OperatorById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OperatorById) Less(i, j int) bool { return a[i].ID < a[j].ID }

type ServiceById []Service

func (a ServiceById) Len() int           { return len(a) }
func (a ServiceById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ServiceById) Less(i, j int) bool { return a[i].ID < a[j].ID }

type LineById []Line

func (a LineById) Len() int           { return len(a) }
func (a LineById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a LineById) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (txc *TransXChange) UpdateAndOrdered() {
	sort.Sort(StopPointById(txc.StopPoints.StopPointList))
	sort.Sort(RouteSectionById(txc.RouteSections.RouteSectionList))
	for _, rs := range txc.RouteSections.RouteSectionList {
		sort.Sort(RouteLinkById(rs.RouteLinkList))
	}
	sort.Sort(RouteById(txc.Routes.RouteList))
	sort.Sort(OperatorById(txc.Operators.OperatorList))
	sort.Sort(ServiceById(txc.Services.ServiceList))
	for _, s := range txc.Services.ServiceList {
		sort.Sort(LineById(s.Lines.LineList))
	}
}

// CheckDifference compares one string to another and appends to a difference comments array
func CheckDifference(s, sc, field string, comments []string) []string {
	if s != sc {
		comments = append(comments, fmt.Sprintf("%s changed from \"%s\" to \"%s\"", field, s, sc))
	}
	return comments
}
