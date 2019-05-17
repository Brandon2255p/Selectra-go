package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

const (
	splitterRetract       = 600
	initRetract           = 10
	extraRetract          = 5
	oozeX                 = 220
	oozeY                 = 15
	oozeZ                 = 5
	wipeX                 = 220
	wipeY                 = 0
	wipeZ                 = 5
	extrusionModeUnknown  = -1
	extrusionModeAbsolute = 0
	extrusionModeRelative = 1
)

var globalExtrusionMode = extrusionModeUnknown
var lastZ = 0.0
var fileInfo os.FileInfo

func main() {
	args := os.Args[1:]

	fmt.Println("Starting")
	noInputFile := len(args) != 1
	if noInputFile {
		log.Fatal("No input file")
	}
	inputFile := args[0]
	outputFile := fmt.Sprintf("%s.bak.gcode", inputFile)
	if fileInfo, err := os.Stat(inputFile); err != nil {
		log.Fatal("File does not exist")
	}
	generateTempGcode(inputFile, outputFile)

	err := os.Rename(outputFile, inputFile)
	if err != nil {
		log.Fatal(err)
	}
}

func generateTempGcode(inputFile, outputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ofile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer ofile.Close()
	w := bufio.NewWriter(ofile)

	re := regexp.MustCompile(`^T(\d{1})`)
	scanner := bufio.NewScanner(file)
	currentTool := 0
	for scanner.Scan() {
		currentLine := scanner.Text()
		extrusionMode, err := detectExtrusionMode(currentLine)
		if err == nil {
			globalExtrusionMode = extrusionMode
		}
		z, err := detectLastZ(currentLine)
		if err == nil {
			lastZ = z
		}
		matches := re.FindStringSubmatch(currentLine)
		isToolChange := matches != nil
		if isToolChange {
			nextTool, err := strconv.Atoi(matches[1])
			if err == nil {
				currentLine = getToolChangeGCode(currentTool, nextTool)
			} else {
				log.Fatalf("Encounterd and error %v", err)
			}
			currentTool = nextTool
		}
		w.WriteString(currentLine)
		w.WriteString("\n")
	}
	w.Flush()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func detectExtrusionMode(currentLine string) (mode int, err error) {
	re := regexp.MustCompile(`^M(\d{2})`)
	matches := re.FindStringSubmatch(currentLine)
	isExtrusionMode := matches != nil
	if isExtrusionMode {
		switch matches[1] {
		case "82":
			return extrusionModeAbsolute, nil
		case "83":
			return extrusionModeRelative, nil
		}
	}
	return extrusionModeUnknown, errors.New("Not an extrusion mode")
}

func detectLastZ(currentLine string) (z float64, err error) {
	re := regexp.MustCompile(`^G1.+[Zz]([\d\.]+)`)
	matches := re.FindStringSubmatch(currentLine)
	foundZ := matches != nil
	if foundZ {
		z, _ = strconv.ParseFloat(matches[1], 64)
		return
	}
	return 0, errors.New("Not a z move")
}

func getToolChangeGCode(currentTool, nextTool int) string {
	headerTemplate := `
;      SELECTRA GO     ;
;;; FOR PRUSA I3 BED ;;;
;;;;;   MODE 1     ;;;;;
;;;; TC from %d to %d ;;;;
;Z=%f
`
	header := fmt.Sprintf(headerTemplate, currentTool, nextTool, lastZ)

	if currentTool == nextTool {
		return header + "T0\n"
	}

	moveNozzel := fmt.Sprintf(`M907 E1300		;Set E axis current
G92 E0; Set position
;/// Move nozzle to safe zone & start pump \\\
M83					;Relative E Marlin
G91					;Relative movement
G1 E-%d F9000		;Lift nozzle off part and retract
`, initRetract)

	tipShaping := `
;/// Perform tip shaping \\
G92 E0 				;Zero E
M83					;Relative E Marlin
M203 E100 T0     	;Increase the max feedrate of extruder
G1 E9 F2000 		;Push forward
G1 E-100  F500000 	;Retract as fast as possible
G91					;Relative mode
`

	retractIntoSplitter := fmt.Sprintf(`;// Pull into splitter \\
G92 E0 				;Zero E
G1 E-%d F8000			;m3 Go all the way up
G92 E0
M84 E
`, splitterRetract-50)
	footerTemplate := `
G90					;Absolute position
;;;;      TC END      ;;;;
`
	return header + moveNozzel + moveToOozeArea() + tipShaping + retractIntoSplitter +
		changeChannel(currentTool, nextTool) + reloadFilament() +
		primeFilament() + wipeFilament() + footerTemplate
}

func channelChime(nextTool int) string {
	chimeGCode := `;// Chime to indicate next channel \\
M300 S440 P200
`
	channelChime := `M300 S440 P100
G4 P100
`
	for index := 0; index < nextTool; index++ {
		chimeGCode += channelChime
	}
	return chimeGCode
}

func changeChannel(currentTool, nextTool int) string {
	template := `
;// Switch channel \\
T1
;M92 E400			;Set E1 steps/mm for selector cam
M907 E580			;Set amps for selector stepper
G90					;Absolute mode based on entire Selectra range
M82					;Absolute E
G92 E%d0				;Set the starting position
G1 E%d0 F2000		;Move to the new position
M84 E				;Disable E axis ready for switch
T0					;Force T0 (driver stepper) on
`
	gcode := fmt.Sprintf(template, currentTool, nextTool)
	return gcode
}

func reloadFilament() string {
	template := `
;/// Preform re-load \\\
G91					;Relative
;M92 E138.54		;Reset to feeder steps/mm (per experiment)
M907 E1300			;Set current again 
G92 E0				;Set position
G1 E%d F1500		;Get past splitter
G1 E%d F400			;Get into hotend
M907 E1050			;Set amps for constant 
`
	return fmt.Sprintf(template, splitterRetract+50, initRetract)
}

func primeFilament() string {
	template := `
;/// Start Priming \\\
G90					;Absolute position
G1 X%d Y%d Z%d F50		;Go to ooze area
M83					;Relative E
G1 E30 F50			;Prime
G92 E0				;Reset E position
`
	return fmt.Sprintf(template, oozeX, oozeY, oozeZ)
}

func wipeFilament() string {
	template := `
;/// Start Wiping \\\
M83					;Relative E
G90					;Absolute position
G1 X%d Y%d Z%d F10000	;Clip the ooze
G1 X180 Y10 F10000	;Recover Z
G1 Z%f F10000		;Recover Z
G92 E-1				;Set position
`
	return fmt.Sprintf(template, wipeX, wipeY, wipeZ, lastZ)
}

func moveToOozeArea() string {
	template := `
;/// Go to Ooze Area \\\
G90 				;Absolute movement
`
	if oozeZ > lastZ {
		template += `
G1 Z%d F10000		;Move to ooze area
G1 X%d Y%d F10000	;Move to ooze area
`
		return fmt.Sprintf(template, oozeZ, oozeX, oozeY)
	} else {
		template += `
G1 X%d Y%d F10000	;Move to ooze area
G1 Z%d F10000		;Move to ooze area
`
		return fmt.Sprintf(template, oozeX, oozeY, oozeZ)
	}
}
