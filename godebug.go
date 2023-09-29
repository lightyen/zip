package zip

import (
	"log"
	"os"
	"strings"
)

type godebugSetting struct {
	k string
	v *string
}

var godebugTable [][2]string

func init() {
	godebug, ok := os.LookupEnv("GODEBUG")
	if ok {
		list := strings.Split(godebug, ",")
		for i := range list {
			kv := strings.SplitN(list[i], "=", 2)

			log.Println(list[i])
			log.Println(kv)
			if len(kv) == 2 {
				k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
				godebugTable = append(godebugTable, [2]string{k, v})
			}
		}
	}
}

func godebugNew(key string) *godebugSetting {
	return &godebugSetting{k: key}
}

func (g *godebugSetting) Value() string {
	if g.v != nil {
		return *g.v
	}
	godebug, ok := os.LookupEnv("GODEBUG")
	if ok {
		list := strings.Split(godebug, ",")
		for i := range list {
			kv := strings.SplitN(list[i], "=", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			if g.k != k {
				continue
			}
			v := strings.TrimSpace(kv[1])
			g.v = &v
			break
		}
	}
	if g.v == nil {
		v := ""
		g.v = &v
	}
	return *g.v
}
