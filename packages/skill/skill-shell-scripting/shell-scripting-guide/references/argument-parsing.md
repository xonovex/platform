# argument-parsing: Argument Parsing

Check `$#` before touching positional params; supply optional args via `${2:-default}`. Parse flags with a `while [ "$#" -gt 0 ]` loop over a `case`, `shift`ing each; support short and long forms, reject unknown `-*` with exit code 2, and `break` on the first non-flag.

```sh
verbose=0
while [ "$#" -gt 0 ]; do
    case "$1" in
        -v|--verbose) verbose=1; shift ;;
        -h|--help)    show_usage; exit 0 ;;
        -*)           die "unknown option: $1" 2 ;;
        *)            break ;;
    esac
done
[ "$#" -ge 1 ] || die "missing required argument" 2
input="$1"
```
