# Copilot instructions for `uber-go-lint-style`

## Repository overview

This repository contains the source for Uber’s Go style guide content. The editable source lives in `style_guide/rules/`, and `style_guide/style.md` is generated from those rule files.

`style_guide/rules/SUMMARY.md` controls the document structure, and `style_guide/rules/README.md` explicitly says the directory is used to generate the top-level style guide. Do not edit `style_guide/style.md` directly.

## Build, test, and lint

There are no repo-local build or test scripts checked in here.

The guide’s own linting recommendations are:

- `goimports` to format Go code and manage imports
- `go vet` to catch common mistakes
- `golint` is referenced in the guide, but `revive` is the modern replacement mentioned in the linting section
- `golangci-lint` as the preferred lint runner
- `errcheck` and `staticcheck` as part of the recommended lint set

If you are checking a `Printf`-style helper, `go vet -printfuncs=wrapf,statusf` is called out in the guide as an example.

## High-level architecture

The project is a stitched Markdown book, not an application. The source material is split into one file per topic under `style_guide/rules/`, and those files are assembled into the published guide in `style_guide/style.md`.

The top-level README is intentionally minimal; the detailed structure is in the generated guide and the rule files. `style_guide/rules/preface.txt` and `style_guide/rules/SUMMARY.md` are part of the stitching pipeline, so keep the source order and headings aligned with them.

## Key conventions

- Prefer editing the rule source files in `style_guide/rules/`, not the generated output.
- Keep Markdown examples aligned with the Uber Go style rules already documented in the guide.
- Use `goimports`-style import grouping in examples: standard library imports first, then everything else.
- Prefer field names when initializing structs; the guide allows positional literals only as a narrow exception in small test tables.
- Treat `nil` slices as valid unless a rule explicitly says otherwise.
- The guide is organized around specific conventions such as error handling, struct initialization, import ordering, goroutines, and performance; match existing rule headings and phrasing when adding or revising content.

