package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

func GameStringHashNodes(str string, initial uint32) uint32 {
	if initial == 0 && strings.HasPrefix(str, "@hash(") && strings.HasSuffix(str, ")") {
		var hash uint32
		if n, err := fmt.Sscanf(str, "@hash(%x)", &hash); err != nil {
			log.Panicf("invalid string with @hash prefix %q unhash error: %v", str, err)
		} else if n != 1 {
			log.Panicf("invalid string with @hash prefix %q n count: %v", str, n)
		}
		return hash
	}
	hash := initial
	for _, c := range str {
		hash = hash*127 + uint32(byte(c))
	}
	return hash
}

var hashesMap sync.Map

func loadHashes(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		var hash, init uint32
		n, err := fmt.Sscanf(line, "%x:%x:", &hash, &init)
		if n != 2 {
			continue
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(line, ":", 3)
		input := parts[2]

		if hash != GameStringHashNodes(input, init) {
			log.Printf("!!! Invalid hash in hashes txt (%q)", input)
		}

		if init == 0 {
			GameStringHashRemember(input)
		}
	}
}

func GameStringHashRemember(s string) {
	if strings.HasPrefix(s, "@hash(") {
		return
	}
	hashesMap.Store(GameStringHashNodes(s, 0), s)
}

func loadStringHashes(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		GameStringHashRemember(line)
		GameStringHashRemember(strings.ToUpper(line))
	}
}

func init() {
	go func() {
		if err := loadHashes("hashes.dump.txt"); err != nil {
			log.Printf("Failed to load hash file: %v", err)
		}

		if err := loadStringHashes("strings.dump.txt"); err != nil {
			log.Printf("Failed to load string hashes file: %v", err)
		}
	}()
}

func GameStringUnhashGenerate(hash uint32) string {
	s := ""
	for hash != 0 {
		s += string(rune(hash % 127))
		hash /= 127
	}
	return ReverseString(s)
}

func GameStringUnhashNodes(hash uint32) string {
	if hash == 0 {
		return ""
	}
	v, ok := hashesMap.Load(hash)
	if ok {
		return v.(string)
	} else {
		unhashed := GameStringUnhashGenerate(hash)
		for _, c := range unhashed {
			if c < 0x20 || c >= 0x80 {
				return fmt.Sprintf("@hash(%.8x)", hash)
			}
		}
		return unhashed
	}
}
