package main

import (
	"bufio"
	"io"
	"math"
	"strings"
)

// https://www.marcel.is/random-word-detector/

type RandomTextDetector struct {
	maxWordLength int
	minWordLength int
	weight        map[string]float32
	defaultWeight float32
}

func NewRandomTextDetector(maxWordLength, minWordLength int) *RandomTextDetector {
	return &RandomTextDetector{
		maxWordLength: maxWordLength,
		minWordLength: minWordLength,
	}
}

func bigram(w string) (c map[string]int) {
	c = make(map[string]int)
	if len(w) == 0 {
		return
	}
	if len(w) == 1 {
		c[w] = 1
		return
	}
	for i := 0; i < len(w)-1; i++ {
		c[w[i:i+2]] += 1
	}
	return
}

func adjIdf(maxTermCount int) func(int) float32 {
	return func(termCount int) float32 {
		return float32(math.Log(float64(maxTermCount+1) / float64(termCount+1)))
	}
}

func (r *RandomTextDetector) Fit(rd io.Reader) error {
	tf := make(map[string]int)
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lineTf := bigram(strings.ToLower(scanner.Text()))
		for k, v := range lineTf {
			tf[k] += v
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	maxTf := 0
	for _, v := range tf {
		if v > maxTf {
			maxTf = v
		}
	}
	idf := adjIdf(maxTf)
	weight := make(map[string]float32)
	for k, v := range tf {
		weight[k] = idf(v)
	}
	r.weight = weight
	r.defaultWeight = idf(0)
	return nil
}

func (r *RandomTextDetector) Predict(w string) float32 {
	if len(w) > r.maxWordLength || IsN(w) {
		return 100
	}
	if len(w) < r.minWordLength {
		return 0
	}
	w = strings.ToLower(w)
	var score float32
	for k, v := range bigram(w) {
		weight, ok := r.weight[k]
		if !ok {
			weight = r.defaultWeight
		}
		score += weight * float32(v)
	}
	return score
}
