package JSON

import (
	"encoding/json"
	"github.com/getevo/evo/v2/lib/log"
)

func Stringify(object interface{}) string {
	var b, _ = json.Marshal(object)
	return string(b)
}

func Parse(text string, out interface{}) error {
	if text == "" {
		return nil
	}
	var err = json.Unmarshal([]byte(text), out)
	if err != nil {
		log.Error(err)
	}
	return err
}
