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
	Label string `yaml:"label"`
}

type Shortcuts struct {
	Keys  []key `yaml:"keys"`
	dirty bool
}

func (s *Shortcuts) Len() int          { return len(s.Keys) }
func (s *Shortcuts) Swap(i int, j int) { s.Keys[i], s.Keys[j] = s.Keys[j], s.Keys[i] }
func (s *Shortcuts) Less(i int, j int) bool {
	ki, kj := s.Keys[i], s.Keys[j]
	if ki.Label != kj.Label {
		return ki.Label < kj.Label
	}
	return ki.MapTo < kj.MapTo
}

func (s *Shortcuts) MapsTo(c rune, label string) (string, bool) {
	ch := string(c)
	for _, k := range s.Keys {
		if ch == k.Ch && label == k.Label {
			return k.MapTo, true
		}
	}
	return "", false
}

func (s *Shortcuts) index(mapTo, label string) int {
	if !sort.IsSorted(s) {
		log.Fatalf("This should be sorted by MapTo %v", s.Keys)
	}

	i := sort.Search(len(s.Keys), func(i int) bool {
		ki := s.Keys[i]
		if ki.Label != label {
			return ki.Label > label
		}
		return ki.MapTo >= mapTo
	})
	if i >= len(s.Keys) {
		return -1
	}
	k := s.Keys[i]
	if k.Label == label && k.MapTo == mapTo {
		return i
	}
	return -1
}

func (s *Shortcuts) assign(opt, mapTo, label string) bool {
	for _, r := range opt {
		if r == ' ' {
			continue
		}
		if _, has := s.MapsTo(r, label); !has {
			s.Keys = append(s.Keys, key{Ch: string(r), MapTo: mapTo, Label: label})
			return true
		}
	}
	return false
}

func (s *Shortcuts) AutoAssign(mapTo, label string) {
	if idx := s.index(mapTo, label); idx > -1 {
		// Already assigned to some char. No need to assign again.
		return
	}

	s.dirty = true
	defer sort.Sort(s)

	if ok := s.assign(strings.ToLower(mapTo), mapTo, label); ok {
		return
	}
	if ok := s.assign(strings.ToUpper(mapTo), mapTo, label); ok {
		return
	}
	if ok :=
		s.assign("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,.?;{}[]|`~!@#$%^&*()",
			mapTo, label); ok {
		return
	}
	log.Fatalf("Unable to assign any char for %v\n", mapTo)
}

func (s *Shortcuts) BestEffortAssign(ch rune, mapTo, label string) {
	if idx := s.index(mapTo, label); idx > -1 {
		return
	}
	if _, has := s.MapsTo(ch, label); has {
		s.AutoAssign(mapTo, label)
		return
	}
	s.dirty = true
	s.Keys = append(s.Keys, key{Ch: string(ch), MapTo: mapTo, Label: label})
	sort.Sort(s)
	return
}

func (s *Shortcuts) Validate() {
	m := make(map[string]string)
	for _, k := range s.Keys {
		ck := fmt.Sprintf("%s:%s", k.Ch, k.Label)
		if mapTo, has := m[ck]; has {
			log.Fatalf("Same key %q assigned to multiple mappings [%v, %v]\n", k.Ch, k.MapTo, mapTo)
		}
		m[ck] = k.MapTo
	}
}

func (s *Shortcuts) HasLabel(label string) bool {
	for _, k := range s.Keys {
		if label == k.Label {
			return true
		}
	}
	return false
}

func (s *Shortcuts) Print(label string, compact bool) {
	fmt.Println()
	cor := color.New(color.FgRed)
	cog := color.New(color.FgGreen)
	var prev byte
	var count int
	for _, k := range s.Keys {
		if k.Label != label {
			continue
		}

		if prev != k.MapTo[0] {
			fmt.Println()
			if !compact {
				cog.Printf("\t--------------------- %s\n", string(k.MapTo[0]))
			}
			prev = k.MapTo[0]
			count = 0
		} else {
			count++
			if count%3 == 0 {
				fmt.Println()
			}
		}
		fmt.Printf("\t")
		cor.Printf("%s:", k.Ch)
		fmt.Printf(" %-20s\t", k.MapTo)
	}
	fmt.Println()
}

// Persist would write out the mappings in YAML format.
func (s *Shortcuts) Persist(path string) {
	if !s.dirty {
		fmt.Printf("\nUnchanged keyboard shortcuts. Skipping overwrite to %s.\n", path)
		return
	}
	fmt.Printf("\nKeyboard shortcuts changed. Writing to %s.\n", path)
	data, err := yaml.Marshal(s)
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("While syncing to key config file: %v", err)
	}
}

func ParseConfig(path string) *Shortcuts {
	fmt.Printf("Opening file: %v for reading key mappings\n", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("File %v doesn't exist. Creating empty shortcuts\n", path)
		return &Shortcuts{}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read file: %v. Error: %v", path, err)
	}
	s := &Shortcuts{}
	if err := yaml.Unmarshal(data, s); err != nil {
		log.Fatalf("Unable to unmarshal data for file: %v. Error: %v", path, err)
	}
	sort.Sort(s)
	s.Validate()
	return s
}
