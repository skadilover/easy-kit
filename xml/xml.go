package xml

import "github.com/clbanning/x2j"

func GetString(doc string, path string, def string) string {
    if v, err := x2j.DocValue(doc, path); err != nil {
        return def
    } else {
        if retCode, ok := v.(string); !ok {
            return def
        } else {
            return retCode
        }
    }
}
