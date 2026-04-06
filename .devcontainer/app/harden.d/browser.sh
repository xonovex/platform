# Host browser — prevents container processes from opening URLs in the
# host's browser (minor escape vector, but closes the surface).

unset BROWSER 2>/dev/null
