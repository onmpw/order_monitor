package main

import (
	"fmt"
	"monitor/platform/Jd"
	"monitor/platform/Pdd"
	"testing"
)

/*func TestParsePdd(t *testing.T) {
	ParsePdd()
}
func TestParseJd(t *testing.T) {
	ParseJd()
}

func TestParseWm(t *testing.T) {
	ParseWm()
}

func TestParseCompany(t *testing.T) {
	ParseCompany()
}

func TestParseShop(t *testing.T) {
	ParseShop()
}*/

func BenchmarkParseJd(b *testing.B) {
	fmt.Println(b.N)
	for i := 0; i < b.N; i++ {
		Jd.ParseJd()
	}
}

func BenchmarkParsePdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Pdd.ParsePdd()
	}
}
func BenchmarkParsePddParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Pdd.ParsePdd()
		}
	})
}
