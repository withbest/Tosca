// Copyright (c) 2024 Fantom Foundation
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at fantom.foundation/bsl11.
//
// Change Date: 2028-4-16
//
// On the date above, in accordance with the Business Source License, use of
// this software will be governed by the GNU Lesser General Public License v3.

package lfvm

import (
	"math"
	"math/rand"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Fantom-foundation/Tosca/go/tosca"
	"github.com/Fantom-foundation/Tosca/go/tosca/vm"
)

func TestNewConverter_UsesDefaultCapacity(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	if want, got := (1 << 30), converter.config.CacheSize; got != want {
		t.Errorf("Expected default cache capacity of %d, got %d", want, got)
	}
}

func TestNewConverter_CacheCanBeDisabled(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{
		CacheSize: -1,
	})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	if want, got := -1, converter.config.CacheSize; got != want {
		t.Errorf("Expected default cache capacity of %d, got %d", want, got)
	}
	if converter.cache != nil {
		t.Errorf("Expected cache to be disabled")
	}
	// Conversion should still work without a nil pointer dereference.
	converter.Convert([]byte{0}, &tosca.Hash{0})
}

func TestNewConverter_TooSmallCapacityLeadsToCreationIssues(t *testing.T) {
	_, err := NewConverter(ConversionConfig{
		CacheSize: maxCachedCodeLength / 2,
	})
	if err == nil {
		t.Fatalf("expected error when creating converter with too small cache size")
	}
}

func TestConverter_LongExampleCode(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	converter.Convert(longExampleCode, nil)
}

func TestConverter_InputsAreCachedUsingHashAsKey(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	code := []byte{byte(vm.STOP)}
	hash := tosca.Hash{byte(1)}
	want := converter.Convert(code, &hash)
	got := converter.Convert(code, &hash)
	if &want[0] != &got[0] { // < it needs to be the same slice
		t.Errorf("cached conversion result not returned")
	}
}

func TestConverter_CacheSizeLimitIsEnforced(t *testing.T) {
	for _, limit := range []int{10, 100, 1000} {
		converter, err := NewConverter(ConversionConfig{
			CacheSize: limit * maxCachedCodeLength,
		})
		if err != nil {
			t.Fatalf("failed to create converter: %v", err)
		}
		for i := 0; i < limit*10; i++ {
			hash := tosca.Hash{byte(i), byte(i >> 8), byte(i >> 16)}
			converter.Convert([]byte{0}, &hash)
		}
		if got := len(converter.cache.Keys()); got > limit {
			t.Errorf("Conversion cache grew to %d entries", got)
		}
	}
}

func TestConverter_ExceedinglyLongCodesAreNotCached(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	if want, got := 0, len(converter.cache.Keys()); want != got {
		t.Errorf("Expected %d entries in the cache, got %d", want, got)
	}
	converter.Convert([]byte{0}, &tosca.Hash{0})
	if want, got := 1, len(converter.cache.Keys()); want != got {
		t.Errorf("Expected %d entries in the cache, got %d", want, got)
	}
	// Codes with an excessive length should not be cached.
	converter.Convert(make([]byte, maxCachedCodeLength+1), &tosca.Hash{1})
	if want, got := 1, len(converter.cache.Keys()); want != got {
		t.Errorf("Expected %d entries in the cache, got %d", want, got)
	}
}

func TestConverter_ResultsAreCached(t *testing.T) {
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	code := []byte{byte(vm.STOP)}
	hash := tosca.Hash{byte(1)}
	want := converter.Convert(code, &hash)
	if got, found := converter.cache.Get(hash); !found || !slices.Equal(want, got) {
		t.Errorf("converted code not added to cache")
	}
}

func TestConverter_ConverterIsThreadSafe(t *testing.T) {
	// This test is to be run with --race to detect concurrency issues.
	const (
		NumGoroutines = 100
		NumSteps      = 1000
	)

	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		t.Fatalf("failed to create converter: %v", err)
	}
	code := []byte{byte(vm.STOP)}
	hash := tosca.Hash{byte(1)}

	var wg sync.WaitGroup
	wg.Add(NumGoroutines)
	for i := 0; i < NumGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < NumSteps; j++ {
				// read a value every go routine is requesting
				converter.Convert(code, &hash)
				// convert a value only this go routine is requesting
				converter.Convert(code, &tosca.Hash{byte(i), byte(j)})
			}
		}(i)
	}
	wg.Wait()
}

func TestConvertWithObserver_MapsEvmToLfvmPositions(t *testing.T) {
	code := []byte{
		byte(vm.ADD),
		byte(vm.PUSH1), 1,
		byte(vm.PUSH3), 1, 2, 3,
		byte(vm.SWAP1),
		byte(vm.JUMPDEST),
	}

	type pair struct {
		evm, lfvm int
	}
	var pairs []pair
	res := convertWithObserver(code, ConversionConfig{}, func(evm, lfvm int) {
		pairs = append(pairs, pair{evm, lfvm})
	})

	want := []pair{
		{0, 0},
		{1, 1},
		{3, 2},
		{7, 4},
		{8, 8},
	}

	if !slices.Equal(pairs, want) {
		t.Errorf("Expected %v, got %v", want, pairs)
	}

	for _, p := range pairs {
		if want, got := OpCode(code[p.evm]), res[p.lfvm].opcode; want != got {
			t.Errorf("Expected %v, got %v", want, got)
		}
	}
}

func TestConvertWithObserver_PreservesJumpDestLocations(t *testing.T) {
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))

	for i := 0; i < 100; i++ {
		code := make([]byte, 100)
		r.Read(code)

		type pair struct {
			evm, lfvm int
		}
		var pairs []pair
		res := convertWithObserver(code, ConversionConfig{}, func(evm, lfvm int) {
			pairs = append(pairs, pair{evm, lfvm})
		})

		// Check that all operations are mapped to matching operations.
		for _, p := range pairs {
			if want, got := OpCode(code[p.evm]), res[p.lfvm].opcode; want != got {
				t.Errorf("Expected %v, got %v", want, got)
			}
		}

		// Check that the position of JUMPDESTs is preserved.
		for _, p := range pairs {
			if vm.OpCode(code[p.evm]) == vm.JUMPDEST {
				if p.evm != p.lfvm {
					t.Errorf("Expected JUMPDEST at %d, got %d", p.evm, p.lfvm)
				}
			}
		}
	}
}

func TestConvert_ProgramCounterBeyond16bitAreConvertedIntoInvalidInstructions(t *testing.T) {
	max := math.MaxUint16
	positions := []int{0, 1, max / 2, max - 1, max, max + 1}
	code := make([]byte, max+2)
	for _, pos := range positions {
		code[pos] = byte(vm.PC)
	}
	res := convert(code, ConversionConfig{})

	for _, pos := range positions {
		want := PC
		if pos > max {
			want = INVALID
		}
		if got := res[pos].opcode; want != got {
			t.Errorf("Expected %v at position %d, got %v", want, pos, got)
		}
	}
}

func benchmarkConvertCode(b *testing.B, code []byte, config ConversionConfig) {
	converter, err := NewConverter(config)
	if err != nil {
		b.Fatalf("failed to create converter: %v", err)
	}
	for i := 0; i < b.N; i++ {
		converter.Convert(code, nil)
	}
}

func BenchmarkConvertLongExampleCodeNoCache(b *testing.B) {
	benchmarkConvertCode(b, longExampleCode, ConversionConfig{CacheSize: -1})
}

func BenchmarkConvertLongExampleCode(b *testing.B) {
	benchmarkConvertCode(b, longExampleCode, ConversionConfig{})
}

func BenchmarkConversionCacheLookupSpeed(b *testing.B) {
	// This benchmark measures the lookup speed of the conversion cache
	// by converting the same code snippet multiple times.
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		b.Fatalf("failed to create converter: %v", err)
	}
	hash := &tosca.Hash{0}
	for i := 0; i < b.N; i++ {
		converter.Convert(nil, hash)
	}
}

func BenchmarkConversionCacheUpdateSpeed(b *testing.B) {
	// This benchmark measures the update speed of the conversion cache
	// by converting codes with different (reported) hashes.
	converter, err := NewConverter(ConversionConfig{})
	if err != nil {
		b.Fatalf("failed to create converter: %v", err)
	}
	for i := 0; i < b.N; i++ {
		hash := tosca.Hash{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
		converter.Convert(nil, &hash)
	}
}
