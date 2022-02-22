package u_conf

import (
	"fmt"
	"testing"
	"time"
)

type Config struct {
	S1 *string          `conf:"s1"`
	S2 []int            `conf:"s2"`
	S3 map[string][]int `conf:"s3"`
	S4 S                `conf:"s4"`
	S5 []S2             `conf:"s5"`
	S6 string           `conf:"s6"`
	S7 map[string]int   `conf:"s7"`
}

type S struct {
	S_S1 uint64 `conf:"s1"`
}

type S2 struct {
	Name string `conf:"name"`
	S_S1 int    `conf:"s1"`
}

type ConfPack struct {
	Opt1   string        `conf:"s_pack1"`
	Opt1_1 *int          `conf:"s_pack2"`
	Opt2   int           `conf:"-"`
	Opt3   time.Duration `conf:"s_pack3" time:"Second"`
	Opt4   time.Duration
	Opt5   time.Duration `conf:"s_pack5" time:"Second"`
}

// В структурах конфига не должно быть полей вида map[string](someStruct)
var json_b = []byte(`{
	"common":{
		"s6":"value_test_in_common"
	},
    "pack":{
        "s1":"value_test",
        "s2":[1,2,3,4,5],
        "s3":{
            "key1":[123, 5],
            "key2":[345]
        },
        "s4":{
            "s1":555555555555
        },
        "s5":[
            {
                "name":"m1",
                "s1":77777777777
            },
            {
                "name":"m2",
                "s1":77777777777
            }
        ],
		"s500":"not_in_structs",
		"s7":{
            "key1":500,
            "key2":700
        }
    },
    "sp":{
        "s_pack1":"value_package",
		"s_pack2":null,
		"s_pack3":10,
		"s_pack5":"10s323ms"
    }
}`)

func TestConfig(t *testing.T) {
	err := parseConfigData(json_b)
	if err != nil {
		fmt.Println(err)
		return
	}

	mainConf := Config{}
	err = ParsePackageConfig(&mainConf, "pack")
	if err != nil {
		fmt.Println(err)
	}

	if mainConf.S1 != nil {
		fmt.Println(*mainConf.S1)
	}
	fmt.Println(mainConf)

	c2 := ConfPack{}
	err = ParsePackageConfig(&c2, "sp")
	if err != nil {
		fmt.Println(err)
	}
	if c2.Opt1_1 != nil {
		fmt.Println(*c2.Opt1_1)
	}
	fmt.Println(c2)
}
