package codec

import (
	"encoding/hex"
	"encoding/json"
	"github.com/eoscanada/eos-go"
)

type InjectDataMap map[string]interface{}

type InjectData struct {
	Id    string `json:"id"`
	Payer string `json:"payer"`
	Scope string `json:"scope"`
	Table string `json:"table"`
}
type InjectDataRead struct {
	Data string `json:"data"`
	InjectData
}
type InjectDataWrite struct {
	Data InjectDataMap `json:"data"`
	InjectData
}

func DecodeTableInject(data []byte, abi *eos.ABI) ([]byte, error) {
	dataRd := InjectDataRead{}
	err := json.Unmarshal(data, &dataRd)
	if err != nil {
		return nil, err
	}

	rowData, err := hex.DecodeString(dataRd.Data)
	if err != nil {
		return nil, err
	}

	subactionData, err := abi.DecodeTableRow(eos.TableName(dataRd.Table), rowData)
	if err != nil {
		return nil, err
	}

	subactionDataMap := InjectDataMap{}
	err = json.Unmarshal([]byte(subactionData), &subactionDataMap)
	if err != nil {
		return nil, err
	}

	dataWr := InjectDataWrite{InjectData: dataRd.InjectData, Data: subactionDataMap}

	jsonData, err := json.Marshal(dataWr)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
