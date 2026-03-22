#!/bin/sh

set -eu

status=0
titles_tmp=$(mktemp)
trap 'rm -f "$titles_tmp"' EXIT

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  status=1
}

warn() {
  printf 'WARN: %s\n' "$1" >&2
}

for file in recipes/*.md; do
  base=$(basename "$file")
  slug=${base%.md}

  case "$base" in
    *[!a-z0-9.-]*)
      fail "$file: filename must be lowercase kebab-case"
      ;;
  esac

  if ! printf '%s\n' "$slug" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$'; then
    fail "$file: slug must match lowercase-kebab-case"
  fi

  h1_count=$(grep -Ec '^# ' "$file" || true)
  if [ "$h1_count" -eq 0 ]; then
    fail "$file: missing top-level title"
  elif [ "$h1_count" -gt 1 ]; then
    fail "$file: expected exactly one top-level title, found $h1_count"
  fi

  title=$(sed -n 's/^# //p' "$file" | head -n 1)
  if [ -n "$title" ]; then
    printf '%s\t%s\n' "$title" "$file" >>"$titles_tmp"
  fi

  if ! grep -Eq '^\*\*Yield / (Target|Pan Target)\*\*$' "$file"; then
    warn "$file: missing canonical yield block"
  fi

  if ! grep -Eq '^## Ingredients($| \(|:)' "$file"; then
    fail "$file: missing canonical ## Ingredients heading"
  fi

  if ! grep -Eq '^## Process($| \(|:)' "$file"; then
    fail "$file: missing canonical ## Process heading"
  fi

  analysis=$(
    awk '
      BEGIN {
        section = ""
        yield_seen = 0
        yield_bullets = 0
        ingredients_seen = 0
        ingredients_bullets = 0
        ingredients_unstructured = 0
        process_seen = 0
        process_steps = 0
        formula_seen = 0
        formula_bullets = 0
        order_index = 0
        yield_order = 0
        ingredients_order = 0
        process_order = 0
      }
      /^\*\*Yield \/ (Target|Pan Target)\*\*$/ {
        section = "yield"
        yield_seen = 1
        order_index++
        if (yield_order == 0) yield_order = order_index
        next
      }
      /^## / {
        heading = $0
        sub(/^## /, "", heading)
        if (heading ~ /^Ingredients($| \(|:)/) {
          section = "ingredients"
          ingredients_seen = 1
          order_index++
          if (ingredients_order == 0) ingredients_order = order_index
          next
        }
        if (heading ~ /^Process($| \(|:)/) {
          section = "process"
          process_seen = 1
          order_index++
          if (process_order == 0) process_order = order_index
          next
        }
        if (heading ~ /^Formula($| \(|:)/) {
          section = "formula"
          formula_seen = 1
          next
        }
        section = "other"
        next
      }
      /^# / {
        section = "other"
        next
      }
      /^---$/ {
        section = "other"
        next
      }
      /^[[:space:]]*$/ { next }
      section == "yield" && ($0 ~ /^- / || $0 ~ /^\* /) {
        yield_bullets++
        next
      }
      section == "ingredients" && ($0 ~ /^- / || $0 ~ /^\* /) {
        ingredients_bullets++
        line = $0
        sub(/^[-*] /, "", line)
        if (line !~ /:/) ingredients_unstructured++
        next
      }
      section == "process" && $0 ~ /^[0-9]+\. / {
        process_steps++
        next
      }
      section == "formula" && ($0 ~ /^- / || $0 ~ /^\* /) {
        formula_bullets++
        next
      }
      END {
        printf "yield_seen=%d\n", yield_seen
        printf "yield_bullets=%d\n", yield_bullets
        printf "ingredients_seen=%d\n", ingredients_seen
        printf "ingredients_bullets=%d\n", ingredients_bullets
        printf "ingredients_unstructured=%d\n", ingredients_unstructured
        printf "process_seen=%d\n", process_seen
        printf "process_steps=%d\n", process_steps
        printf "formula_seen=%d\n", formula_seen
        printf "formula_bullets=%d\n", formula_bullets
        printf "yield_order=%d\n", yield_order
        printf "ingredients_order=%d\n", ingredients_order
        printf "process_order=%d\n", process_order
      }
    ' "$file"
  )

  eval "$analysis"

  if [ "${yield_seen:-0}" -eq 1 ] && [ "${yield_bullets:-0}" -eq 0 ]; then
    fail "$file: yield block must include at least one bullet line"
  fi

  if [ "${ingredients_bullets:-0}" -eq 0 ]; then
    fail "$file: ingredients section must include at least one bullet item"
  fi

  if [ "${process_steps:-0}" -eq 0 ]; then
    fail "$file: process section must include at least one numbered step"
  fi

  if [ "${yield_order:-0}" -gt 0 ] && [ "${ingredients_order:-0}" -gt 0 ] && [ "$yield_order" -gt "$ingredients_order" ]; then
    fail "$file: yield block should appear before ingredients"
  fi

  if [ "${ingredients_order:-0}" -gt 0 ] && [ "${process_order:-0}" -gt 0 ] && [ "$ingredients_order" -gt "$process_order" ]; then
    fail "$file: ingredients section should appear before process"
  fi

  if [ "${formula_seen:-0}" -eq 1 ] && [ "${formula_bullets:-0}" -eq 0 ]; then
    warn "$file: formula section is present but has no bullet items"
  fi

  if [ "${ingredients_unstructured:-0}" -gt 0 ]; then
    warn "$file: ${ingredients_unstructured} ingredient item(s) do not use \"name: amount\" format"
  fi
done

duplicates=$(
  awk -F '\t' '
    { count[$1]++; files[$1] = files[$1] ? files[$1] ", " $2 : $2 }
    END {
      for (title in count) {
        if (count[title] > 1) {
          printf "%s\t%s\n", title, files[title]
        }
      }
    }
  ' "$titles_tmp"
)

if [ -n "$duplicates" ]; then
  status=1
  while IFS="$(printf '\t')" read -r title files; do
    [ -n "$title" ] || continue
    printf 'FAIL: duplicate recipe title "%s": %s\n' "$title" "$files" >&2
  done <<EOF
$duplicates
EOF
fi

if [ "$status" -ne 0 ]; then
  exit "$status"
fi

printf 'Validated %s recipe files\n' "$(find recipes -maxdepth 1 -type f -name '*.md' | wc -l | tr -d ' ')"
