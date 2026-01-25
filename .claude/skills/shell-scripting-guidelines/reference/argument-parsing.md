# argument-parsing: Argument Parsing

**Guideline:** Validate script arguments and parse flags with clear usage messages.

**Rationale:** Proper argument parsing prevents script misuse, provides helpful feedback to users, and makes scripts more robust and user-friendly.

**Example:**

```sh
# ✅ Simple argument parsing
parse_args() {
    if [ "$#" -lt 1 ]; then
        die "Usage: $0 <input> [output]" 2
    fi

    input="$1"
    output="${2:-output.txt}"

    if [ ! -f "$input" ]; then
        die "Input file not found: $input" 2
    fi
}

# ✅ Flag parsing
parse_flags() {
    verbose=0
    force=0

    while [ "$#" -gt 0 ]; do
        case "$1" in
            -v|--verbose)
                verbose=1
                shift
                ;;
            -f|--force)
                force=1
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            -*)
                die "Unknown option: $1" 2
                ;;
            *)
                break
                ;;
        esac
    done

    # Remaining args
    if [ "$#" -lt 1 ]; then
        die "Missing required argument" 2
    fi

    input="$1"
}
```

**Techniques:**
- Check argument count before accessing positional parameters
- Provide clear usage messages showing required and optional arguments
- Use a while loop with case statement for flag parsing
- Support both short (-v) and long (--verbose) flags
- Handle unknown flags with an error message
- Validate argument values after parsing
