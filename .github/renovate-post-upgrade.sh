#!/usr/bin/env bash

# regenerate dummy module in test files with the updated module versions
tmpl="$(awk '
BEGIN { print "module dummy" }
$1=="module" { next }
$1=="go" {
  print
  print ""
  print "require github.com/coopnorge/mage v0.23.3"
  next
}
$1=="require" { req=1 }
req { print }
$1==")" && req { exit }
END {
  print ""
  print "tool github.com/magefile/mage"
  print "replace github.com/coopnorge/mage => {{ . }}"
}
' go.mod)" && \
find . -type f -name '*.go' -exec perl -0777 -pi -e 's|var goModTemplateString = `.*?`|var goModTemplateString = `'"$tmpl"'`|s' {} +

# copy go.sum from root to test directories
find . -type f -name 'go.sum' ! -path './go.sum' -exec cp ./go.sum {} \;
