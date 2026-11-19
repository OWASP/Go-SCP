#!/usr/bin/env bash
# build-books.sh builds the PDF, ePub and DOCX editions of the book with
# pandoc, running in Docker so the only local requirement is Docker.
#
# The chapters and their order come from src/SUMMARY.md, the same file
# GitBook uses, so there is no second list to keep in sync.
#
# Environment:
#   OUT_DIR       where to write the books, inside the repository, since only
#                 the repository is mounted into the container (default: dist)
#   FORMATS       which editions to build (default: "pdf epub docx")
#   PANDOC_IMAGE  image to run; if unset, Dockerfile.pandoc is built and used
set -euo pipefail

cd "$(dirname "$0")"

out_dir=${OUT_DIR:-dist}
formats=${FORMATS:-pdf epub docx}
name=go-webapp-scp
version=$(tr -d '[:space:]' < VERSION)

image=${PANDOC_IMAGE:-}
if [[ -z $image ]]; then
    image=go-scp-pandoc
    docker build --quiet --file Dockerfile.pandoc --tag "$image" . > /dev/null
fi

# Every Markdown file linked from SUMMARY.md, in order, e.g.
# "* [Validation](input-validation/validation.md)" -> src/input-validation/validation.md
chapters=()
while IFS= read -r path; do
    chapters+=("src/$path")
done < <(sed -nE 's/^ *\* \[[^]]*\]\(([^)]+\.md)\).*/\1/p' src/SUMMARY.md)
if (( ${#chapters[@]} == 0 )); then
    echo "build-books.sh: no chapters found in src/SUMMARY.md" >&2
    exit 1
fi

mkdir -p "$out_dir"
case $(cd "$out_dir" && pwd -P)/ in
"$(pwd -P)"/*) ;;
*)
    echo "build-books.sh: OUT_DIR must be inside the repository: $out_dir" >&2
    exit 1
    ;;
esac

# run_pandoc runs pandoc in the image with the options shared by all formats,
# then the extra options given, then the chapters.
run_pandoc() {
    docker run --rm \
        --user "$(id -u):$(id -g)" \
        --volume "$PWD:/data" \
        "$image" \
        --from=gfm+rebase_relative_paths \
        --file-scope \
        --toc --toc-depth=1 \
        --metadata-file=book-metadata.yaml \
        --metadata=subtitle:"Version $version" \
        --lua-filter=fix-internal-links.lua \
        "$@" \
        "${chapters[@]}"
}

for format in $formats; do
    output=$out_dir/$name.$format
    echo "building $output"
    case $format in
    pdf)
        run_pandoc --pdf-engine=xelatex --include-in-header=book-cover.tex \
            --lua-filter=wrap-wide-tables.lua --output="$output"
        ;;
    epub | docx)
        run_pandoc --output="$output"
        ;;
    *)
        echo "build-books.sh: unknown format $format (want pdf, epub or docx)" >&2
        exit 1
        ;;
    esac
done
