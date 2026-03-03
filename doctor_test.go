package main

import (
	"context"
	"testing"
)

func TestCheckHealth_Race(t *testing.T) {
	d := CreateDoctor(WithWorkers(50))

	apis := make([]API, 100)
	for i := range apis {
		apis[i] = API{
			URL:        "https://example.com",
			MethodType: "GET",
		}
	}

	d.CheckHealth(context.Background(), apis)
}
