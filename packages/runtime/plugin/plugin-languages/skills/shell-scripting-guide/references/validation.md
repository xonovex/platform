# validation: Shellcheck and Shfmt

Lint with `shellcheck script.sh` (or `find . -name '*.sh' -exec shellcheck {} +`) and fix findings rather than suppressing. When a suppression is genuinely warranted, scope it with an inline `# shellcheck disable=SC2086` on the next line only. Format with `shfmt`; house options: `-i 2` (2-space indent), `-bn` (break before binary ops), `-ci` (indent case bodies), `-sr` (space after redirect).

```sh
shfmt -i 2 -bn -ci -sr -w script.sh
```
