#!/usr/bin/env python3
"""Render scripted Jin TUI cell captures as a CPU-only GIF."""

import argparse
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("frames", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--font", type=Path)
    args = parser.parse_args()
    candidates = [
        args.font,
        Path("/System/Library/Fonts/Menlo.ttc"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
    ]
    font_path = next((path for path in candidates if path and path.is_file()), None)
    if font_path is None:
        parser.error("provide a monospace font with --font")
    font = ImageFont.truetype(str(font_path), 18)
    cell_width, cell_height = round(font.getlength("M")), 24
    captures = json.loads(args.frames.read_text())
    images, delays = [], []
    for frame in captures:
        width, height = frame["width"], frame["height"]
        image = Image.new("RGB", (width * cell_width + 32, height * cell_height + 88), "#080a0f")
        draw = ImageDraw.Draw(image)
        draw.text((16, 10), "Jin v0.8.0 | Scripted TUI demo", font=font, fill="#9da8bd")
        for index, cell in enumerate(frame["cells"]):
            x, y = index % width * cell_width + 16, index // width * cell_height + 44
            draw.rectangle((x, y, x + cell_width - 1, y + cell_height - 1), fill=tuple(cell["bg"]))
        for index, cell in enumerate(frame["cells"]):
            if cell["text"].strip():
                x, y = index % width * cell_width + 16, index // width * cell_height + 42
                draw.text((x, y), cell["text"], font=font, fill=tuple(cell["fg"]))
        draw.text((16, height * cell_height + 53), frame["label"], font=font, fill="#c7d4ec")
        images.append(image)
        delays.append(frame["delay"])
    if not images:
        parser.error("the cell capture contains no frames")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    images[0].save(args.output, save_all=True, append_images=images[1:], duration=delays,
                   loop=0, optimize=True, disposal=2)
    print(f"Rendered {len(images)} frames to {args.output} ({args.output.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
