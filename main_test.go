package main

import "testing"

func Test_channelChime(t *testing.T) {
	actualGcode := channelChime(0)
	expectedGcode := `;// Chime to indicate next channel \\
M300 S440 P200
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = channelChime(1)
	expectedGcode =
		`;// Chime to indicate next channel \\
M300 S440 P200
M300 S440 P100
G4 P100
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = channelChime(2)
	expectedGcode =
		`;// Chime to indicate next channel \\
M300 S440 P200
M300 S440 P100
G4 P100
M300 S440 P100
G4 P100
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}
}

func Test_changeChannel(t *testing.T) {
	actualGcode := changeChannel(0, 1)
	expectedGcode := `
;// Switch channel \\
T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
M82				;Absolute E
G92 E00			;Set the starting position
G1 E10 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(1, 2)
	expectedGcode = `
;// Switch channel \\
T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
M82				;Absolute E
G92 E10			;Set the starting position
G1 E20 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(2, 0)
	expectedGcode = `
;// Switch channel \\
T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
M82				;Absolute E
G92 E20			;Set the starting position
G1 E00 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(4, 2)
	expectedGcode = `
;// Switch channel \\
T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
M82				;Absolute E
G92 E40			;Set the starting position
G1 E20 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}
}

func Test_detectExtrusionMode(t *testing.T) {
	mode, err := detectExtrusionMode("M83")
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if mode != extrusionModeRelative {
		t.Errorf("Unexpected mode")
	}

	mode, err = detectExtrusionMode("M82")
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if mode != extrusionModeAbsolute {
		t.Errorf("Unexpected mode")
	}

	mode, err = detectExtrusionMode("M0")
	if err == nil {
		t.Errorf("Unexpected error")
	}

	mode, err = detectExtrusionMode("M")
	if err == nil {
		t.Errorf("Unexpected error")
	}
}
