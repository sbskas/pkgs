package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var testmpath = regexp.MustCompile("^mpath-*")
var testRaid = regexp.MustCompile(`\(([[:digit:]]+),`)

func main() {

	args := os.Args
	if len(args) < 3 {
		fmt.Printf("missing arguments: usage %s major minor\n", args[0])
		os.Exit(1)
	}
	dmsetup := "/sbin/dmsetup"
	if !fileExists(dmsetup) {
		os.Exit(0)
	}

	uuid := os.Args[3]

	splitUuid := strings.SplitAfterN(uuid, "-", 2)
	dmuuid := splitUuid[1]
	dmtbltmp := splitUuid[0]
	dmpartmp := strings.TrimPrefix(dmtbltmp, "part")
	dmpart := strings.ReplaceAll(dmpartmp, "-", "")
	dmtbl := strings.ReplaceAll(dmtbltmp, "-", "")

	if dmpart == dmtbl {
		dmpart = ""
	} else {
		dmtbl = "part"
	}

	var dmserial string
	var dmdeps []byte

	if dmtbl == "part" {
		dmname, _ := exec.Command(dmsetup, "info", "-c", "--noheadings", "-o", "name", "-u", dmuuid).Output()
		fmt.Printf("DM_PATH=%s\nDM_TYPE=raid\n", string(dmname))
		if strings.HasPrefix(dmuuid,"mpath-") {
			dmdeps, _ = exec.Command(dmsetup, "deps", "-u", dmuuid).Output()
			dmserial = strings.TrimPrefix(dmuuid, "mpath-")
		}
	} else if dmtbl == "mpath" {
		dmserial = dmuuid
		dmdeps, _ = exec.Command(dmsetup, "deps", "-u", uuid).Output()
	}
	
	if len(string(dmdeps)) > 0 {
		resregex1 := testRaid.FindString(string(dmdeps))
		resregex2 := strings.TrimPrefix(resregex1, "(")
		resregex := trimSuffix(resregex2, ",")
		var major, _ = strconv.Atoi(resregex)
		if major != 0 {
			switch {
			case major == 94:
				fmt.Println("DMTYPE=ccw")
			case (major >= 104 && major < 111):
				fmt.Println("DM_TYPE=cciss")
			case major == 112:
				fmt.Println("DM_TYPE=cciss")
			case major > 89 && major < 100:
				fmt.Println("DM_TYPE=raid")
			default:
				fmt.Println("DM_TYPE=scsi")

				fmt.Printf("DM_WWN=0x%s\n",trimLeftChars(dmserial, 1))
			}
		}
	}

	if string(dmpart) != "" {
		fmt.Printf("DM_PART=%s\n", dmpart)
	}
	if string(dmdeps) != "" {

	}
	if len(dmserial) > 0 {
		fmt.Printf("DM_SERIAL=%s\n", dmserial)
	}

	os.Exit(0)
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func trimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}