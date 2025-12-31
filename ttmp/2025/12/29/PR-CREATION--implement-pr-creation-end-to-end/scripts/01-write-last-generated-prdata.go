package main

import (
	"flag"
	"fmt"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/prdata"
	"github.com/pkg/errors"
)

func main() {
	repo := flag.String("repo", ".", "Repo path (where .pr-builder lives)")
	title := flag.String("title", "LastTitle", "PR title to write")
	body := flag.String("body", "LastBody", "PR body to write")
	flag.Parse()

	p := prdata.LastGeneratedPRDataPath(*repo)
	err := prdata.WriteGeneratedPRDataToYAMLFile(p, &domain.GeneratedPRData{
		Title: *title,
		Body:  *body,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to write last-generated-pr.yaml"))
	}
	fmt.Println(p)
}
