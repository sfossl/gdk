package command

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestStress_CommandNameBoundaries(t *testing.T) {
	for n := 1; n <= 32; n++ {
		name := strings.Repeat("a", n)
		cmd := Command{
			Name:        name,
			Description: "valid description",
		}
		if err := Validate(cmd); err != nil {
			t.Fatalf("expected valid command name of length %d to pass, got %v", n, err)
		}
	}

	name33 := strings.Repeat("a", 33)
	cmdOver := Command{
		Name:        name33,
		Description: "valid description",
	}
	if err := Validate(cmdOver); !errors.Is(err, ErrNameTooLong) {
		t.Fatalf("expected ErrNameTooLong for 33 runes, got %v", err)
	}
}

func TestStress_DescriptionBoundaries(t *testing.T) {
	for n := 1; n <= 100; n++ {
		desc := strings.Repeat("d", n)
		cmd := Command{
			Name:        "test",
			Description: desc,
		}
		if err := Validate(cmd); err != nil {
			t.Fatalf("expected valid description of length %d to pass, got %v", n, err)
		}
	}

	desc101 := strings.Repeat("d", 101)
	cmdOver := Command{
		Name:        "test",
		Description: desc101,
	}
	if err := Validate(cmdOver); !errors.Is(err, ErrDescriptionTooLong) {
		t.Fatalf("expected ErrDescriptionTooLong for 101 runes, got %v", err)
	}
}

func TestStress_OptionCountBoundaries(t *testing.T) {
	var opts25 []Option
	for i := 0; i < 25; i++ {
		opts25 = append(opts25, StringOption(fmt.Sprintf("opt_%d", i), "description"))
	}
	cmd25 := Command{
		Name:        "test",
		Description: "test command",
		Options:     opts25,
	}
	if err := Validate(cmd25); err != nil {
		t.Fatalf("expected 25 options to pass, got %v", err)
	}

	opts26 := append(opts25, StringOption("opt_25", "description"))
	cmd26 := Command{
		Name:        "test",
		Description: "test command",
		Options:     opts26,
	}
	if err := Validate(cmd26); !errors.Is(err, ErrTooManyOptions) {
		t.Fatalf("expected ErrTooManyOptions for 26 options, got %v", err)
	}
}

func TestStress_ChoiceCountBoundaries(t *testing.T) {
	var choices25 []Choice
	for i := 0; i < 25; i++ {
		choices25 = append(choices25, Choice{
			Name:  fmt.Sprintf("choice_%d", i),
			Value: fmt.Sprintf("val_%d", i),
		})
	}
	opt25 := StringOption("test_opt", "test desc").WithChoices(choices25...)
	cmd := Command{
		Name:        "test",
		Description: "test command",
		Options:     []Option{opt25},
	}
	if err := Validate(cmd); err != nil {
		t.Fatalf("expected 25 choices to pass, got %v", err)
	}

	choices26 := append(choices25, Choice{
		Name:  "choice_25",
		Value: "val_25",
	})
	opt26 := StringOption("test_opt", "test desc").WithChoices(choices26...)
	cmdOver := Command{
		Name:        "test",
		Description: "test command",
		Options:     []Option{opt26},
	}
	if err := Validate(cmdOver); !errors.Is(err, ErrTooManyChoices) {
		t.Fatalf("expected ErrTooManyChoices for 26 choices, got %v", err)
	}
}

func TestStress_ConcurrentValidation(t *testing.T) {
	cmd := Command{
		Name:        "concurrent",
		Description: "concurrent command",
		Options: []Option{
			StringOption("name", "name desc").WithRequired(),
			IntOption("count", "count desc"),
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if err := Validate(cmd); err != nil {
					t.Errorf("unexpected error in concurrent validation: %v", err)
				}
			}
		}()
	}
	wg.Wait()
}

