package avlhash

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestAVL(t *testing.T) {
	avl := &AVLTree[string, string]{}
	avl.Add("nome", "Lucas")
	fmt.Printf("Valor salvo: %v\n", *avl.Search("nome"))
	fmt.Printf("Valor nadave: %v\n", avl.Search("bruh"))
}

type MapMapCandidate map[string]int

func (m MapMapCandidate) Add(key string, value int) {
	m[key] = value
}

func (m MapMapCandidate) Remove(key string) {
	delete(m, key)
}

func (m MapMapCandidate) Search(key string) *int {
	ret, ok := m[key]
	if !ok {
		return nil
	}
	return &ret
}

func (m MapMapCandidate) Clear() {
	for k := range m {
		delete(m, k)
	}
}

const stringLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomString(size int) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = stringLetters[rand.Intn(len(stringLetters))]
	}
	return string(buf)
}

func GenerateTestSet(amount, minSize, maxSize int) []string {
	arr := make([]string, amount)
	for i := 0; i < amount; i++ {
		arr[i] = GenerateRandomString(rand.Intn(maxSize-minSize) + minSize)
	}
	return arr
}

type MapCandidate interface {
	Add(string, int)
	Remove(string)
	Search(string) *int
	Clear()
}

func RunMapTest(candidate MapCandidate, keys []string) {
	for _, key := range keys {
		candidate.Add(key, 2)
	}
	for _, key := range keys {
		candidate.Search(key)
	}
	for _, key := range keys {
		candidate.Remove(key)
	}
}

func BenchmarkMap(b *testing.B) {
	var sizes []int
	for i := 1; i < 12; i++ {
		sizes = append(sizes, i*100)
	}
	keyRanges := []struct {
		From int
		To   int
	}{
		{From: 2, To: 10},
		{From: 10, To: 20},
		{From: 30, To: 50},
	}
	implementations := map[string]func() MapCandidate{}
	implementations["std"] = func() MapCandidate { return MapMapCandidate{} }
	implementations["avl"] = func() MapCandidate { return &AVLTree[string, int]{} }

	for _, size := range sizes {
		for _, keyRange := range keyRanges {
			for name, fn := range implementations {
				keys := GenerateTestSet(size, keyRange.From, keyRange.To)
				b.Run(fmt.Sprintf("%s size=%d min=%d max=%d", name, size, keyRange.From, keyRange.To), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						impl := fn()
						for _, key := range keys {
							impl.Add(key, 2)
						}
						for _, key := range keys {
							impl.Search(key)
						}
						for _, key := range keys {
							impl.Remove(key)
						}
					}
				})
			}
		}
	}

}
