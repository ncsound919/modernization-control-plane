# Module 05 — Project Importers

## Purpose
Provides specialized parsers and migration converters for professionals stranded by the discontinuation of Adobe Animate (.FLA) and Autodesk EAGLE (.BRD/.SCH). Creates a direct migration path without forcing adoption of large CAD/creative suites.

## Supported File Formats

| Format | Source Tool | Discontinued | Output |
|---|---|---|---|
| `.FLA` | Adobe Animate | March 1, 2026 | Lottie JSON, SVG, HTML5 Canvas |
| `.BRD` | Autodesk EAGLE (PCB layout) | June 7, 2026 | KiCad `.kicad_pcb`, neutral schema |
| `.SCH` | Autodesk EAGLE (schematic) | June 7, 2026 | KiCad `.kicad_sch`, neutral schema |

## FLA Importer (.FLA → Modern Formats)
- Parse FLA XML structure (timeline, layers, symbols, scripts)
- Extract embedded assets (bitmaps, sounds, vector shapes)
- Convert vector animations to Lottie JSON for web
- Convert timeline animations to HTML5 Canvas / CSS animations
- Preserve ActionScript annotations for manual migration guidance

## EAGLE Importer (.BRD/.SCH → KiCad)
- Parse EAGLE XML board and schematic files
- Map EAGLE design rules to KiCad DRC equivalents
- Convert component libraries, footprints, and net lists
- Output KiCad 7.x compatible `.kicad_pcb` and `.kicad_sch` files
- Generate migration report: unmapped components, manual review items

## Tech Stack
- **Language:** Python 3.11+
- **FLA parsing:** XML/ZIP parsing, custom FLA schema decoder
- **Animation output:** `lottie-python`, `svgwrite`
- **EAGLE parsing:** EAGLE XML schema, `lxml`
- **KiCad output:** `pcbnew` Python API, KiCad file format v7
- **API:** FastAPI (REST import service)

## Directory Structure
```
05-project-importers/
├── fla/
│   ├── parser/           # FLA file parser
│   ├── extractors/       # Asset and timeline extractors
│   └── converters/       # Lottie, SVG, Canvas output
├── eagle/
│   ├── parser/           # BRD/SCH file parser
│   └── converters/       # KiCad output converters
├── api/                  # FastAPI import service
├── tests/
│   ├── fixtures/         # Sample .FLA, .BRD, .SCH test files
│   └── test_importers.py
└── README.md
```

## Key Metrics
- Conversion fidelity score (% of elements successfully mapped)
- Number of files successfully imported per format
- Migration report items requiring manual review (lower = better)

## Status
`Phase 3 — Planned`
