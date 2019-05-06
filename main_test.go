package main

import "testing"

func Test_channelChime(t *testing.T) {
	actualGcode := channelChime(0)
	expectedGcode := "M300 S440 P200\n"
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime %s", actualGcode)
	}

	actualGcode = channelChime(1)
	expectedGcode =
		`M300 S440 P200
M300 S440 P100
G4 P100
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime %s", actualGcode)
	}

	actualGcode = channelChime(2)
	expectedGcode =
		`M300 S440 P200
M300 S440 P100
G4 P100
M300 S440 P100
G4 P100
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected chime %s", actualGcode)
	}
}

func Test_changeChannel(t *testing.T) {
	actualGcode := changeChannel(0, 1)
	expectedGcode := `T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
G92 E00		;Set the starting position
G1 E10 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(1, 2)
	expectedGcode = `T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
G92 E10		;Set the starting position
G1 E20 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(2, 0)
	expectedGcode = `T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
G92 E20		;Set the starting position
G1 E00 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}

	actualGcode = changeChannel(4, 2)
	expectedGcode = `T1
;M92 E400		;Set E1 steps/mm for selector cam
M907 E580		;Set amps for selector stepper
G90				;Absolute mode based on entire Selectra range
G92 E40		;Set the starting position
G1 E20 F2000	;Move to the new position
M84 E			;Disable E axis ready for switch
T0				;Force T0 (driver stepper) on
`
	if actualGcode != expectedGcode {
		t.Errorf("unexpected channel ---%s--- vs ---%s---", actualGcode, expectedGcode)
	}
}
