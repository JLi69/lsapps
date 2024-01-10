package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func isDesktopFile(path string) bool {
	if len(path) < len(".desktop") {
		return false
	}

	return path[len(path)-len(".desktop"):] == ".desktop"
}

func findEqualSign(line string) int {
	for i, ch := range line {
		if ch == '=' {
			return i
		}
	}
	return len(line)
}

func parseDesktopFile(path string) map[string]string {
	info := make(map[string]string)

	file, err := os.Open(path)

	if err != nil {
		return info
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 0 && line[0] == '[' && line[len(line)-1] == ']' && line != "[Desktop Entry]" {
			return info
		}

		index := findEqualSign(line)
		if index == len(line) {
			continue
		}

		name := line[:index]
		str := line[(index + 1):]

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

func getName(name string, exec string) string {
	if len(name) < len(exec) {
		return strings.Replace(strings.TrimSpace(name), "=", " ", -1)
	}

	if len(exec) <= len(name) && !strings.Contains(exec, "/") && !strings.Contains(exec, "=") {
		return strings.TrimSpace(exec)
	}

	return strings.Replace(strings.TrimSpace(name), "=", " ", -1)
}

func printDesktopInfo(info map[string]string) {
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

	if info["NoDisplay"] != "true" {
		execStripped := stripExec(exec)
		nameLowercase := strings.ToLower(name)
		fmt.Printf("%s=%s\n", getName(nameLowercase, execStripped), execStripped)
	}
}

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

	if info["NoDisplay"] != "true" {
		execStripped := stripExec(exec)
		nameLowercase := strings.ToLower(name)
		name = strings.Replace(getName(nameLowercase, execStripped), "=", " ", -1)
		if name != execStripped {
			fmt.Printf("%s=%s\n", name, execStripped)
		}
	}
}

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

	if info["NoDisplay"] != "true" {
		execStripped := stripExec(exec)
		nameLowercase := strings.ToLower(name)
		fmt.Printf("%s\n", getName(nameLowercase, execStripped))
	}
}

func printExec(info map[string]string) {
	if len(info) == 0 {
		return
	}

	exec := info["Exec"]

	if len(exec) == 0 {
		return
	}

	if info["NoDisplay"] != "true" {
		execStripped := stripExec(exec)
		fmt.Printf("%s\n", execStripped)
	}
}

type ArgOption int

const (
	listNames ArgOption = iota
	listExec
	listAll
	outputAliases
)

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

	return listAll
}

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
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	dirs := strings.Split(xdgDataDirs, ":")

	option := parseArgs()

	for _, dir := range dirs {
		applicationDir := dir + "/applications"

		files, err := os.ReadDir(applicationDir)

		if err != nil {
			continue
		}

		for _, file := range files {
			if isDesktopFile(file.Name()) {
				info := parseDesktopFile(applicationDir + "/" + file.Name())
				output(option, info)
			}
		}
	}
}
