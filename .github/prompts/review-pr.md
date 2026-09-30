# Review PR — additional instructions

Appended to vesvabot's default review-pr prompt. Repository rules for vesvai:

- Verify the change matches the SDK surface: anything new in `pkg/sdk/` must
  be exported and covered by a test in the same package.
- Flag changes that skip `make test` or `make lint`.
- Check error handling style: wrapped errors (`%w`), lowercase messages.
- Internal package boundaries matter: no imports of `internal/` from outside
  the module.