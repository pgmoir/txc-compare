# txc-compare
Tool to compare the contents of two zip files with TXC data

This is a simple app developed to benefit business in verifying transport data before it goes into the production pipeline. It extracts 2 zip files, into child zips, then extracts them, converts the contents from xml to txt file (via object structure in memory) then compares the contents of the original 2 zip files, file by file, creating a report of changes. In its first two runs (locally) it helped identify two problems before they went live, thus saving hours of debugging once the data was already deployed.

go run compare.go helper.go log.go main.go txc.go txcbytes.go xml-win1252.go 0 1 true true

README