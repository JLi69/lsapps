package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Returns whether the file path ends in .desktop, then that must mean
// it is a desktop file (we could also check for the magic bytes but I
// have decided just to check for the file extension for now)
func isDesktopFile(path string) bool {
	if len(path) < len(".desktop") {
		return false
	}

	return path[len(path)-len(".desktop"):] == ".desktop"
}

// Finds an equal sign in the file, used for parsing it as desktop files
// have entries of the form "PropertyName=value" so returning the index of
// the equal sign is useful for splitting it into those two values
func findEqualSign(line string) int {
	for i, ch := range line {
		if ch == '=' {
			return i
		}
	}
	//Return the length of the string to indicate no equal sign was found
	return len(line)
}

// Returns a map of variable names found in the desktop file and the values
// associated with those variables
// NOTE: This does not fully follow the actual standard for desktop files and
// likely is missing some features but likely will work on most desktop files
// and is simple and sufficient for this use case
func parseDesktopFile(path string) map[string]string {
	info := make(map[string]string)

	//Attempt to open the desktop file
	file, err := os.Open(path)

	//if not able to open, return an empty map
	if err != nil {
		return info
	}

	defer file.Close()

	//read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		//Check if the line is of the form [*] and not equal to the magic bytes
		//This is so that we don't read some stuff that we don't necessarily need
		if len(line) > 0 && line[0] == '[' && line[len(line)-1] == ']' && line != "[Desktop Entry]" {
			return info
		}

		//Split the string into the variable name and string value
		index := findEqualSign(line)
		if index == len(line) {
			continue
		}

		name := line[:index]
		str := line[(index + 1):]

		//If the line is valid, add it to the map
		if len(name) > 0 && len(str) > 0 {
			info[name] = str
		}
	}

	return info
}

// Strips %f %F %u %U from the exec string as some desktop files might have
// that but for our purposes this is not necessary and just pollutes the string
func stripExec(exec string) string {
	execStripped := strings.Replace(exec, "%F", "", -1)
	execStripped = strings.Replace(execStripped, "%f", "", -1)
	execStripped = strings.Replace(execStripped, "%U", "", -1)
	execStripped = strings.Replace(execStripped, "%u", "", -1)
	execStripped = strings.TrimSpace(execStripped)
	return execStripped
}

// Returns the name of the application
// If the exec string is shorter than the name, it will prefer that over the
// actual name UNLESS the exec string contains '/' or '='
// '/' is excluded to avoid having '/usr/bin/*' showing up as a name and
// '=' is excluded to avoid breaking the parser of 'dmenu_alias'
func getName(name string, exec string) string {
	if len(name) < len(exec) {
		return strings.TrimSpace(name)
	}

	if len(exec) <= len(name) && !strings.Contains(exec, "/") && !strings.Contains(exec, "=") {
		return strings.TrimSpace(exec)
	}

	return strings.TrimSpace(name)
}

// Prints out every application, their preferred name, and exec string
// formatted like this: name=exec
// for the '-a' or '--all' option
func printDesktopInfo(info map[string]string) {
	if len(info) == 0 {
		return
	}

	name := info["Name"]
	exec := info["Exec"]

	//If name or exec strings are empty, ignore this
	if len(name) == 0 {
		return
	}

	if len(exec) == 0 {
		return
	}

	//If NoDisplay is set to true, ignore
	if info["NoDisplay"] == "true" {
		return
	}

	execStripped := stripExec(exec)
	nameLowercase := strings.ToLower(name)
	fmt.Printf(
		"%s=%s\n",
		strings.Replace(getName(nameLowercase, execStripped), "=", "\\=", -1),
		execStripped,
	)
}

// Prints out an alias list for 'dmenu_alias'
// if the executable name is not the same as the preferred name of the
// application, output it as an alias of the format 'name=exec'
// for the 'g' or '--gen-alias' option
func printAliases(info map[string]string) {
	if len(info) == 0 {
		return
	}

	name := info["Name"]
	exec := info["Exec"]

	if len(name) == 0 {
		return
	}

	if len(exec) == 0 {
		return
	}

	if info["NoDisplay"] == "true" {
		return
	}

	execStripped := stripExec(exec)
	nameLowercase := strings.ToLower(name)
	name = strings.Replace(getName(nameLowercase, execStripped), "=", "\\=", -1)
	if name != execStripped {
		fmt.Printf("%s=%s\n", name, execStripped)
	}
}

// Prints out the preferred names of the applications
// default option, also for '--names' and '-n' options
func printNames(info map[string]string) {
	if len(info) == 0 {
		return
	}

	name := info["Name"]
	exec := info["Exec"]

	if len(name) == 0 {
		return
	}

	if len(exec) == 0 {
		return
	}

	if info["NoDisplay"] == "true" {
		return
	}

	execStripped := stripExec(exec)
	nameLowercase := strings.ToLower(name)
	fmt.Printf("%s\n", getName(nameLowercase, execStripped))
}

// Prints out the exec string for the application
// for the '-e' or '--exec' option
func printExec(info map[string]string) {
	if len(info) == 0 {
		return
	}

	exec := info["Exec"]

	if len(exec) == 0 {
		return
	}

	if info["NoDisplay"] == "true" {
		return
	}

	execStripped := stripExec(exec)
	fmt.Printf("%s\n", execStripped)
}

type ArgOption int

const (
	listNames ArgOption = iota
	listExec
	listAll
	outputAliases
)

// Parse the arguments of the application and determine what to do
func parseArgs() ArgOption {
	for _, arg := range os.Args {
		if arg == "--names" || arg == "-n" {
			return listNames
		} else if arg == "--exec" || arg == "-e" {
			return listExec
		} else if arg == "--all" || arg == "-a" {
			return listAll
		} else if arg == "--gen-alias" || arg == "-g" {
			return outputAliases
		}
	}

	//Default option is '-n'
	return listNames
}

// Output desktop info to stdout based on what we are supposed to do
func output(option ArgOption, info map[string]string) {
	switch option {
	case listNames:
		printNames(info)
		break
	case listExec:
		printExec(info)
		break
	case listAll:
		printDesktopInfo(info)
		break
	case outputAliases:
		printAliases(info)
		break
	default:
		break
	}
}

func main() {
	//Get the environment variable 'XDG_DATA_DIRS' so that we can iterate
	//over each of the directories and read $XDG_DATA_DIRS/applications
	//as that is where all the application desktop files are stored
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	dirs := strings.Split(xdgDataDirs, ":")
	//Add $HOME/.local/share to the directory list so that way we don't miss
	//any locally installed applications
	dirs = append(dirs, os.Getenv("HOME")+"/.local/share")

	option := parseArgs()

	for _, dir := range dirs {
		//$XDG_DATA_DIRS/applications is where the desktop files are stored
		applicationDir := dir + "/applications"

		//Attempt to open the directory and read all the files in it
		files, err := os.ReadDir(applicationDir)

		if err != nil {
			continue
		}

		//Iterate over every file and check if it is a desktop file
		for _, file := range files {
			if isDesktopFile(file.Name()) {
				//parse it and then output the info
				info := parseDesktopFile(applicationDir + "/" + file.Name())
				output(option, info)
			}
		}
	}
}
