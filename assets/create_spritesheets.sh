#!/bin/zsh

set -e

BASE_DIR="art-archive"
INPUT_DIR="art-archive/art-raw"
OUTPUT_DIR="art-archive/spritesheets/player"

mkdir -p "$OUTPUT_DIR"

if ! command -v convert &> /dev/null
then
    echo "ImageMagick is not installed. Please install it to run this script." >&2
    echo "On macOS with Homebrew, you can run: brew install imagemagick" >&2
    echo "On Debian/Ubuntu, you can run: sudo apt-get install imagemagick" >&2
    exit 1
fi

echo "Processing animations from $INPUT_DIR..."

process_directory() {
    local source_dir="$1"
    local dir_name="$2"
    
    local png_files=("$source_dir"/*.png)
    if [ ! -e "${png_files[1]}" ]; then
        echo "    No PNG files found in $source_dir to process, skipping."
        return
    fi
    
    local output_name=$(echo "$dir_name" | tr '[:upper:]' '[:lower:]')
    local output_file="$OUTPUT_DIR/$output_name.png"

    echo "  - Creating $output_file from $dir_name"
    convert "$source_dir"/*.png +append "$output_file"
}

for dir in "$INPUT_DIR"/*/; do
    local dir_name=$(basename "$dir")

    if [ "$dir_name" = "striking" ]; then
        echo "Processing 'striking' animations..."
        for sub_dir in "$dir"/*/; do
            local sub_dir_name=$(basename "$sub_dir")
            process_directory "$sub_dir" "$sub_dir_name"
        done
    else
        process_directory "$dir" "$dir_name"
    fi
done

echo "Successfully created spritesheets in $OUTPUT_DIR" 