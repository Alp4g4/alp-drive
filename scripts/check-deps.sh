#!/usr/bin/env bash
# Checks the dependency rule of go-clean-architecture on the Go module in the
# current directory. Only imports between the parts of the project are checked:
# the standard library and third-party modules are allowed everywhere.
#
#   domain         may import  domain
#   application    may import  domain, application
#   infrastructure may import  domain, application, infrastructure
#   delivery       may import  domain, application, delivery
#   no package under internal/ imports cmd/
#
# Test files are checked too: a test may not reach outward either.
# Exit code: 0 when the rule holds, 1 on a violation, 2 when it cannot run.
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
	echo "check-deps: go is not installed or not on PATH" >&2
	exit 2
fi

gomod=$(go env GOMOD)
if [ -z "$gomod" ] || [ "$gomod" = /dev/null ]; then
	echo "check-deps: no Go module here, run it from the directory holding go.mod" >&2
	exit 2
fi
module=$(go list -m)

internal="$module/internal"

# The part a package belongs to, from its import path; empty when it is none.
part_of() {
	case "$1" in
		"$internal"/domain | "$internal"/domain/*) echo domain ;;
		"$internal"/application | "$internal"/application/*) echo application ;;
		"$internal"/infrastructure | "$internal"/infrastructure/*) echo infrastructure ;;
		"$internal"/delivery | "$internal"/delivery/*) echo delivery ;;
		"$module"/cmd | "$module"/cmd/*) echo main ;;
	esac
}

allowed() {
	case "$1" in
		domain) echo "domain" ;;
		application) echo "domain application" ;;
		infrastructure) echo "domain application infrastructure" ;;
		delivery) echo "domain application delivery" ;;
	esac
}

format='{{.ImportPath}}{{range .Imports}} {{.}}{{end}}{{range .TestImports}} {{.}}{{end}}{{range .XTestImports}} {{.}}{{end}}'
# Listed first, so a package that does not load (an import cycle, a missing
# module) stops the check instead of being silently skipped.
if ! packages=$(go list -f "$format" ./...); then
	echo "check-deps: go list failed, fix the errors above and run it again" >&2
	exit 2
fi

violations=0
while read -r pkg imports; do
	from=$(part_of "$pkg")
	# main may import everything; packages outside the four parts are not checked.
	[ -z "$from" ] || [ "$from" = main ] && continue
	ok=" $(allowed "$from") "
	for imp in $imports; do
		to=$(part_of "$imp")
		[ -z "$to" ] && continue
		if [[ "$ok" != *" $to "* ]]; then
			echo "${pkg#"$module"/} imports ${imp#"$module"/}: $from may import only: $(allowed "$from")"
			violations=$((violations + 1))
		fi
	done
done <<<"$packages"

if [ "$violations" -gt 0 ]; then
	echo "check-deps: $violations import(s) break the dependency rule"
	exit 1
fi
echo "check-deps: dependency rule respected"
