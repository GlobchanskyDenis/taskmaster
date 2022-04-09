package cli_parser

import (
	"testing"
)

func TestSplit(t *testing.T) {
	testCases := []struct{
		name     string
		payload  string
		expected []string
	}{
		{
			name: "single word trim",
			payload: " 	 	  		 trim   	  		  		   	 ",
			expected: []string{"trim"},
		},{
			name: "two words trim",
			payload: " 	 	  		 trim   	  		  		   	 -arg1   	  		  		   	 ",
			expected: []string{"trim", "-arg1"},
		},{
			name: "real case",
			payload: "  STATUS 		 -n 100  rabbitmq  ",
			expected: []string{"STATUS", "-n", "100", "rabbitmq"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			entity := newParser(tc.payload)
			entity.split()

			/*	Проверяю общее количество слов в результате в сравнении с ожидаемым  */
			if len(entity.rawCliCommandParts) != len(tc.expected) {
				t.Errorf("Fail: expected %d parts (%#v) got %d parts (%#v)", len(tc.expected), tc.expected, len(entity.rawCliCommandParts), entity.rawCliCommandParts)
				t.FailNow()
			}

			/*	Проверяю чтобы совпало каждое слово по отдельности  */
			for i:=0; i<len(tc.expected); i++ {
				if tc.expected[i] != entity.rawCliCommandParts[i] {
					t.Errorf("Fail: expected '%s' got '%s'", tc.expected[i], entity.rawCliCommandParts[i])
				}
			}
		})
	}
}

func TestParseCommandName(t *testing.T) {
	testCases := []struct{
		name             string
		payload          string
		isValid          bool
		expectedSliceLen int
	}{
		{
			name: "valid no command",
			payload: "  	  ",
			isValid: true,
			expectedSliceLen: 0,
		},{
			name: "valid status command",
			payload: "  status    rabbitmq  ",
			isValid: true,
			expectedSliceLen: 1,
		},{
			name: "valid status command with arguments",
			payload: "  STATUS  -n 100  rabbitmq  ",
			isValid: true,
			expectedSliceLen: 3,
		},{
			name: "invalid status command",
			payload: "  STAtus  rabbitmq  ",
			isValid: false,
		},{
			name: "invalid unknown command",
			payload: "  unknown_command ",
			isValid: false,
		},{
			name: "valid stop command",
			payload: "  stop    rabbitmq  ",
			isValid: true,
			expectedSliceLen: 1,
		},{
			name: "valid start command",
			payload: "  start    rabbitmq  ",
			isValid: true,
			expectedSliceLen: 1,
		},{
			name: "valid restart command",
			payload: "  restart    rabbitmq  ",
			isValid: true,
			expectedSliceLen: 1,
		},{
			name: "valid kill command",
			payload: "  kill    rabbitmq  ",
			isValid: true,
			expectedSliceLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			entity := newParser(tc.payload)
			entity.split()
			if err := entity.parseCommandName(); err != nil {
				if tc.isValid == true {
					t.Errorf("Error: %s", err)
				} else {
					t.Logf("Success: error found as it expected")
				}
			} else {
				if tc.isValid == true {
					/*	Дополнительно проверяю чтобы имя команды было удалено из слайса. Проверка просто по количеству оставшихся нераспарсенных слов  */
					if len(entity.rawCliCommandParts) != tc.expectedSliceLen {
						t.Errorf("Fail: expected %d not parsed slice length, got %d", tc.expectedSliceLen, len(entity.rawCliCommandParts))
					} else {
						t.Logf("Success")
					}
				} else {
					t.Errorf("Fail: expected error, but it not found")
				}
			}
		})
	}
}

func TestParseUnitName(t *testing.T) {
	testCases := []struct{
		name             string
		payload          string
		isValid          bool
		expectedUnitName string
		expectedSliceLen int
	}{
		{
			name: "valid no command",
			payload: "  	  ",
			isValid: true,
			expectedUnitName: "",
			expectedSliceLen: 0,
		},{
			name: "valid status command",
			payload: "  status    rabbitmq  ",
			isValid: true,
			expectedUnitName: "rabbitmq",
			expectedSliceLen: 0,
		},{
			name: "valid status command with arguments",
			payload: "  STATUS  -n 100  rabbitmq  ",
			isValid: true,
			expectedUnitName: "rabbitmq",
			expectedSliceLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			entity := newParser(tc.payload)
			entity.split()
			if err := entity.parseCommandName(); err != nil {
				t.Errorf("Error: %s", err)
			}
			if err := entity.parseUnitName(); err != nil {
				if tc.isValid == true {
					t.Errorf("Error: %s", err)
				} else {
					t.Logf("Success: error found as it expected")
				}
			} else {
				if tc.isValid == true {
					/*	Дополнительно проверяю чтобы имя команды было удалено из слайса. Проверка просто по количеству оставшихся нераспарсенных слов  */
					if len(entity.rawCliCommandParts) != tc.expectedSliceLen {
						t.Errorf("Fail: expected %d not parsed slice length, got %d", tc.expectedSliceLen, len(entity.rawCliCommandParts))
					} else if tc.expectedUnitName != entity.cliCommand.UnitName {
						t.Errorf("Fail: expected unit name %s got %s", tc.expectedUnitName, entity.cliCommand.UnitName)
					} else {
						t.Logf("Success")
					}
				} else {
					t.Errorf("Fail: expected error, but it not found")
				}
			}
		})
	}
}

func TestParseCliCommand(t *testing.T) {
	testCases := []struct{
		name             string
		payload          string
		isValid          bool
	}{
		{
			name: "valid no command",
			payload: "  	  ",
			isValid: true,
		},{
			name: "valid status command",
			payload: "  status    rabbitmq  ",
			isValid: true,
		},{
			name: "valid status command with arguments",
			payload: "  STATUS  -n 100  rabbitmq  ",
			isValid: true,
		},{
			name: "valid status command with arguments",
			payload: "  STATUS  -all  ",
			isValid: true,
		},{
			name: "invalid status -all with program name",
			payload: "  STATUS  -all rabbitmq ",
			isValid: false,
		},{
			name: "invalid status -n without value",
			payload: "  STATUS  -all rabbitmq ",
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			if _, err := ParseCliCommand(tc.payload); err != nil {
				if tc.isValid == true {
					t.Errorf("Error: %s", err)
				} else {
					t.Logf("Success: error found as it expected")
				}
			} else {
				if tc.isValid == true {
					t.Logf("Success")
				} else {
					t.Errorf("Fail: expected error, but it not found")
				}
			}
		})
	}
}