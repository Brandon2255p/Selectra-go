package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

const (
	splitterRetract = 600
)

func main() {
	args := os.Args[1:]

	fmt.Println("Starting")
	noInputFile := len(args) != 1
	if noInputFile {
		log.Fatal("No input file")
	}
	inputFile := args[0]
	outputFile := fmt.Sprintf("%s.bak.gcode", inputFile)
	if _, err := os.Stat(inputFile); err != nil {
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

func getToolChangeGCode(currentTool, nextTool int) string {
	headerTemplate := `;      SELECTRA GO     ;
;;; FOR PRUSA I3 BED ;;;
;;;;;   MODE 1     ;;;;;
;;;; FOR LINEAR CAM ;;;;
;; FOR TR8  CAM SCREW ;;
;;;;;;;;TC START;;;;;;;;
;;;; TC from %d to %d ;;;;
`
	header := fmt.Sprintf(headerTemplate, currentTool, nextTool)

	if currentTool == nextTool {
		return header + "T0\n"
	}

	moveNozzel := `M907 E1300		;Set E axis current
G92 E0; Set position
;/// Move nozzle to safe zone & start pump \\\
G91; Relative
G1 E-60 F9000	;Lift nozzle off part and retract
G90; Absolute
G1 X50 Y200.00 F9000	;Go to ooze area
M82; Absolute E
`

	tipShaping := `;/// Perform tip shaping \\
G92 E0 				;Zero E
G90					;Absolute
M203 E90 T0     	;Increase the max feedrate of extruder
G1 E13 F400 		;Extrude some filament
G1 E11.5 F800 		;Pull back
G1 E12.5 F1000 		;Push forward
G1 E-100  F500000 	;Retract as fast as possible
G91					;Relative mode
`

	retractIntoSplitter := `;// Pull into splitter \\
G92 E0 		;Zero E
G1 E-600  F280 ;;;;;;;;;;MUST BE LOW F 
G92 E0
M84 E
;G-Sum 602.5
`
	return header + moveNozzel + tipShaping + retractIntoSplitter + channelChime(nextTool) +
		changeChannel(currentTool, nextTool) + reloadFilament() + primeFilament()
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
	template := `T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
G92 E%d0		;Set the starting position
G1 E%d0 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	gcode := fmt.Sprintf(template, currentTool, nextTool)
	return gcode
}

func reloadFilament() string {
	template := `;/// Preform re-load \\\
G91					;Relative
;M92 E138.54		;Reset to feeder steps/mm (per experiment)
M907 E1300			;Set current again 
G92 E0				;Set position
G1 E1 F8500			;Kick
G1 E600 			;Get past splitter
G1 X28.88 E60 F210	;Get into hotend
					;Sum 582
M907 E1050			;Set amps for constant 
`
	return template
}

func primeFilament() string {
	template := `;/// Start Priming \\\
M83					;Relative E
G90					;Absolute position
G92 E0				;Set position
G1 E60
`
	return template
}
