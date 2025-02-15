package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil" // io/ioutil is deprecated, use io and os packages instead
	"os"
	"regexp"
	"strings"

	parser "notashelf.dev/hyprkeys/util/parser"
)

// Read Hyprland configuration file and return lines that start with bind= and bindm=
func readHyprlandConfig() ([]string, []string, []string, map[string]string) {

	// If --test flag is passed, read from test file
	// otherwise read from ~/.config/hypr/hyprland.conf
	var configPath string
	if len(os.Args) > 1 && os.Args[1] == "--test" {
		configPath = "test/hyprland.conf"
	} else {
		configPath = os.Getenv("HOME") + "/.config/hypr/hyprland.conf"
	}

	// Open the file
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	vm := make(map[string]string)

	var kbKeybinds []string
	var mKeybinds []string
	var variables []string
	var variableMap = vm

	for scanner.Scan() {
		line := scanner.Text()

		matched, err := regexp.MatchString("^bind.*[lrme]*.*=", line)
		// TODO: regexp.Compile() instead of regexp.MatchString()

		if err != nil {
			panic(err)
		}

		if matched {
			// If the line starts with any bind type, append it to the keybinds slice
			mKeybinds = append(mKeybinds, line)

		} else if strings.HasPrefix(line, "$") {
			// Probably not the best way to do this, but can't think of another occasion where a line would start with "$"
			// and include "=", yet still not be a variable
			if strings.Contains(line, "=") {
				// Store variables and their values in a map
				// This will be used to replace variables in the markdown table
				// with their values
				variable := strings.SplitN(line, "=", 2)
				variableMap[variable[0]] = variable[1]
			}
		}

	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return kbKeybinds, mKeybinds, variables, variableMap
}

// Return each keybind as a markdown table row
// like this: | <kbd>SUPER + L</kbd> | firefox | , firefox
// we also account for no MOD key.

// Pass both kbKeybinds and mKeybinds to this function
func keybindsToMarkdown(kbKeybinds, mKeybinds []string) []string {
	var markdown []string
	for _, keybind := range kbKeybinds {
		keybind = strings.TrimPrefix(keybind, "bind=")

		// Split "keybind" into a slice of strings
		// based on the comma delimiter
		keybindSlice := strings.SplitN(keybind, ",", 4)

		// Trim whitespace from keybindSlice[1] to keybindSlice[3]
		keybindSlice[1] = strings.TrimSpace(keybindSlice[1])
		keybindSlice[2] = strings.TrimSpace(keybindSlice[2])
		keybindSlice[3] = strings.TrimSpace(keybindSlice[3])

		// Print the keybind as a markdown table row

		// Check if keybindSlice is empty
		// Trim the whitespace and "+" if it is
		if keybindSlice[0] == "" {
			keybindSlice[1] = strings.TrimSpace(keybindSlice[1])
			markdown = append(markdown, "| <kbd>"+keybindSlice[1]+"</kbd> | "+keybindSlice[2]+" | "+keybindSlice[3]+" |")

		} else {
			markdown = append(markdown, "| <kbd>"+keybindSlice[0]+" + "+keybindSlice[1]+"</kbd> | "+keybindSlice[2]+" | "+keybindSlice[3]+" |")
		}
	}

	for _, keybind := range mKeybinds {
		keybind = strings.TrimPrefix(keybind, "bindm=")

		// Split "keybind" into a slice of strings
		// based on the comma delimiter
		keybindSlice := strings.SplitN(keybind, ",", 3)

		// Trim whitespace from keybindSlice[1] to keybindSlice[2]
		keybindSlice[1] = strings.TrimSpace(keybindSlice[1])
		keybindSlice[2] = strings.TrimSpace(keybindSlice[2])

		// Print the keybind as a markdown table row

		// Check if keybindSlice[0] is null
		// Trim the whitespace and "+" if it is
		if keybindSlice[0] == "" {
			markdown = append(markdown, "| <kbd>"+keybindSlice[1]+"</kbd> | | "+keybindSlice[2]+" |")
		} else {
			// put "| |" inbetween the keybindSlice[0] and keybindSlice[1]
			markdown = append(markdown, "| <kbd>"+keybindSlice[0]+" + "+keybindSlice[1]+"</kbd> | | "+keybindSlice[2]+" |")
		}

	}

	return markdown
}

func main() {
	kbKeybinds, mKeybinds, variables, variableMap := readHyprlandConfig()

	// If the first argument is empty, show the help message
	if len(os.Args) == 1 {
		fmt.Println("Usage: hyprkeys [OPTIONS]")
		fmt.Println("Generate a markdown table of keybinds from a Hyprland configuration file.")
		fmt.Println("If no file is specified, the default configuration file is used.")
		fmt.Println("Options:")
		fmt.Println("  -h, --help\t\tShow this help message")
		fmt.Println("  -t, --test\t\tUse the test configuration file")
		fmt.Println("  -m, --markdown\t\tPrint the binds as a markdown table")
		fmt.Println("  -v, --verbose\t\tPrint text as is, without making it pretty")
		fmt.Println("  -V, --version\t\tShow the version number")
	} else if len(os.Args) > 1 {
		// Print args

		// If --verbose is passed as an argument, print the keybinds
		// to the terminal
		if os.Args[1] == "--verbose" {
			for _, keybind := range kbKeybinds {
				fmt.Println(keybind)

			}
			for _, keybind := range mKeybinds {
				println(keybind)
			}
		}

		// If --markdown is passed as an argument, print the keybinds
		// as a markdown table
		if os.Args[1] == "--markdown" {
			markdown := keybindsToMarkdown(kbKeybinds, mKeybinds)
			println("| Keybind | Dispatcher | Command |")
			println("|---------|------------|---------|")
			for _, row := range markdown {
				println(row)
			}
		}

		if os.Args[1] == "--variables" {
			for _, variable := range variables {
				println(variable)
			}

			// Now we replace the variables in the markdown table with their values
			// and print the table if --markdown is also passed as an argument
			markdown := keybindsToMarkdown(kbKeybinds, mKeybinds)
			println("| Keybind | Dispatcher | Command |")
			println("|---------|------------|---------|")
			for _, row := range markdown {
				for key, value := range variableMap {

					row = strings.ReplaceAll(row, key, value)
				}
				println(row)
			}
		}

		if os.Args[1] == "--blocks" {
			file, err := ioutil.ReadFile("test/hyprland.conf") // TODO: make this use configPath
			if err != nil {
				panic(err)
			}
			content := string(file)
			config := parser.Parse(content)
			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("%s\n", data)
			save := parser.BuildConf(config)
			err = ioutil.WriteFile("test/hyprland-generated.conf", []byte(save), 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}
