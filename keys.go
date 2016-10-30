package keys

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/fatih/color"
)

type key struct {
	Ch    string `yaml:"short"`
	MapTo string `yaml:"mapto"`
}

type Shortcuts struct {
	Keys  []key `yaml:"keys"`
	dirty bool
	path  string
}

func (s *Shortcuts) Len() int               { return len(s.Keys) }
func (s *Shortcuts) Less(i int, j int) bool { return s.Keys[i].MapTo < s.Keys[j].MapTo }
func (s *Shortcuts) Swap(i int, j int)      { s.Keys[i], s.Keys[j] = s.Keys[j], s.Keys[i] }

func (s *Shortcuts) MapsTo(c rune) (string, bool) {
	ch := string(c)
	for _, k := range s.Keys {
		if ch == k.Ch {
			return k.MapTo, true
		}
	}
	return "", false
}

func (s *Shortcuts) index(mapTo string) int {
	if !sort.IsSorted(s) {
		log.Fatalf("This should be sorted by MapTo %v", s.Keys)
	}

	idx := sort.Search(len(s.Keys), func(i int) bool {
		return s.Keys[i].MapTo >= mapTo
	})
	if idx < len(s.Keys) && s.Keys[idx].MapTo == mapTo {
		return idx
	}
	return -1
}

func (s *Shortcuts) assign(opt string, mapTo string) bool {
	for _, r := range opt {
		if r == ' ' {
			continue
		}
		if _, has := s.MapsTo(r); !has {
			s.Keys = append(s.Keys, key{Ch: string(r), MapTo: mapTo})
			return true
		}
	}
	return false
}

func (s *Shortcuts) AutoAssign(mapTo string) {
	if idx := s.index(mapTo); idx > -1 {
		// Already assigned to some char. No need to assign again.
		return
	}

	s.dirty = true
	defer sort.Sort(s)

	if ok := s.assign(mapTo, mapTo); ok {
		return
	}
	if ok := s.assign(strings.ToUpper(mapTo), mapTo); ok {
		return
	}
	if ok :=
		s.assign("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,.?;{}[]|`~!@#$%^&*()",
			mapTo); ok {
		return
	}
	log.Fatalf("Unable to assign any char for %v\n", mapTo)
}

func (s *Shortcuts) BestEffortAssign(ch rune, mapTo string) {
	if idx := s.index(mapTo); idx > -1 {
		return
	}
	if _, has := s.MapsTo(ch); has {
		s.AutoAssign(mapTo)
		return
	}
	s.dirty = true
	s.Keys = append(s.Keys, key{Ch: string(ch), MapTo: mapTo})
	sort.Sort(s)
	return
}

func (s *Shortcuts) Validate() {
	m := make(map[string]string)
	for _, k := range s.Keys {
		if mapTo, has := m[k.Ch]; has {
			log.Fatalf("Same key %q assigned to multiple mappings [%v, %v]\n", k.Ch, k.MapTo, mapTo)
		}
		m[k.Ch] = k.MapTo
	}
}

func (s *Shortcuts) Print() {
	fmt.Println()
	cor := color.New(color.FgRed)
	cog := color.New(color.FgGreen)
	var prev byte
	var count int
	for _, k := range s.Keys {
		if prev != k.MapTo[0] {
			cog.Printf("\n\t--------------------- %s", string(k.MapTo[0]))
			prev = k.MapTo[0]
			count = 0
		}
		count++
		if count%3 == 1 {
			fmt.Println()
		}
		fmt.Printf("\t")
		cor.Printf("%s:", k.Ch)
		fmt.Printf(" %-20s\t", k.MapTo)
	}
	fmt.Println()
}

// Persist would write out the mappings in YAML format.
func (s *Shortcuts) Persist() {
	if !s.dirty {
		return
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	if err := ioutil.WriteFile(s.path, data, 0644); err != nil {
		log.Fatalf("While syncing to key config file: %v", err)
	}
}

func ParseConfig(pathConfig string) *Shortcuts {
	fmt.Printf("Opening file: %v for reading key mappings\n", pathConfig)
	if _, err := os.Stat(pathConfig); os.IsNotExist(err) {
		fmt.Printf("File %v doesn't exist. Creating empty shortcuts\n", pathConfig)
		return &Shortcuts{path: pathConfig}
	}

	data, err := ioutil.ReadFile(pathConfig)
	if err != nil {
		log.Fatalf("Unable to read file: %v. Error: %v", pathConfig, err)
	}
	s := &Shortcuts{path: pathConfig}
	if err := yaml.Unmarshal(data, s); err != nil {
		log.Fatalf("Unable to unmarshal data for file: %v. Error: %v", pathConfig, err)
	}
	sort.Sort(s)
	s.Validate()
	return s
}
