package remapper

import "fmt"

type RemapperResp struct {
	Direction       string `json:"direction"`        //To v1
	InvertDirection string `json:"invert_direction"` //From v1
	DirectionFrom   string `json:"direction_from"`   //To v2
	DirectionTo     string `json:"direction_to"`     //From v2
	Text            []int  `json:"text"`             //Text all versions
}

type RemapperRespV1 struct {
	Direction       string `json:"direction"`
	InvertDirection string `json:"invert_direction"`
	Text            []int  `json:"text"`
}

type RemapperRespV2 struct {
	DirectionFrom string `json:"direction_from"`
	DirectionTo   string `json:"direction_to"`
	Text          []int  `json:"text"`
}

var RemapperPath = "/remap/:version/:content/:direction"

// Define which version we have and set values
func GetPath(v int) {
	fmt.Println(RemapperPath)
}
